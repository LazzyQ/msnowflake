package client

import (
	"encoding/json"
	"errors"
	myrpc "github.com/LazzyQ/msnowflake/rpc"
	"github.com/samuel/go-zookeeper/zk"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/rpc"
	"path"
	"sort"
	"strconv"
	"sync"
	"time"
)

const (
	zkNodeDelaySleep    = 1 * time.Second // zk error delay sleep
	rpcClientPingSleep  = 1 * time.Second // rpc client ping need sleep
	rpcClientRetrySleep = 1 * time.Second // rpc client retry connect need sleep

	RPCPing    = "SnowflakeRPC.Ping"
	RPCNextId  = "SnowflakeRPC.NextId"
	RPCNextIds = "SnowflakeRPC.NextIds"
)

var (
	ErrNoRpcClient = errors.New("rpc: no rpc client service")
	mutex          sync.Mutex
	zkPath         string
	zkConn         *zk.Conn
	zkServers      []string
	zkTimeout      time.Duration
	// worker
	workerIdMap = map[int64]*Client{}
)

func Init(zServers []string, zPath string, zTimeout time.Duration) (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	if zkConn != nil {
		return
	}
	zkPath = zPath
	zkServers = zServers
	zkTimeout = zTimeout
	conn, session, err := zk.Connect(zkServers, zkTimeout)
	if err != nil {
		log.WithFields(log.Fields{
			"zkServers": zkServers,
			"zkTimeout": zkTimeout,
			"error":     err,
		}).Error("连接zk失败")
	}
	zkConn = conn
	go func() {
		for {
			event := <-session
			log.WithField("event", event.Type.String()).Info("收到zk event")
		}
	}()
	return
}

// msnowflake的客户端
type Client struct {
	workerId int64
	clients  []*rpc.Client
	stop     chan bool
	leader   string
}

// Peer store data in zookeeper.
type Peer struct {
	RPC    []string `json:"rpc"`
	Thrift []string `json:"thrift"`
}

func NewClient(workerId int64) (c *Client) {
	var ok bool
	mutex.Lock()
	defer mutex.Unlock()
	if c, ok = workerIdMap[workerId]; ok {
		return
	}
	c = &Client{
		workerId: workerId,
		clients:  nil,
		leader:   "",
	}
	go c.watchWorkerId(strconv.FormatInt(workerId, 10))
	workerIdMap[workerId] = c
	return
}

func (c *Client) Id() (id int64, err error) {
	client, err := c.client()
	if err != nil {
		return
	}
	if err = client.Call(RPCNextId, c.workerId, &id); err != nil {
		log.WithFields(log.Fields{
			"call":     RPCNextId,
			"workerId": c.workerId,
			"error":    err,
		}).Error("rpc.Call 调用失败")
	}
	return
}

func (c *Client) Ids(num int) (ids []int64, err error) {
	client, err := c.client()
	if err != nil {
		return
	}
	if err = client.Call(RPCNextIds, &myrpc.NextIdsArgs{WorkerId: c.workerId, Num: num}, &ids); err != nil {
		log.WithFields(log.Fields{
			"call":     RPCNextIds,
			"workerId": c.workerId,
			"error":    err,
		}).Error("rpc.Call 调用失败")
	}
	return
}

func (c *Client) watchWorkerId(workerIdStr string) {
	workerIdPath := path.Join(zkPath, workerIdStr)
	log.Debugf("workerIdPath: %s", workerIdPath)
	for {
		rpcs, _, watch, err := zkConn.ChildrenW(workerIdPath)
		if err != nil {
			log.WithFields(log.Fields{
				"workerIdPath": workerIdPath,
				"error": err,
			}).Error("zkConn.ChildrenW失败")
			time.Sleep(zkNodeDelaySleep)
			continue
		}
		if len(rpcs) == 0 {
			log.WithField("workerIdPath", workerIdPath).Error("zkConn.ChildrenW没有节点")
			time.Sleep(zkNodeDelaySleep)
			continue
		}
		sort.Strings(rpcs)
		newLeader := rpcs[0]
		if c.leader == newLeader {
			log.WithField("workerId", workerIdStr).Info("workerId添加了一个备用节点", workerIdStr)
		} else {

			log.WithFields(log.Fields{
				"workerId": workerIdStr,
				"oldLeader": c.leader,
				"newLeader": newLeader,
			}).Info("leader发生变化，重新选举")
			// get new leader info
			workerNodePath := path.Join(zkPath, workerIdStr, newLeader)
			bs, _, err := zkConn.Get(workerNodePath)
			if err != nil {
				log.WithFields(log.Fields{
					"workerNodePath": workerNodePath,
					"error": err,
				}).Error("zk Get失败", workerNodePath, err)
				time.Sleep(zkNodeDelaySleep)
				continue
			}
			peer := &Peer{}
			if err = json.Unmarshal(bs, peer); err != nil {
				log.WithFields(log.Fields{
					"bytes内容": string(bs),
					"error": err,
				}).Error("json反序列失败")
				time.Sleep(zkNodeDelaySleep)
				continue
			}
			// 初始化RPC
			tmpClients := make([]*rpc.Client, len(peer.RPC))
			tmpStop := make(chan bool, 1)
			for i, addr := range peer.RPC {
				clt, err := rpc.Dial("tcp", addr)
				if err != nil {
					log.WithFields(log.Fields{
						"addr": addr,
						"error": err,
					}).Error("rpc.Dial失败", addr, err)
					continue
				}
				tmpClients[i] = clt
				go c.pingAndRetry(tmpStop, clt, addr)
			}
			oldClients := c.clients
			oldStop := c.stop

			c.leader = newLeader
			c.clients = tmpClients
			c.stop = tmpStop

			if oldClients != nil {
				closeRpc(oldClients, oldStop)
			}
		}
		event := <-watch
		log.WithFields(log.Fields{
			"workerIdPath": workerIdPath,
			"event": event.Type.String(),
		}).Error("zk node发生改变")
	}
}

func (c *Client) pingAndRetry(stop <-chan bool, client *rpc.Client, addr string) {
	defer func() {
		if err := client.Close(); err != nil {
			log.WithField("error", err).Error("client.Close()失败")
		}
	}()

	var (
		failed bool
		status int
		err    error
		tmp    *rpc.Client
	)

	for {
		select {
		case <-stop:
			log.WithField("addr", addr).Info("pingAndRetry goroutine 退出", addr)
			return
		default:
		}
		if !failed {
			if err = client.Call(RPCPing, 0, &status); err != nil {
				log.WithFields(log.Fields{
					"rpc": RPCPing,
					"error": err,
				}).Error("client.Call失败")
				failed = true
				continue
			} else {
				failed = false
				time.Sleep(rpcClientPingSleep)
				continue
			}
		}

		if tmp, err = rpc.Dial("tcp", addr); err != nil {
			log.WithFields(log.Fields{
				"addr": addr,
				"error": err,
			}).Error("rpc.Dial失败")
			time.Sleep(rpcClientRetrySleep)
			continue
		}
		client = tmp
		failed = false

		log.WithField("addr", addr).Info("client reconnect 成功", addr)
	}
}

func closeRpc(clients []*rpc.Client, stop chan bool) {
	for _, client := range clients {
		if client != nil {
			if err := client.Close(); err != nil {
				log.WithField("error", err).Error("client.Close()失败", err)
			}
		}
	}

	if stop != nil {
		close(stop)
	}
}

func (c *Client) Close() {
	closeRpc(c.clients, c.stop)
	mutex.Lock()
	defer mutex.Unlock()
	delete(workerIdMap, c.workerId)
}

func (c *Client) client() (*rpc.Client, error) {
	clientNum := len(c.clients)
	if clientNum == 0 {
		return nil, ErrNoRpcClient
	} else if clientNum == 1 {
		return c.clients[0], nil
	} else {
		return c.clients[rand.Intn(clientNum)], nil
	}
}

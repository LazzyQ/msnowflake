package client

import (
	"encoding/json"
	"github.com/samuel/go-zookeeper/zk"
	log "github.com/sirupsen/logrus"
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
	mutex sync.Mutex
	zkPath    string
	zkConn    *zk.Conn
	zkServers []string
	zkTimeout time.Duration
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
			"error": err,
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
	clients []*rpc.Client
	stop chan bool
	leader string
}

// Peer store data in zookeeper.
type Peer struct {
	RPC    []string `json:"rpc"`
	Thrift []string `json:"thrift"`
}


func NewClient(workerId int64) (c *Client)  {
	var ok bool
	mutex.Lock()
	defer mutex.Unlock()
	if c, ok = workerIdMap[workerId]; ok {
		return
	}
	c = &Client{
		workerId: workerId,
		clients: nil,
		leader: "",
	}
	go c.watchWorkerId(workerId, strconv.FormatInt(workerId, 10))
	workerIdMap[workerId] = c
	return
}


func (c *Client) watchWorkerId(workerId int64, workerIdStr string) {
	workerIdPath := path.Join(zkPath, workerIdStr)
	log.Debugf("workerIdPath: %s", workerIdPath)
	for {
		rpcs, _, watch, err := zkConn.ChildrenW(workerIdStr)
		if err != nil {
			log.Errorf("zkConn.ChildrenW(%s) error(%v)", workerIdPath, err)
			time.Sleep(zkNodeDelaySleep)
			continue
		}
		if len(rpcs) == 0 {
			log.Errorf("zkConn.ChildrenW(%s) no nodes", workerIdPath)
			time.Sleep(zkNodeDelaySleep)
			continue
		}
		sort.Strings(rpcs)
		newLeader := rpcs[0]
		if c.leader == newLeader {
			log.Info("workerId: %s add a new standby msnowflake node", workerIdStr)
		} else {
			log.Info("workerId: %s oldLeader: \"%s\", newLeader: \"%s\" not equals, continue leader selection", workerIdStr, c.leader, newLeader)
			// get new leader info
			workerNodePath := path.Join(zkPath, workerIdStr, newLeader)
			bs, _, err := zkConn.Get(workerNodePath)
			if err != nil {
				log.Error("zkConn.Get(%s) error(%v)", workerNodePath, err)
				time.Sleep(zkNodeDelaySleep)
				continue
			}
			peer := &Peer{}
			if err = json.Unmarshal(bs, peer); err != nil {
				log.Error("json.Unmarshal(%s, peer) error(%v)", string(bs), err)
				time.Sleep(zkNodeDelaySleep)
				continue
			}
			// 初始化RPC
			tmpClients := make([]*rpc.Client, len(peer.RPC))
			tmpStop := make(chan bool, 1)
			for i, addr := range peer.RPC {
				clt, err := rpc.Dial("tcp", addr)
				if err != nil {
					log.Error("rpc.Dial(tcp, \"%s\") error(%v)", addr, err)
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
		log.Error("zk node(\"%s\") changed %s", workerIdPath, event.Type.String())
	}
}

func (c *Client) pingAndRetry(stop <-chan bool, client *rpc.Client, addr string) {
	defer func() {
		if err := client.Close(); err != nil {
			log.Error("client.Close() error(%v)", err)
		}
	}()

	var (
		failed bool
		status int
		err error
		tmp *rpc.Client
	)

	for {
		select {
		case <-stop:
			log.Info("addr: \"%s\" pingAndRetry goroutine exit", addr)
			return
		default:
		}
		if !failed {
			if err = client.Call(RPCPing, 0, &status); err != nil {
				log.Error("client.Call(%s) error(%v)", RPCPing, err)
				failed = true
				continue
			} else {
				failed = false
				time.Sleep(rpcClientPingSleep)
				continue
			}
		}

		if tmp, err = rpc.Dial("tcp", addr); err != nil {
			log.Error("rpc.Dial(tcp, %s) error(%v)", addr, err)
			time.Sleep(rpcClientRetrySleep)
			continue
		}
		client = tmp
		failed = false
		log.Info("client reconnect %s ok", addr)
	}
}

func closeRpc(clients []*rpc.Client, stop chan bool) {
	for _, client := range clients {
		if client != nil {
			if err := client.Close(); err != nil {
				log.Error("client.Close() error(%v)", err)
			}
		}
	}

	if stop != nil {
		close(stop)
	}
}
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	log "github.com/sirupsen/logrus"
	"net/rpc"
	"strconv"
	"time"
)

const (
	timestampMaxDelay = int64(10 * time.Second)
)

type Peer struct {
	RPC    []string `json:"rpc"`
	Thrift []string `json:"thrift"`
}

var (
	zkConn *zk.Conn
)

func InitZK() (err error) {
	conn, session, err := zk.Connect(MyConf.Zookeeper.ZKAddr, MyConf.Zookeeper.ZKTimeout)

	if err != nil {
		log.WithFields(log.Fields{
			"addr":    MyConf.Zookeeper.ZKAddr,
			"timeout": MyConf.Zookeeper.ZKTimeout,
			"error":   err,
		}).Error("zk.Connect error")
	}

	zkConn = conn
	go func() {
		for {
			event := <-session
			log.WithField("event", event.State.String()).Info("zk get a event")
		}
	}()
	return nil
}

func CloseZK() {
	zkConn.Close()
}

func SanityCheckPeers() error {
	peers, err := getPeers()
	if err != nil {
		return err
	}

	if len(peers) == 0 {
		return nil
	}

	timestamps := int64(0)
	timestamp := int64(0)
	dataCenterId := int64(0)
	peerCount := int64(0)

	for id, workers := range peers {
		for _, peer := range workers {
			if len(peer.RPC) > 0 {
				// 取worker的第一个是因为第一个是leader，其他的都是standBy的
				cli, err := rpc.Dial("tcp", peer.RPC[0])
				if err != nil {
					log.Error("rpc.Dial(\"tcp\", \"%s\") error(%v)", peer.RPC[0], err)
					return err
				}
				defer cli.Close()
				if err = cli.Call("SnowflakeRPC.DatacenterId", 0, &dataCenterId); err != nil {
					log.Error("rpc.Call(\"SnowflakeRPC.DatacenterId\", 0) error(%v)", err)
					return err
				}
				if err = cli.Call("SnowflakeRPC.Timestamp", 0, &timestamp); err != nil {
					log.Error("rpc.Call(\"SnowflakeRPC.Timestamp\", 0) error(%v)", err)
					return err
				}
			} else if len(peer.Thrift) > 0 {
				// TODO thrift call
			} else {
				log.Error("workerId: %d don't have any rpc address", id)
				return errors.New("workerId no rpc")
			}
			if dataCenterId != MyConf.Snowflake.DataCenterId {
				log.Error("workerId: %d has datacenterId %d, but ours is %d", id, dataCenterId, MyConf.Snowflake.DataCenterId)
				return errors.New("Datacenter id insanity")
			}
			// add timestamps
			timestamps += timestamp
			peerCount++
		}
	}
	// check 10s
	// calc avg timestamps
	now := time.Now().Unix()
	avg := int64(timestamps / peerCount)
	log.Debug("timestamps: %d, peer: %d, avg: %d, now - avg: %d, maxdelay: %d", timestamps, peerCount, avg, now-avg, timestampMaxDelay)
	if now-avg > timestampMaxDelay {
		log.Error("timestamp sanity check failed. Mean timestamp is %d, but mine is %d so I'm more than 10s away from the mean", avg, now)
		return errors.New("timestamp sanity check failed")
	}
	return nil
}

// /zk-path
//   /worker-id
//     /ephemeral|sequence
//     /leader
//     /standby
func getPeers() (map[int][]*Peer, error) {
	if _, err := zkConn.Create(MyConf.Zookeeper.ZKPath, []byte(""), 0, zk.WorldACL(zk.PermAll)); err != nil {
		if err == zk.ErrNodeExists {
			log.WithField("path", MyConf.Zookeeper.ZKPath).Warn("zk node exists")
		} else {
			log.WithFields(log.Fields{
				"path":  MyConf.Zookeeper.ZKPath,
				"error": err,
			}).Error("创建节点失败")
			return nil, err
		}
	}

	workers, _, err := zkConn.Children(MyConf.Zookeeper.ZKPath)
	if err != nil {
		log.WithFields(log.Fields{
			"path":  MyConf.Zookeeper.ZKPath,
			"error": err,
		}).Error("获取子节点失败")
	}
	res := make(map[int][]*Peer, len(workers))
	for _, worker := range workers {
		id, err := strconv.Atoi(worker)
		if err != nil {
			log.Error("strconv.Atoi(\"%s\") error(%v)", worker, err)
			return nil, err
		}
		workerIdPath := fmt.Sprintf("%s/%s", MyConf.Zookeeper.ZKPath, worker)
		nodes, _, err := zkConn.Children(workerIdPath)
		for _, node := range nodes {
			nodePath := fmt.Sprintf("%s/%s", workerIdPath, node)
			d, _, err := zkConn.Get(nodePath)
			if err != nil {
				log.Error("zk.Get(\"%s\") error(%v)", nodePath, err)
				return nil, err
			}
			peer := &Peer{}
			if err = json.Unmarshal(d, peer); err != nil {
				log.Error("json.Unmarshal(\"%s\", peer) error(%v)", d, err)
				return nil, err
			}
			peers, ok := res[id]
			if !ok {
				peers = []*Peer{peer}
			} else {
				peers = append(peers, peer)
			}
			res[id] = peers
		}
	}
	return res, nil
}

func RegWorkerId(workerId int64) (err error) {
	workerIdPath := fmt.Sprintf("%s/%d", MyConf.Zookeeper.ZKPath, workerId)
	if _, err = zkConn.Create(workerIdPath, []byte(""), 0, zk.WorldACL(zk.PermAll)); err != nil {
		if err == zk.ErrNodeExists {
			log.WithField("workerIdPath", workerIdPath).Warn("zk创建的节点已经存在")
		} else {
			log.WithFields(log.Fields{
				"workerIdPath": workerIdPath,
				"err":          err,
			}).Error("zk创建的节点已经存在失败", workerIdPath, err)
			return
		}
	}

	d, err := json.Marshal(&Peer{RPC: MyConf.Base.RPCBind, Thrift: MyConf.Base.ThriftBind})
	if err != nil {
		log.WithField("err", err).Error("节点数据Peer序列化失败")
		return
	}
	workerIdPath += "/"
	if _, err = zkConn.Create(workerIdPath, d, zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll)); err != nil {
		log.Error("zk.create(\"%s\") error(%v)", workerIdPath, err)
		return
	}
	return
}

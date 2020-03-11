package zk

import (
	"github.com/LazzyQ/msnowflake/basic/config"
	log "github.com/micro/go-micro/v2/logger"
	myzk "github.com/samuel/go-zookeeper/zk"
	"strconv"
	"strings"
)

var (
	zkConn            *myzk.Conn
	snowflakeRootPath string
)

func Init() {
	zkConfig := config.GetZookeeperConfig()
	snowflakeRootPath = zkConfig.GetPath()
	conn, session, err := myzk.Connect(zkConfig.GetAddr(), zkConfig.GetTimeout())

	if err != nil {
		log.Errorf("连接zk失败, addr:%v, timeout:%v, err:%v", zkConfig.GetAddr(), zkConfig.GetTimeout(), err)
		panic(err)
	}

	zkConn = conn

	go func() {
		for {
			event := <-session
			log.Infof("收到zk事件:%v", event.State.String())
		}
	}()


	if _, err := zkConn.Create(snowflakeRootPath, []byte(""), 0, myzk.WorldACL(myzk.PermAll)); err != nil {
		if err != myzk.ErrNodeExists {
			log.Errorf("zk创建snowflake根节点失败, path:%v, err:%v", snowflakeRootPath, err)
			panic(err)
		}
	}
}

func CreateSnowflakeWorkerNode(workerId int64) {
	path := strings.Join([]string{snowflakeRootPath, strconv.FormatInt(workerId, 10)}, "/")
	if _, err := zkConn.Create(path, []byte(""), myzk.FlagEphemeral, myzk.WorldACL(myzk.PermAll)); err != nil {
		log.Errorf("zk创建snowflake节点失败 path:%v, err:%v", path, err)
		panic(err)
	}
}

func CloseZK() {
	zkConn.Close()
}

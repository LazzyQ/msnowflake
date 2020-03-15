package zk

import (
	"github.com/LazzyQ/msnowflake/basic/config"
	myzk "github.com/samuel/go-zookeeper/zk"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

var (
	zkConn            *myzk.Conn
	snowflakeRootPath string
)

func Init() error {
	zkConfig := config.GetZookeeperConfig()
	snowflakeRootPath = zkConfig.GetPath()
	conn, session, err := myzk.Connect(zkConfig.GetAddr(), zkConfig.GetTimeout())

	if err != nil {
		log.WithFields(log.Fields{
			"addr": zkConfig.GetAddr(),
			"timeout": zkConfig.GetTimeout(),
			"err": err,
		}).Error("连接zk失败" )
		return err
	}

	zkConn = conn

	go func() {
		for {
			event := <-session
			log.WithField("事件", event.State.String()).Info("收到zk事件")
		}
	}()

	if _, err := zkConn.Create(snowflakeRootPath, []byte(""), 0, myzk.WorldACL(myzk.PermAll)); err != nil {
		if err != myzk.ErrNodeExists {
			log.WithFields(log.Fields{
				"根路径": snowflakeRootPath,
				"err": err,
			}).Error("zk创建snowflake根节点失败")
			return err
		}
	}
	return nil
}

func CreateSnowflakeWorkerNode(workerId int64) error {
	path := strings.Join([]string{snowflakeRootPath, strconv.FormatInt(workerId, 10)}, "/")
	if _, err := zkConn.Create(path, []byte(""), myzk.FlagEphemeral, myzk.WorldACL(myzk.PermAll)); err != nil {
		log.WithFields(log.Fields{
			"路径": path,
			"err": err,
		}).Errorf("zk创建snowflake节点失败")
		return err
	}
	return nil
}

func CloseZK() {
	zkConn.Close()
}

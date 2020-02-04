package main

import (
	"github.com/samuel/go-zookeeper/zk"
	log "github.com/sirupsen/logrus"
)

var (
	zkConn *zk.Conn
)

func InitZK() (err error) {
	conn, session, err := zk.Connect(MyConf.Zookeeper.ZKAddr, MyConf.Zookeeper.ZKTimeout)

	if err != nil {
		log.WithFields(log.Fields{
			"addr": MyConf.Zookeeper.ZKAddr,
			"timeout": MyConf.Zookeeper.ZKTimeout,
			"error": err,
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
package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"runtime"
)

func main() {

	log.WithField("maxWorkerId", maxWorkerId).Info("======")
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(MyConf.Base.MaxProc)

	log.WithFields(log.Fields{
		"data-center": MyConf.Snowflake.DataCenterId,
	}).Info("msnowflake 服务启动")

	if err := InitProcess(); err != nil {
		panic(err)
	}
	InitPprof()
	if err := InitZK(); err != nil {
		panic(err)
	}
	defer CloseZK()

	if err := SanityCheckPeers(); err != nil {
		panic(err)
	}

	workers, err := NewWorkers()
	if err != nil {
		panic(err)
	}

	if err := InitRPC(workers); err != nil {
		panic(err)
	}

}

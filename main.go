package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"runtime"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(MyConf.Base.MaxProc)

	log.Info("xx")

}

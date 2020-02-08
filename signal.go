package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func InitSignal() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	return c
}

func HandleSignal(c chan os.Signal) {
	for {
		s := <-c
		log.WithField("signal", s.String()).Info("gosnowflake收到信息")
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			return
		case syscall.SIGHUP:
			// todo reload
			// return
		}
	}
}

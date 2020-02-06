package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/pprof"
)

func InitPprof() {
	pprofServeMux := http.NewServeMux()
	pprofServeMux.HandleFunc("/debug/pprof", pprof.Index)
	pprofServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	for _, addr := range MyConf.Base.PprofBind {
		go func() {
			if err := http.ListenAndServe(addr, pprofServeMux); err != nil {
				log.WithFields(log.Fields{
					"addr":  addr,
					"error": err,
				}).Error("http.ListenAndServe error")
			}
		}()
	}
}

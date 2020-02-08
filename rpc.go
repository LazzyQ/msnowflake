package main

import (
	"errors"
	myrpc "github.com/LazzyQ/msnowflake/rpc"
	log "github.com/sirupsen/logrus"
	"net"
	"net/rpc"
)

type SnowflakeRPC struct {
	workers Workers
}

func InitRPC(workers Workers) error {
	s := &SnowflakeRPC{workers: workers}
	rpc.Register(s)
	for _, bind := range MyConf.Base.RPCBind {
		log.Info("start listen rpc addr: \"%s\"", bind)
		go rpcListen(bind)
	}
	return nil
}

func rpcListen(bind string) {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		log.Error("net.Listen(\"tcp\", \"%s\") error(%v)", bind, err)
		panic(err)
	}
	defer func() {
		log.Info("rpc addr: \"%s\" close", bind)
		if err := l.Close(); err != nil {
			log.Error("listener.Close() error(%v)", err)
		}
	}()
	rpc.Accept(l)
}

func (s *SnowflakeRPC) NextId(workerId int64, id *int64) error {
	worker, err := s.workers.Get(workerId)
	if err != nil {
		return err
	}
	if tid, err := worker.NextId(); err != nil {
		log.WithField("error", err).Error("worker.NextId()失败")
		return err
	} else {
		*id = tid
		return nil
	}
}

func (s *SnowflakeRPC) NextIds(args *myrpc.NextIdsArgs, ids *[]int64) error {
	if args == nil {
		return errors.New("参数不能为空")
	}
	worker, err := s.workers.Get(args.WorkerId)
	if err != nil {
		return err
	}
	if tids, err := worker.NextIds(args.Num); err != nil {
		log.Error("worker.NextIds(%d) error(%v)", args.Num, err)
		return err
	} else {
		*ids = tids
		return nil
	}
}

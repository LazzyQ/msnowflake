package main

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type Workers []*IdWorker

func NewWorkers() (Workers, error) {
	idWorkers := make([]*IdWorker, maxWorkerId)
	for _, workerId := range MyConf.Snowflake.WorkerId {
		if t := idWorkers[workerId]; t != nil {
			log.WithField("workerId", workerId).Error("初始化，workerId已经存在", workerId)
			return nil, fmt.Errorf("初始化 workerId: %d 已经存在", workerId)
		}
		idWorker, err := NewIdWorker(workerId, MyConf.Snowflake.DataCenterId, MyConf.Twepoch)
		if err != nil {
			log.WithFields(log.Fields{
				"dataCenterId": MyConf.Snowflake.DataCenterId,
				"workerId":     workerId,
				"err":          err,
			}).Error("初始化IdWorker失败")
			return nil, err
		}
		idWorkers[workerId] = idWorker
		if err := RegWorkerId(workerId); err != nil {
			log.WithFields(log.Fields{
				"workerId": workerId,
				"err":      err,
			}).Error("注册Worker失败")
			return nil, err
		}
	}
	return Workers(idWorkers), nil
}

func (w Workers) Get(workerId int64) (*IdWorker, error) {
	if workerId > maxWorkerId || workerId < 0 {
		log.Error("worker Id can't be greater than %d or less than 0", maxWorkerId)
		return nil, errors.New(fmt.Sprintf("worker Id: %d error", workerId))
	}
	if worker := w[workerId]; worker == nil {
		log.WithField("workerId", workerId).Warn("workerId没有注册", workerId)
		return nil, fmt.Errorf("msnowflake workerId: %d 没有在服务上进行注册", workerId)
	} else {
		return worker, nil
	}
}

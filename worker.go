package main

import (
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

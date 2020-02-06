package main

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
)

// id = [timestamp][dataCenterId:5][workerId:5][sequence:12]
const (
	twepoch            = int64(1288834974657)
	workerIdBits       = uint(5)
	dataCenterIdBits   = uint(5)
	maxWorkerId        = -1 ^ (-1 << workerIdBits)
	maxDataCenterId    = -1 ^ (-1 << dataCenterIdBits)
	sequenceBits       = uint(12)
	workerIdShift      = sequenceBits
	dataCenterIdShift  = sequenceBits + workerIdBits
	timestampLeftShift = sequenceBits + workerIdBits + dataCenterIdBits
	sequenceMask       = -1 ^ (-1 << sequenceBits)
	maxNextIdsNum      = 100
)

type IdWorker struct {
	sequence      int64
	lastTimestamp int64
	workerId      int64
	twepoch       int64
	dataCenterId  int64
	mutex         sync.Mutex
}

func NewIdWorker(workerId, dataCenterId int64, twepoch int64) (*IdWorker, error) {
	idWorker := &IdWorker{}
	if workerId > maxWorkerId || workerId < 0 {
		log.WithFields(log.Fields{
			"upper": maxWorkerId,
			"lower": 0,
		}).Error("workerId必须在区间内", maxWorkerId)
		return nil, errors.New(fmt.Sprintf("worker Id: %d error", workerId))
	}
	if dataCenterId > maxDataCenterId || dataCenterId < 0 {
		log.WithFields(log.Fields{
			"upper": maxDataCenterId,
			"lower": 0,
		}).Error("dataCenterId必须在区间内", maxDataCenterId)
		return nil, errors.New(fmt.Sprintf("datacenter Id: %d error", dataCenterId))
	}

	idWorker.workerId = workerId
	idWorker.dataCenterId = dataCenterId
	idWorker.lastTimestamp = -1
	idWorker.sequence = 0
	idWorker.twepoch = twepoch
	idWorker.mutex = sync.Mutex{}
	log.WithFields(log.Fields{
		"timestamp左移":    timestampLeftShift,
		"dataCenterId位数": dataCenterIdBits,
		"workerId位数":     workerIdBits,
		"sequence位数":     sequenceBits,
		"workerId":       workerId,
	}).Debug("worker启动")
	return idWorker, nil
}

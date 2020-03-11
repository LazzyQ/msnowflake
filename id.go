package main

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
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
	twepoch       int64 // 起始时间
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

func (id *IdWorker) NextId() (int64, error) {
	id.mutex.Lock()
	defer id.mutex.Unlock()
	timestamp := timeGen()
	if timestamp < id.lastTimestamp {
		log.WithFields(log.Fields{
			"timestamp":     timestamp,
			"lastTimestamp": id.lastTimestamp,
		}).Error("时钟回调，请求拒绝")
		return 0, errors.New(fmt.Sprintf("时钟回调. 请求拒绝%dms", id.lastTimestamp-timestamp))
	}
	if id.lastTimestamp == timestamp {
		id.sequence = (id.sequence + 1) & sequenceMask
		if id.sequence == 0 {
			timestamp = tilNextMillis(id.lastTimestamp)
		}
	} else {
		id.sequence = 0
	}
	id.lastTimestamp = timestamp
	return ((timestamp - id.twepoch) << timestampLeftShift) | (id.dataCenterId << dataCenterIdShift) | (id.workerId << workerIdShift) | id.sequence, nil
}

func (id *IdWorker) NextIds(num int) ([]int64, error) {
	if num > maxNextIdsNum || num < 0 {
		log.WithFields(log.Fields{
			"maxIdNum":     maxNextIdsNum,
			"currentIdNum": num,
		}).Error("获取id超过NextIds限制的数量或小于0")
		return nil, errors.New(fmt.Sprintf("NextIds数量参数不对: %d", num))
	}
	ids := make([]int64, num)
	id.mutex.Lock()
	defer id.mutex.Unlock()
	for i := 0; i < num; i++ {
		timestamp := timeGen()
		if timestamp < id.lastTimestamp {
			log.WithFields(log.Fields{
				"timestamp":     timestamp,
				"lastTimestamp": id.lastTimestamp,
			}).Error("时钟回调，请求拒绝")
			return nil, errors.New(fmt.Sprintf("时钟回调. 请求拒绝%dms", id.lastTimestamp-timestamp))
		}
		if id.lastTimestamp == timestamp {
			id.sequence = (id.sequence + 1) & sequenceMask
			if id.sequence == 0 {
				timestamp = tilNextMillis(id.lastTimestamp)
			}
		} else {
			id.sequence = 0
		}
		id.lastTimestamp = timestamp
		ids[i] = ((timestamp - id.twepoch) << timestampLeftShift) | (id.dataCenterId << dataCenterIdShift) | (id.workerId << workerIdShift) | id.sequence
	}
	return ids, nil
}

// 返回的是当前时间戳，但是是ms
func timeGen() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

func tilNextMillis(lastTimestamp int64) int64 {
	timestamp := timeGen()
	for timestamp <= lastTimestamp {
		timestamp = timeGen()
	}
	return timestamp
}

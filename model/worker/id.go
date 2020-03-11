package worker

import (
	"errors"
	"fmt"
	log "github.com/micro/go-micro/v2/logger"
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
	return ((timestamp - id.twepoch) << timestampLeftShift) | (id.dataCenterId << dataCenterIdBits) | (id.workerId << workerIdShift) | id.sequence, nil
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

package worker

import (
	"errors"
	"github.com/LazzyQ/msnowflake/basic/config"
	"github.com/LazzyQ/msnowflake/model/zk"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	worker *IdWorker
)

func Init() (err error) {
	worker, err = newIdWorker()
	if err != nil {
		return
	}
	return nil
}

type IdWorker struct {
	sequence      int64
	lastTimestamp int64
	workerId      int64
	twepoch       int64 // 起始时间
	dataCenterId  int64
	mutex         sync.Mutex
}

func newIdWorker() (*IdWorker, error) {
	snowflakeConfig := config.GetSnowflakeConfig()
	dataCenterId := snowflakeConfig.GetDataCenter()
	workerId := snowflakeConfig.GetWorkerId()
	twepoch := snowflakeConfig.GetTwepoch()

	idWorker := &IdWorker{}
	if workerId > maxWorkerId || workerId < 0 {
		log.WithFields(log.Fields{
			"upper": maxWorkerId,
			"lower": 0,
		}).Error("workerId必须在区间内")
		return nil, errors.New("workerId超过限制")
	}
	if dataCenterId > maxDataCenterId || dataCenterId < 0 {
		log.WithFields(log.Fields{
			"upper": maxDataCenterId,
			"lower": 0,
		}).Error("dataCenterId必须在区间内")
		return nil, errors.New("dataCenterId超过限制")
	}

	zk.CreateSnowflakeWorkerNode(workerId)

	idWorker.workerId = workerId
	idWorker.dataCenterId = dataCenterId
	idWorker.lastTimestamp = -1
	idWorker.sequence = 0
	idWorker.twepoch = twepoch.UnixNano() / int64(time.Millisecond)
	idWorker.mutex = sync.Mutex{}
	log.WithFields(log.Fields{
		"timestamp左移":    timestampLeftShift,
		"dataCenterId位数": dataCenterIdBits,
		"workerId位数":     workerIdBits,
		"sequence位数":     sequenceBits,
		"workerId":       workerId,
	}).Debug("worker启动完成...")
	return idWorker, nil
}

func GetIdWorker() (*IdWorker, error) {
	if worker == nil {
		return nil, errors.New("worker未完成初始化")
	}
	return worker, nil
}

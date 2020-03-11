package worker

import (
	"github.com/LazzyQ/msnowflake/basic/config"
	log "github.com/sirupsen/logrus"
	"sync"
)

var (
	worker *IdWorker
)

func Init()  {
	worker = NewIdWorker()
}

type IdWorker struct {
	sequence      int64
	lastTimestamp int64
	workerId      int64
	twepoch       int64 // 起始时间
	dataCenterId  int64
	mutex         sync.Mutex
}

func NewIdWorker() *IdWorker {

	zkConfig := config.GetZookeeperConfig()
	snowflakeConfig := config.GetSnowflakeConfig()
	dataCenterId := snowflakeConfig.GetDataCenter()
	workerId := snowflakeConfig.GetWorkerId()
	twepoch := snowflakeConfig.GetTwepoch()

	idWorker := &IdWorker{}
	if workerId > maxWorkerId || workerId < 0 {
		log.Fatalf("workerId必须在区间内, upper:%d, lower:%d", maxWorkerId, 0)
	}
	if dataCenterId > maxDataCenterId || dataCenterId < 0 {
		log.Fatalf("dataCenterId必须在区间内upper:%d, lower:%d", maxDataCenterId, 0)
	}

	idWorker.workerId = workerId
	idWorker.dataCenterId = dataCenterId
	idWorker.lastTimestamp = -1
	idWorker.sequence = 0
	idWorker.twepoch = twepoch
	idWorker.mutex = sync.Mutex{}
	log.Debugf("worker启动, timestamp左移:%v, dataCenterId位数:%v,workerId位数:%v, sequence位数:%v, workerId:%v",
		timestampLeftShift, dataCenterIdBits, workerIdBits, sequenceBits, workerId)
	return idWorker
}

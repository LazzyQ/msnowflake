package worker

import (
	"errors"
	"github.com/LazzyQ/msnowflake/basic/config"
	"github.com/LazzyQ/msnowflake/model/zk"
	log "github.com/micro/go-micro/v2/logger"
	"sync"
	"time"
)

var (
	worker *IdWorker
)

func Init()  {
	worker = newIdWorker()
}

type IdWorker struct {
	sequence      int64
	lastTimestamp int64
	workerId      int64
	twepoch       int64 // 起始时间
	dataCenterId  int64
	mutex         sync.Mutex
}

func newIdWorker() *IdWorker {
	snowflakeConfig := config.GetSnowflakeConfig()
	dataCenterId := snowflakeConfig.GetDataCenter()
	workerId := snowflakeConfig.GetWorkerId()
	twepoch := snowflakeConfig.GetTwepoch()

	idWorker := &IdWorker{}
	if workerId > maxWorkerId || workerId < 0 {
		log.Errorf("workerId必须在区间内, upper:%d, lower:%d", maxWorkerId, 0)
		panic(errors.New("workerId超过限制"))
	}
	if dataCenterId > maxDataCenterId || dataCenterId < 0 {
		log.Errorf("dataCenterId必须在区间内upper:%d, lower:%d", maxDataCenterId, 0)
		panic(errors.New("dataCenterId超过限制"))
	}

	zk.CreateSnowflakeWorkerNode(workerId)

	idWorker.workerId = workerId
	idWorker.dataCenterId = dataCenterId
	idWorker.lastTimestamp = -1
	idWorker.sequence = 0
	idWorker.twepoch = twepoch.UnixNano() / int64(time.Microsecond)
	idWorker.mutex = sync.Mutex{}
	log.Debugf("worker启动, timestamp左移:%v, dataCenterId位数:%v,workerId位数:%v, sequence位数:%v, workerId:%v",
		timestampLeftShift, dataCenterIdBits, workerIdBits, sequenceBits, workerId)
	return idWorker
}

func GetIdWorker() *IdWorker{
	return worker
}

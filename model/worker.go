package model

import (
	"errors"
	"github.com/LazzyQ/msnowflake/basic"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	worker *IdWorker
)

type IdWorker struct {
	sequence      int64
	lastTimestamp int64
	workerId      int64
	twepoch       int64 // 起始时间
	dataCenterId  int64
	mutex         sync.Mutex
}

func InitIdWorker(config basic.SnowflakeConfig) (*IdWorker, error) {
	dataCenterId := config.GetDataCenter()
	workerId := config.GetWorkerId()
	twepoch, err := config.GetTwepoch()
	if err != nil {
		return nil, err
	}

	idWorker := &IdWorker{}
	if workerId > maxWorkerId || workerId < 0 {
		zap.S().Errorw("workerId必须在区间内", "upper", maxWorkerId, "lower", 0)
		return nil, errors.New("workerId超过限制")
	}
	if dataCenterId > maxDataCenterId || dataCenterId < 0 {
		zap.S().Errorw("dataCenterId超过限制", "upper", maxDataCenterId, "lower", 0)
		return nil, errors.New("dataCenterId超过限制")
	}

	etcd := basic.GetEtcd()
	workerKey := strings.Join([]string{"msnowflake", "worker", strconv.FormatInt(workerId, 10)}, "/")
	txResponse, err := etcd.TxKeepaliveWithTTL(workerKey, string(workerId), 2)
	if err != nil {
		return nil, err
	}

	if !txResponse.Success {
		zap.S().Errorw("worker注册到etcd失败")
		return nil, errors.New("worker注册失败")
	}

	idWorker.workerId = workerId
	idWorker.dataCenterId = dataCenterId
	idWorker.lastTimestamp = -1
	idWorker.sequence = 0
	idWorker.twepoch = twepoch.UnixNano() / int64(time.Millisecond)
	idWorker.mutex = sync.Mutex{}

	zap.S().Infow("worker启动完成...",
		"timestamp左移", timestampLeftShift,
		"dataCenterId位数", dataCenterIdBits,
		"workerId位数", workerIdBits,
		"sequence位数", sequenceBits,
		"workerId", workerId)
	worker = idWorker
	return idWorker, nil
}

func GetIdWorker() (*IdWorker, error) {
	if worker == nil {
		return nil, errors.New("worker未完成初始化")
	}
	return worker, nil
}

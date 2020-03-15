package config

import (
	//log "github.com/sirupsen/logrus"
	"time"
)

type SnowflakeConfig struct {
	Port       int64
	WorkerId   int64
	DataCenter int64
	Twepoch    time.Time
}

func (p SnowflakeConfig) GetPort() int64 {
	return p.Port
}

func (p SnowflakeConfig) GetDataCenter() int64 {
	return p.DataCenter
}

func (p SnowflakeConfig) GetTwepoch() time.Time {
	//twepoch, err := time.Parse("2006-01-02 15:04:05", p.Twepoch)
	//if err != nil {
	//	log.Errorf("Snowflake的Twepoch配置不正确 twepoch:%v, err:%v", p.Twepoch, err)
	//	panic(err)
	//}
	return p.Twepoch
}

func (p SnowflakeConfig) GetWorkerId() int64 {
	return p.WorkerId
}

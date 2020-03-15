package config

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type SnowflakeConfig struct {
	Port       int64
	WorkerId   int64
	DataCenter int64
	Twepoch    string
}

func (p SnowflakeConfig) GetPort() int64 {
	return p.Port
}

func (p SnowflakeConfig) GetDataCenter() int64 {
	return p.DataCenter
}

func (p SnowflakeConfig) GetTwepoch() (time.Time, error) {
	twepoch, err := time.Parse("2006-01-02 15:04:05", p.Twepoch)
	if err != nil {
		log.WithFields(log.Fields{
			"twepoch": p.Twepoch,
			"err": err,
		}).Error("Snowflake的Twepoch配置不正确")
		return twepoch, err
	}
	return twepoch, nil
}

func (p SnowflakeConfig) GetWorkerId() int64 {
	return p.WorkerId
}

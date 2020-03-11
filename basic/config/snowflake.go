package config

import (
	log "github.com/micro/go-micro/v2/logger"
	"time"
)

type SnowflakeConfig interface {
	GetPort() int64
	GetWorkerId() int64
	GetDataCenter() int64
	GetTwepoch() time.Time
}

type defaultSnowflakeConfig struct {
	Port       int64     `json:"port"`
	WorkerId   int64     `json:"workerId"`
	DataCenter int64     `json:"dataCenter"`
	Twepoch    string `json:"twepoch"`
}

func (p defaultSnowflakeConfig) GetPort() int64 {
	return p.Port
}

func (p defaultSnowflakeConfig) GetDataCenter() int64 {
	return p.DataCenter
}

func (p defaultSnowflakeConfig) GetTwepoch() time.Time {
	 twepoch, err :=  time.Parse("2006-01-02 15:04:05", p.Twepoch)
	 if err != nil {
	 	log.Errorf("Snowflake的Twepoch配置不正确 twepoch:%v, err:%v", p.Twepoch, err)
	 	panic(err)
	 }
	 return twepoch
}

func (p defaultSnowflakeConfig) GetWorkerId() int64 {
	return p.WorkerId
}

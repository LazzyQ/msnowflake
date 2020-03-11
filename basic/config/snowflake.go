package config

import "time"

type SnowflakeConfig interface {
	GetPort() int64
	GetWorkerId() int64
	GetDataCenter() int64
	GetTwepoch() time.Time
}

type defaultSnowflakeConfig struct {
	Port       int64     `json:"port"`
	WorkerId   int64     `json:"worker_id"`
	DataCenter int64     `json:"data_center"`
	Twepoch    time.Time `json:"twepoch"`
}

func (p defaultSnowflakeConfig) GetPort() int64 {
	return p.Port
}

func (p defaultSnowflakeConfig) GetDataCenter() int64 {
	return p.DataCenter
}

func (p defaultSnowflakeConfig) GetTwepoch() time.Time {
	return p.Twepoch
}

func (p defaultSnowflakeConfig) GetWorkerId() int64 {
	return p.WorkerId
}

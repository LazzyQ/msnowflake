package config

import "time"

type ZookeeperConfig interface {
	GetAddr() []string
	GetTimeout() time.Duration
	GetPath() string
}

type defaultZookeeperConfig struct {
	Addr    []string      `json:"addr"`
	Timeout time.Duration `json:"timeout"`
	Path    string        `json:"path"`
}

func (p defaultZookeeperConfig) GetAddr() []string {
	return p.Addr
}

func (p defaultZookeeperConfig) GetTimeout() time.Duration {
	return p.Timeout
}


func (p defaultZookeeperConfig) GetPath() string {
	return p.Path
}
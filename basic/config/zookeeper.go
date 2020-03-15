package config

import "time"

type ZookeeperConfig struct {
	Addr    []string
	Timeout time.Duration
	Path    string
}

func (p ZookeeperConfig) GetAddr() []string {
	return p.Addr
}

func (p ZookeeperConfig) GetTimeout() time.Duration {
	return p.Timeout
}


func (p ZookeeperConfig) GetPath() string {
	return p.Path
}
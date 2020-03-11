package config

type ZookeeperConfig interface {
	GetAddr() []string
}

type defaultZookeeperConfig struct {
	Addr []string `json:"addr"`
}

func (p defaultZookeeperConfig) GetAddr() []string {
	return p.Addr
}

package config

// EtcdConfig Etcd 配置
type EtcdConfig interface {
	GetPort() int
	GetHost() string
}


// defaultEtcdConfig 默认Etcd 配置
type defaultEtcdConfig struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

// GetPort Etcd 端口
func (c defaultEtcdConfig) GetPort() int {
	return c.Port
}



// GetHost Etcd 主机地址
func (c defaultEtcdConfig) GetHost() string {
	return c.Host
}

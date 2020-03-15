package config

// EtcdConfig 默认Etcd 配置
type EtcdConfig struct {
	Host    string
	Port    int
}

// GetPort Etcd 端口
func (c EtcdConfig) GetPort() int {
	return c.Port
}


// GetHost Etcd 主机地址
func (c EtcdConfig) GetHost() string {
	return c.Host
}

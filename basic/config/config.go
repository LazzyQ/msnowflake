package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	c Config
)

type Config struct {
	Etcd      EtcdConfig
	Snowflake SnowflakeConfig
	Zookeeper ZookeeperConfig
	Log       LogConfig
}

func Init() error {
	viper.SetConfigFile("application.yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&c); err != nil {
		return err
	}
	log.WithField("内容", c).Info("配置文件解析完成")
	return nil
}

// GetEtcdConfig 获取Etcd配置
func GetEtcdConfig() EtcdConfig {
	return c.Etcd
}

func GetZookeeperConfig() ZookeeperConfig {
	return c.Zookeeper
}

func GetSnowflakeConfig() SnowflakeConfig {
	return c.Snowflake
}

func GetLogConfig() LogConfig {
	return c.Log
}

package config

import (
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source"
	"github.com/micro/go-micro/v2/config/source/file"
	log "github.com/micro/go-micro/v2/logger"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	defaultRootPath         = "app"
	defaultConfigFilePrefix = "application-"
	profiles                defaultProfiles
	etcdConfig              defaultEtcdConfig
	snowflakeConfig         defaultSnowflakeConfig
	zookeeperConfig         defaultZookeeperConfig
	inited                  bool
	m                       sync.RWMutex
	sp                      = string(filepath.Separator)
	err                     error
)

func Init() {
	m.Lock()
	defer m.Unlock()
	if inited {
		log.Info("[Init] 配置已经初始化过")
		return
	}

	appPath, _ := filepath.Abs(filepath.Dir(filepath.Join("."+sp, sp)))
	pt := filepath.Join(appPath, "conf")
	os.Chdir(appPath)
	// 找到application.yml文件
	if err = config.Load(file.NewSource(file.WithPath(pt + sp + "application.yml"))); err != nil {
		panic(err)
	}

	if err = config.Get(defaultRootPath, "profiles").Scan(&profiles); err != nil {
		panic(err)
	}

	log.Infof("[Init] 加载配置文件：path: %s, %+v\n", pt+sp+"application.yml", profiles)

	// 开始导入新文件
	if len(profiles.GetInclude()) > 0 {
		include := strings.Split(profiles.GetInclude(), ",")
		sources := make([]source.Source, len(include))
		for i := 0; i < len(include); i++ {
			filePath := pt + string(filepath.Separator) + defaultConfigFilePrefix + strings.TrimSpace(include[i]) + ".yml"
			log.Infof("[Init] 加载配置文件：path: %s\n", filePath)
			sources[i] = file.NewSource(file.WithPath(filePath))
		}
		// 加载include的文件
		if err = config.Load(sources...); err != nil {
			panic(err)
		}
	}

	// 赋值
	config.Get(defaultRootPath, "etcd").Scan(&etcdConfig)
	config.Get(defaultRootPath, "snowflake").Scan(&snowflakeConfig)
	config.Get(defaultRootPath, "zookeeper").Scan(&zookeeperConfig)

	// 标记已经初始化
	inited = true
}

// GetEtcdConfig 获取Etcd配置
func GetEtcdConfig() EtcdConfig {
	return etcdConfig
}

func GetZookeeperConfig() ZookeeperConfig {
	return zookeeperConfig
}

func GetSnowflakeConfig() SnowflakeConfig {
	return snowflakeConfig
}

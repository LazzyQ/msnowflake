package main

import (
	"fmt"
	"github.com/LazzyQ/msnowflake/basic"
	"github.com/LazzyQ/msnowflake/basic/config"
	"github.com/LazzyQ/msnowflake/handler"
	"github.com/LazzyQ/msnowflake/model"
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"log"
)

func main() {
	// 初始化基础配置
	basic.Init()
	micReg := etcd.NewRegistry(registryOption)

	srv := micro.NewService(
		micro.Name("go.micro.srv.snowflake"),
		micro.Registry(micReg),
	)

	srv.Init(
		micro.Action(func(context *cli.Context) error {
			model.Init()
			handler.Init()
			return nil
		}),
	)

	if err := srv.Run(); err != nil {
		log.Fatal("msnowflake服务启动失败")
	}
}

func registryOption(ops *registry.Options) {
	etcdCfg := config.GetEtcdConfig()
	ops.Addrs = []string{fmt.Sprintf("%s:%d", etcdCfg.GetHost(), etcdCfg.GetPort())}
}

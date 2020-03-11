package main

import (
	"fmt"
	"github.com/LazzyQ/msnowflake/basic"
	"github.com/LazzyQ/msnowflake/basic/config"
	"github.com/LazzyQ/msnowflake/handler"
	"github.com/LazzyQ/msnowflake/model"
	"github.com/LazzyQ/msnowflake/proto"
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2"
	log "github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
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

	_ = msnowflake.RegisterMSnowflakeHandler(srv.Server(), new(handler.MSnowflake))

	if err := srv.Run(); err != nil {
		log.Error("msnowflake服务启动失败")
		panic(err)
	}
}

func registryOption(ops *registry.Options) {
	etcdCfg := config.GetEtcdConfig()
	ops.Addrs = []string{fmt.Sprintf("%s:%d", etcdCfg.GetHost(), etcdCfg.GetPort())}
}

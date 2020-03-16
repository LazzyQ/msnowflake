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
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	log "github.com/sirupsen/logrus"
)

func main() {
	// 初始化基础配置
	if err := basic.Init(); err != nil {
		log.WithField("err", err).Fatal("basic初始化失败")
	}
	micReg := etcd.NewRegistry(registryOption)

	srv := micro.NewService(
		micro.Name("go.micro.srv.snowflake"),
		micro.Registry(micReg),
	)

	srv.Init(
		micro.Action(func(context *cli.Context) error{
			if err := model.Init(); err != nil {
				log.WithField("err", err).Error("model初始化失败")
				return err
			}
			if err := handler.Init(); err != nil {
				log.WithField("err", err).Error("handler初始化失败")
				return err
			}
			return nil
		}),
	)

	if err := msnowflake.RegisterMSnowflakeHandler(srv.Server(), new(handler.MSnowflake)); err != nil {
		log.WithField("err", err).Error("注册处理器失败")
		return
	}

	if err := srv.Run(); err != nil {
		log.WithField("err", err).Error("msnowflake服务启动失败")
		return
	}
}

func registryOption(ops *registry.Options) {
	etcdCfg := config.GetEtcdConfig()
	ops.Addrs = []string{fmt.Sprintf("%s:%d", etcdCfg.GetHost(), etcdCfg.GetPort())}
}

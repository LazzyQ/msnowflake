package main

import (
	"github.com/LazzyQ/msnowflake/basic"
	"github.com/LazzyQ/msnowflake/handler"
	"github.com/LazzyQ/msnowflake/model"
	"github.com/LazzyQ/msnowflake/proto"
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"go.uber.org/zap"
	"strings"
	"time"
)

func main() {
	logConfig := basic.LogConfig{}
	etcdConfig := basic.EtcdConfig{}
	snowflakeConfig := basic.SnowflakeConfig{}

	srv := micro.NewService(
		micro.Name("go.micro.srv.snowflake"),
		micro.Flags(
			&cli.StringFlag{
				Name:        "log_filename",
				Usage:       "设置日志文件名",
				Value:       "/var/wemeng/msnowflake/msnowflake.log",
				Destination: &logConfig.Filename,
			},
			&cli.StringFlag{
				Name:        "log_level",
				Usage:       "设置日志级别",
				Value:       "info",
				Destination: &logConfig.Level,
			},
			&cli.IntFlag{
				Name:        "log_max_size",
				Usage:       "设置日志大小(MB)",
				Value:       200,
				Destination: &logConfig.MaxSize,
			},
			&cli.IntFlag{
				Name:        "log_max_age",
				Usage:       "设置日志文件保存时间(day)",
				Value:       30,
				Destination: &logConfig.MaxAge,
			},
			&cli.StringFlag{
				Name:  "etcd_address",
				Usage: "etcd集群地址",
				Value: "127.0.0.1:2379",
			},
			&cli.IntFlag{
				Name:  "etcd_connection_timeout",
				Usage: "etcd集群连接超时时间(s)",
				Value: 5,
			},
			&cli.IntFlag{
				Name:  "etcd_read_timeout",
				Usage: "etcd集群超时时间(s)",
				Value: 2,
			},
			&cli.Int64Flag{
				Name:        "msnowflake_worker_id",
				Usage:       "workerId",
				Value:       1,
				Destination: &snowflakeConfig.WorkerId,
			},
			&cli.Int64Flag{
				Name:        "msnowflake_datacenter",
				Usage:       "workerId",
				Value:       1,
				Destination: &snowflakeConfig.DataCenter,
			},
			&cli.StringFlag{
				Name:        "msnowflake_twepoch",
				Usage:       "twepoch",
				Value:       "2020-02-02 13:14:52",
				Destination: &snowflakeConfig.Twepoch,
			},
		),
		micro.Action(func(c *cli.Context) error {
			etcdAddrs := c.String("etcd_address")
			endpoints := strings.Split(etcdAddrs, ",")
			etcdConfig.Endpoints = endpoints
			etcdConfig.ReadTimeout = time.Duration(c.Int("etcd_read_timeout")) * time.Second
			etcdConfig.ConnectTimeout = time.Duration(c.Int("etcd_connection_timeout")) * time.Second
			return nil
		}),
	)

	srv.Init(
		micro.BeforeStart(func() (err error) {
			basic.InitLog(logConfig)
			if err = basic.InitEtcd(etcdConfig); err != nil {
				return
			}
			if _, err = model.InitIdWorker(snowflakeConfig); err != nil {
				return
			}

			if err = handler.Init(); err != nil {
				return
			}
			return nil
		}),
		micro.AfterStop(func() (err error) {
			err = zap.L().Sync()
			basic.GetEtcd().Close()
			return err
		}),
	)

	if err := msnowflake.RegisterMSnowflakeHandler(srv.Server(), new(handler.MSnowflake)); err != nil {
		zap.S().Errorw("注册处理器失败", "err", err)
		return
	}

	if err := srv.Run(); err != nil {
		zap.S().Errorw("msnowflake服务启动失败", "err", err)
		return
	}
}

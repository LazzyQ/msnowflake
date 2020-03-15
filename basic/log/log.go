package log

import (
	"github.com/LazzyQ/msnowflake/basic/config"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"time"
)

func Init() error {
	logConfig := config.GetLogConfig()
	logHook, err :=  newLfsHook(logConfig)
	if err != nil {
		return err
	}
	log.AddHook(logHook)
	return nil
}

func newLfsHook(logConfig config.LogConfig) (log.Hook, error) {
	writer, err := rotatelogs.New(
		logConfig.Filename + logConfig.LogSuffix,
		// WithLinkName为最新的日志建立软连接,以方便随着找到当前日志文件
		rotatelogs.WithLinkName(logConfig.Filename),

		// WithRotationTime设置日志分割的时间,这里设置为一小时分割一次
		rotatelogs.WithRotationTime(time.Hour * 24),

		// WithMaxAge和WithRotationCount二者只能设置一个,
		// WithMaxAge设置文件清理前的最长保存时间,
		// WithRotationCount设置文件清理前最多保存的个数.
		//rotatelogs.WithMaxAge(time.Hour*24),
		rotatelogs.WithRotationCount(logConfig.MaxRemainCount),
	)

	if err != nil {
		log.WithField("err", err).Error("本地日志文件配置失败")
		return nil, err
	}

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: writer,
		log.InfoLevel:  writer,
		log.WarnLevel:  writer,
		log.ErrorLevel: writer,
		log.FatalLevel: writer,
		log.PanicLevel: writer,
	}, &log.TextFormatter{DisableColors: false})

	return lfsHook, nil
}

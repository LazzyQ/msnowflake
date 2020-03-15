package log

import (
	"github.com/LazzyQ/msnowflake/basic/config"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)


func Init() error {
	logConfig := config.GetLogConfig()
	logFilename := strings.Join([]string{logConfig.GetPath(), logConfig.GetFilename()}, "/")
	f, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	return nil
}

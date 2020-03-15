package basic

import (
	"github.com/LazzyQ/msnowflake/basic/config"
	"github.com/LazzyQ/msnowflake/basic/log"
)

func Init() (err error) {
	if err = config.Init(); err != nil {
		return
	}
	if err = log.Init(); err != nil {
		return
	}
	return nil
}

package model

import (
	"github.com/LazzyQ/msnowflake/model/worker"
	"github.com/LazzyQ/msnowflake/model/zk"
)

func Init() (err error) {
	if err = zk.Init(); err != nil {
		return err
	}

	if err = worker.Init(); err != nil {
		return err
	}
	return nil
}
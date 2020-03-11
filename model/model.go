package model

import (
	"github.com/LazzyQ/msnowflake/model/worker"
	"github.com/LazzyQ/msnowflake/model/zk"
)

func Init()  {
	zk.Init()
	worker.Init()
}
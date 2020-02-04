package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

// InitProcess: 创建pid文件，设置工作路径，setgid and setuid
func InitProcess() (err error) {
	if err = os.Chdir(MyConf.Base.Dir); err != nil {
		return
	}

	// 创建pid文件
	if err = ioutil.WriteFile(MyConf.Base.PidFile, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644); err != nil {
		return
	}

	return
}
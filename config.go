package main

import (
	"io/ioutil"
	"os"
	"runtime"
	"gopkg.in/yaml.v2"
	"time"
)

var (
	MyConf   *Config
	confPath string
)

type Config struct {
	Base *Base
	Snowflake *Snowflake
	Zookeeper *Zookeeper
	Twepoch int64
}

type Base struct {
	PidFile    string   `yaml:"pid"`
	Dir        string   `yaml:"dir"`
	Log        string   `yaml:"log"`
	MaxProc    int      `yaml:"maxproc"`
	PRCBind    []string `yaml:"rpc-bind"`
	ThriftBind []string `yaml:"thrift-bind"`
	StatBind   []string `yaml:"stat-bind"`
	PprofBind  []string `yaml:"pprof-bind"`
}

type Snowflake struct {
	DataCenterId int64 `yaml:"data-center"`
	WorkerId []int64 `yaml:"worker"`
	Start string `yaml:"start"`

}

type Zookeeper struct {
	ZKAddr []string `yaml:"addr"`
	ZKTimeout time.Duration `yaml:"timeout"`
	ZKPath string `yaml:"path"`
}

func InitConfig() (err error) {

	var (
		file *os.File
		blob []byte
		twepoch time.Time
	)
	c = new(Config)

	if file, err = os.Open(confPath); err != nil {
		return
	}
	if blob, err = ioutil.ReadAll(file); err != nil {
		return
	}

	MyConf = &Config{
		Base: &Base{
			PidFile: "/tmp/gosnowflake.pid",
			Dir: "/dev/null",
			Log: "./log/xml",
			MaxProc: runtime.NumCPU(),
			PRCBind: []string{"localhost:8080"},
			ThriftBind: []string{"localhost:8081"},
		},
		Snowflake: &Snowflake{
			DataCenterId: 0,
			WorkerId: []int64{0},
			Start: "2020-02-02 13:14:52",
		},
		Zookeeper: &Zookeeper{
			ZKAddr: []string{"localhost:2181"},
			ZKTimeout: time.Second * 15,
			ZKPath: "/msnowflake-servers",
		},
	}

	if err = yaml.Unmarshal(blob, MyConf); err == nil {
		if twepoch, err = time.Parse("2006-01-02 15:04:05", MyConf.Snowflake.Start); err != nil {
			return
		} else {
			MyConf.Twepoch = twepoch.UnixNano() / int64(time.Microsecond)
		}
	}
	return
}

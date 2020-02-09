package client

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

var (
	confPath string
	MyConf   *Config
)

type Config struct {
	Base      *Base
	Zookeeper *Zookeeper
}

type Base struct {
	RPCAddr  string `yaml:"rpc-addr"`
	WorkerId int64  `yaml:"workerId"`
}

type Zookeeper struct {
	ZKServers []string      `yaml:"addr"`
	ZKPath    string        `yaml:"path"`
	ZKTimeout time.Duration `yaml:"timeout"`
}

func init() {
	flag.StringVar(&confPath, "conf", "./test.yaml", "msnowflake配置文件路径")
}

func InitConf() (err error) {
	var (
		file *os.File
		blob []byte
	)

	if file, err = os.Open(confPath); err != nil {
		return
	}

	if blob, err = ioutil.ReadAll(file); err != nil {
		return
	}

	MyConf = &Config{
		Base: &Base{
			RPCAddr:  "localhost:8080",
			WorkerId: 0,
		},
		Zookeeper: &Zookeeper{
			ZKServers: []string{"localhost:2181"},
			ZKPath:    "/msnowflake-servers",
			ZKTimeout: time.Second * 5,
		},
	}

	if err = yaml.Unmarshal(blob, MyConf); err != nil {
		return
	}
	return nil
}

func Test(t *testing.T) {
	if err := InitConf(); err != nil {
		t.Error(err)
	}

	if err := Init(MyConf.Zookeeper.ZKServers, MyConf.Zookeeper.ZKPath, MyConf.Zookeeper.ZKTimeout); err != nil {
		t.Error(err)
	}

	c := NewClient(MyConf.Base.WorkerId)
	for i := 0; i < 60; i++ {
		time.Sleep(1 * time.Second)
		id, err := c.Id()
		if err != nil {
			t.Error(err)
		}
		ids, err := c.Ids(5)
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("gosnwoflake id: %d\n", id)
		fmt.Printf("gosnwoflake ids: %d\n", ids)
	}
	c.Close()
	if _, ok := workerIdMap[MyConf.Base.WorkerId]; ok {
		t.Error("workerId exists")
	}
}

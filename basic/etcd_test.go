package basic

import (
	"testing"
	"time"
)

func TestEtcd_Get(t *testing.T) {
	config := EtcdConfig{
		Endpoints:      []string{"192.168.0.200:2379", "192.168.0.200:2369", "192.168.0.200:2389"},
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    2 * time.Second,
	}

	if err := InitEtcd(config); err != nil {
		t.Error("初始化etcd失败", err)
	}

	val, err := etcd.Get("root/foo")
	if err != nil {
		t.Error("获取值失败", err)
	}

	t.Log("val: ", string(val))
}

func TestEtcd_GetWithPrefixKey(t *testing.T) {
	config := EtcdConfig{
		Endpoints:      []string{"192.168.0.200:2379", "192.168.0.200:2369", "192.168.0.200:2389"},
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    2 * time.Second,
	}

	if err := InitEtcd(config); err != nil {
		t.Error("初始化etcd失败", err)
	}

	keys, values, err := etcd.GetWithPrefixKey("root")
	if err != nil {
		t.Error("获取值失败", err)
	}

	for i := 0; i < len(keys); i++ {
		t.Log("key: ", string(keys[i]), ", value: ", string(values[i]))
	}
}

func TestEtcd_GetWithPrefixKeyLimit(t *testing.T) {
	config := EtcdConfig{
		Endpoints:      []string{"192.168.0.200:2379", "192.168.0.200:2369", "192.168.0.200:2389"},
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    2 * time.Second,
	}

	if err := InitEtcd(config); err != nil {
		t.Error("初始化etcd失败", err)
	}

	keys, values, err := etcd.GetWithPrefixKeyLimit("root", 2)
	if err != nil {
		t.Error("获取值失败", err)
	}

	for i := 0; i < len(keys); i++ {
		t.Log("key: ", string(keys[i]), ", value: ", string(values[i]))
	}
}

func TestEtcd_Put(t *testing.T) {
	config := EtcdConfig{
		Endpoints:      []string{"192.168.0.200:2379", "192.168.0.200:2369", "192.168.0.200:2389"},
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    2 * time.Second,
	}

	if err := InitEtcd(config); err != nil {
		t.Error("初始化etcd失败", err)
	}

	if err := etcd.Put("root/oof", "oof"); err != nil {
		t.Error(err)
	}
}

func TestEtcd_PutNotExist(t *testing.T) {
	config := EtcdConfig{
		Endpoints:      []string{"192.168.0.200:2379", "192.168.0.200:2369", "192.168.0.200:2389"},
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    2 * time.Second,
	}

	if err := InitEtcd(config); err != nil {
		t.Error("初始化etcd失败", err)
	}

	success, oldValue, err := etcd.PutNotExist("root/oof", "oof2")
	if err != nil {
		t.Error(err)
	}

	t.Log("success: ", success, " oldValue: ", string(oldValue))
}

func TestEtcd_Update(t *testing.T) {
	config := EtcdConfig{
		Endpoints:      []string{"192.168.0.200:2379", "192.168.0.200:2369", "192.168.0.200:2389"},
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    2 * time.Second,
	}

	if err := InitEtcd(config); err != nil {
		t.Error("初始化etcd失败", err)
	}

	success, err := etcd.Update("root/oof", "oof2", "oof")
	if err != nil {
		t.Error(err)
	}
	t.Log("success: ", success)
}

func TestEtcd_Delete(t *testing.T) {
	config := EtcdConfig{
		Endpoints:      []string{"192.168.0.200:2379", "192.168.0.200:2369", "192.168.0.200:2389"},
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    2 * time.Second,
	}

	if err := InitEtcd(config); err != nil {
		t.Error("初始化etcd失败", err)
	}

	if err := etcd.Delete("root/oof"); err != nil {
		t.Error(err)
	}
}

func TestEtcd_Watch(t *testing.T) {
	config := EtcdConfig{
		Endpoints:      []string{"192.168.0.200:2379", "192.168.0.200:2369", "192.168.0.200:2389"},
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    2 * time.Second,
	}

	if err := InitEtcd(config); err != nil {
		t.Error("初始化etcd失败", err)
	}

	keyChangeEventResponse := etcd.Watch("root/foo")

	go func() {
		for event := range keyChangeEventResponse.Event {
			t.Log(event.Key, ":", event.Type, "=", string(event.Value))
		}
	}()

	_ = etcd.Put("root/foo", "1")
	time.Sleep(time.Second)
	_ = etcd.Put("root/foo", "2")
	time.Sleep(time.Second)

	_ = keyChangeEventResponse.Watcher.Close()
}

func TestEtcd_TxWithTTL(t *testing.T) {
	config := EtcdConfig{
		Endpoints:      []string{"192.168.0.200:2379", "192.168.0.200:2369", "192.168.0.200:2389"},
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    2 * time.Second,
	}

	if err := InitEtcd(config); err != nil {
		t.Error("初始化etcd失败", err)
	}

	txResponse, err := etcd.TxWithTTL("root/ttlkey", "ttlvalue", 5)
	if err != nil {
		t.Error(err)
	}

	t.Log("success: ", txResponse.Success)

}

func TestEtcd_TxKeepaliveWithTTL(t *testing.T) {
	config := EtcdConfig{
		Endpoints:      []string{"192.168.0.200:2379", "192.168.0.200:2369", "192.168.0.200:2389"},
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    2 * time.Second,
	}

	if err := InitEtcd(config); err != nil {
		t.Error("初始化etcd失败", err)
	}

	txResponse, err := etcd.TxKeepaliveWithTTL("root/ttlkey", "ttlvalue", 5)

	if err != nil {
		t.Error(err)
	}

	t.Log("success: ", txResponse.Success)

	time.Sleep(time.Minute)
	_ = txResponse.Lease.Close()
}

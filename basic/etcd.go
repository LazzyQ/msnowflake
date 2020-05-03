package basic

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

// 定义key变更事件常量
const (
	KeyCreateChangeEvent = iota
	KeyUpdateChangeEvent
	KeyDeleteChangeEvent
)

var (
	etcd *Etcd
)

type EtcdConfig struct {
	Endpoints      []string
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
}

type Etcd struct {
	endpoints []string
	client    *clientv3.Client
	kv        clientv3.KV
	timeout   time.Duration
}

type KeyChangeEvent struct {
	Type  int
	Key   string
	Value []byte
}

type WatchKeyChangeResponse struct {
	Event      chan *KeyChangeEvent
	CancelFunc context.CancelFunc
	Watcher    clientv3.Watcher
}

type TxResponse struct {
	Success bool
	LeaseID clientv3.LeaseID
	Lease   clientv3.Lease
	Key     string
	Value   string
}

func InitEtcd(config EtcdConfig) error {
	if etcd != nil {
		return nil
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: config.ConnectTimeout,
	})

	if err != nil {
		return err
	}

	etcd = &Etcd{
		endpoints: config.Endpoints,
		client:    client,
		kv:        clientv3.NewKV(client),
		timeout:   config.ReadTimeout,
	}

	return nil
}

func GetEtcd() *Etcd {
	return etcd
}

// 根据key获取value
func (etcd *Etcd) Get(key string) (value []byte, err error) {
	var (
		getResponse *clientv3.GetResponse
	)
	ctx, cancelFunc := context.WithTimeout(context.Background(), etcd.timeout)
	defer cancelFunc()
	if getResponse, err = etcd.kv.Get(ctx, key); err != nil {
		return
	}
	if len(getResponse.Kvs) == 0 {
		return
	}
	value = getResponse.Kvs[0].Value
	return
}

// 根据key前缀获取value列表
func (etcd *Etcd) GetWithPrefixKey(prefixKey string) (keys [][]byte, values [][]byte, err error) {
	var (
		getResponse *clientv3.GetResponse
	)
	ctx, cancelFunc := context.WithTimeout(context.Background(), etcd.timeout)
	defer cancelFunc()

	if getResponse, err = etcd.kv.Get(ctx, prefixKey, clientv3.WithPrefix()); err != nil {
		return
	}

	if len(getResponse.Kvs) == 0 {
		return
	}
	keys = make([][]byte, 0)
	values = make([][]byte, 0)

	for i := 0; i < len(getResponse.Kvs); i++ {
		keys = append(keys, getResponse.Kvs[i].Key)
		values = append(values, getResponse.Kvs[i].Value)
	}
	return
}

// 根据key前缀获取指定条数
func (etcd *Etcd) GetWithPrefixKeyLimit(prefixKey string, limit int64) (keys [][]byte, values [][]byte, err error) {
	var (
		getResponse *clientv3.GetResponse
	)
	ctx, cancelFunc := context.WithTimeout(context.Background(), etcd.timeout)
	defer cancelFunc()

	if getResponse, err = etcd.kv.Get(ctx, prefixKey, clientv3.WithPrefix(), clientv3.WithLimit(limit)); err != nil {
		return
	}

	if len(getResponse.Kvs) == 0 {
		return
	}
	keys = make([][]byte, 0)
	values = make([][]byte, 0)

	for i := 0; i < len(getResponse.Kvs); i++ {
		keys = append(keys, getResponse.Kvs[i].Key)
		values = append(values, getResponse.Kvs[i].Value)
	}
	return
}

// put一个值
func (etcd *Etcd) Put(key, value string) (err error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), etcd.timeout)
	defer cancelFunc()

	if _, err = etcd.client.Put(ctx, key, value); err != nil {
		return
	}
	return
}

// Put一个不存在的值
func (etcd *Etcd) PutNotExist(key, value string) (success bool, oldValue []byte, err error) {
	var (
		txnResponse *clientv3.TxnResponse
	)
	ctx, cancelFunc := context.WithTimeout(context.Background(), etcd.timeout)
	defer cancelFunc()

	txn := etcd.client.Txn(ctx)
	txnResponse, err = txn.If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, value)).
		Else(clientv3.OpGet(key)).
		Commit()

	if err != nil {
		return
	}

	if txnResponse.Succeeded {
		success = true
	} else {
		oldValue = make([]byte, 0)
		oldValue = txnResponse.Responses[0].GetResponseRange().Kvs[0].Value
	}
	return
}

// 更新一个已经存在的值
func (etcd *Etcd) Update(key, value, oldValue string) (success bool, err error) {
	var (
		txnResponse *clientv3.TxnResponse
	)
	ctx, cancelFunc := context.WithTimeout(context.Background(), etcd.timeout)
	defer cancelFunc()

	txn := etcd.client.Txn(ctx)
	txnResponse, err = txn.If(clientv3.Compare(clientv3.Value(key), "=", oldValue)).
		Then(clientv3.OpPut(key, value)).
		Commit()

	if err != nil {
		return
	}

	return txnResponse.Succeeded, err
}

// 根据key删除
func (etcd *Etcd) Delete(key string) (err error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), etcd.timeout)
	defer cancelFunc()

	_, err = etcd.kv.Delete(ctx, key)
	return
}

// 根据一个key前缀删除
func (etcd *Etcd) DeleteWithPrefixKey(prefixKey string) (err error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), etcd.timeout)
	defer cancelFunc()
	_, err = etcd.client.Delete(ctx, prefixKey, clientv3.WithPrefix())
	return
}

// watch 一个key
func (etcd *Etcd) Watch(key string) (keyChangeEventResponse *WatchKeyChangeResponse) {
	watcher := clientv3.NewWatcher(etcd.client)
	watchChans := watcher.Watch(context.Background(), key)

	keyChangeEventResponse = &WatchKeyChangeResponse{
		Event:   make(chan *KeyChangeEvent, 250),
		Watcher: watcher,
	}

	go func() {
		for ch := range watchChans {
			if ch.Canceled {
				goto End
			}
			for _, event := range ch.Events {
				etcd.handleKeyChangeEvent(event, keyChangeEventResponse.Event)
			}
		}
	End:
		// log.Println("the watcher lose for key:", key)
	}()
	return
}

// watch一个key前缀
func (etcd *Etcd) WatchWithPrefixKey(prefixKey string) (keyChangeEventResponse *WatchKeyChangeResponse) {
	watcher := clientv3.NewWatcher(etcd.client)
	watchChans := watcher.Watch(context.Background(), prefixKey, clientv3.WithPrefix())

	keyChangeEventResponse = &WatchKeyChangeResponse{
		Event:   make(chan *KeyChangeEvent, 250),
		Watcher: nil,
	}

	go func() {
		for ch := range watchChans {
			if ch.Canceled {
				goto End
			}
			for _, event := range ch.Events {
				etcd.handleKeyChangeEvent(event, keyChangeEventResponse.Event)
			}
		}
	End:
		// log
	}()
	return
}

// 创建一个指定时间的临时key
func (etcd *Etcd) TxWithTTL(key, value string, ttl int64) (txResponse *TxResponse, err error) {
	var (
		txnResponse *clientv3.TxnResponse
		leaseID     clientv3.LeaseID
		v           []byte
	)
	lease := clientv3.NewLease(etcd.client)
	grantResponse, err := lease.Grant(context.Background(), ttl)
	if err != nil {
		return
	}
	leaseID = grantResponse.ID

	ctx, cancelFunc := context.WithTimeout(context.Background(), etcd.timeout)
	defer cancelFunc()

	txn := etcd.client.Txn(ctx)
	txnResponse, err = txn.If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, value, clientv3.WithLease(leaseID))).
		Commit()

	if err != nil {
		_ = lease.Close()
		return
	}

	txResponse = &TxResponse{
		LeaseID: leaseID,
		Lease:   lease,
	}
	if txnResponse.Succeeded {
		txResponse.Success = true
	} else {
		_ = lease.Close()
		v, err = etcd.Get(key)
		if err != nil {
			return
		}
		txResponse.Success = false
		txResponse.Key = key
		txResponse.Value = string(v)
	}
	return
}

// 创建一个不间断续约的临时key
func (etcd *Etcd) TxKeepaliveWithTTL(key, value string, ttl int64) (txResponse *TxResponse, err error) {
	var (
		txnResponse   *clientv3.TxnResponse
		leaseId       clientv3.LeaseID
		aliveResponse <-chan *clientv3.LeaseKeepAliveResponse
		v             []byte
	)
	lease := clientv3.NewLease(etcd.client)

	grantResponse, err := lease.Grant(context.Background(), ttl)
	if err != nil {
		return
	}
	leaseId = grantResponse.ID
	if aliveResponse, err = lease.KeepAlive(context.Background(), leaseId); err != nil {
		return
	}

	go func() {
		for ch := range aliveResponse {
			if ch == nil {
				goto End
			}
		}
	End:
		// log
	}()

	ctx, cancelFunc := context.WithTimeout(context.Background(), etcd.timeout)
	defer cancelFunc()

	txn := etcd.client.Txn(ctx)
	txnResponse, err = txn.If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, value, clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(key)).Commit()

	if err != nil {
		_ = lease.Close()
		return
	}
	txResponse = &TxResponse{
		LeaseID: leaseId,
		Lease:   lease,
	}

	if txnResponse.Succeeded {
		txResponse.Success = true
	} else {
		_ = lease.Close()
		txResponse.Success = false
		if v, err = etcd.Get(key); err != nil {
			return
		}
		txResponse.Key = key
		txResponse.Value = string(v)
	}
	return
}

func (etcd *Etcd) Close() {
	etcd.client.Close()
}

func (etcd *Etcd) handleKeyChangeEvent(event *clientv3.Event, events chan *KeyChangeEvent) {
	changeEvent := &KeyChangeEvent{
		Key: string(event.Kv.Key),
	}

	switch event.Type {
	case mvccpb.PUT:
		if event.IsCreate() {
			changeEvent.Type = KeyCreateChangeEvent
		} else {
			changeEvent.Type = KeyUpdateChangeEvent
		}
		changeEvent.Value = event.Kv.Value
	case mvccpb.DELETE:
		changeEvent.Type = KeyDeleteChangeEvent
	}
	events <- changeEvent
}

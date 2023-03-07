package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type EtcdConfig struct {
	Endpoints   []string      `json:"endpoints"`
	DialTimeout time.Duration `json:"dialTimeout"`
}

type EtcdClient struct {
	config  *EtcdConfig
	etcdCli *clientv3.Client
	lease   clientv3.Lease
}

type LockResult struct {
	IsLock  bool
	LeaseID clientv3.LeaseID
	Cancel  context.CancelFunc
}

func InitEtcdClient(c *EtcdConfig) (*EtcdClient, error) {
	config := clientv3.Config{
		Endpoints:   c.Endpoints,
		DialTimeout: c.DialTimeout,
	}
	cli, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}
	lease := clientv3.NewLease(cli)
	return &EtcdClient{
		config:  c,
		etcdCli: cli,
		lease:   lease,
	}, nil
}
func (cli *EtcdClient) Close() error {
	return cli.etcdCli.Close()
}

func (cli *EtcdClient) Put(key, metadata string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 1 * time.Second
	}
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	_, err := cli.etcdCli.Put(ctx, key, metadata)
	if err != nil {
		return err
	}
	return nil
}

func (cli *EtcdClient) GetMetadata(key string, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = 1 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	resp, err := cli.etcdCli.Get(ctx, key)
	cancel()
	if err != nil {
		return "", err
	}
	return string(resp.Kvs[0].Value), nil
}

func (cli *EtcdClient) GetNodes(key string, timeout time.Duration) ([]string, error) {
	if timeout <= 0 {
		timeout = 1 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	resp, err := cli.etcdCli.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}
	var nodes []string
	for _, item := range resp.Kvs {
		nodes = append(nodes, string(item.Key))
	}
	return nodes, nil
}

func (cli *EtcdClient) Update(key, update string, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = 1 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	resp, err := cli.etcdCli.Put(ctx, key, update, clientv3.WithPrevKV())
	cancel()
	if err != nil {
		return "", err
	}
	return string(resp.PrevKv.Value), nil
}

func (cli *EtcdClient) Delete(key string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 1 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_, err := cli.etcdCli.Delete(ctx, key)
	cancel()
	if err != nil {
		return err
	}
	return nil
}

func (cli *EtcdClient) Watch(key string, cb func(int32, string, string)) {
	go func() {
		watchKeys := cli.etcdCli.Watch(context.Background(), key)
		for resp := range watchKeys {
			for _, item := range resp.Events {
				cb(int32(item.Type), string(item.Kv.Key), string(item.Kv.Value))
			}
		}
	}()
}

// etcd 分布式锁
func (cli *EtcdClient) Lock(key, val string, ttl int64) (*LockResult, error) {
	// 取消续租，释放租约
	release := func(leaseID clientv3.LeaseID, cancel context.CancelFunc) {
		cancel()
		_, _ = cli.lease.Revoke(context.Background(), leaseID)
	}
	// 创建租约
	leaseResp, err := cli.lease.Grant(context.Background(), ttl)
	if err != nil {
		return nil, err
	}
	leaseID := leaseResp.ID
	// 自动续租
	ctx, cancel := context.WithCancel(context.Background())
	if _, err = cli.lease.KeepAlive(ctx, leaseID); err != nil {
		release(leaseID, cancel)
		return nil, err
	}

	txn := clientv3.NewKV(cli.etcdCli).Txn(context.Background())
	txn = txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, val, clientv3.WithLease(leaseID))).
		Else()
	txnRest, err := txn.Commit()
	if err != nil {
		release(leaseID, cancel)
		return nil, err
	}
	if !txnRest.Succeeded {
		release(leaseID, cancel)
		return &LockResult{IsLock: false}, nil
	}
	return &LockResult{
		IsLock:  true,
		LeaseID: leaseID,
		Cancel:  cancel,
	}, err
}

func (cli *EtcdClient) Unlock(res *LockResult) {
	res.Cancel()
	_, _ = cli.lease.Revoke(context.Background(), res.LeaseID)
}

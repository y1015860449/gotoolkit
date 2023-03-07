package hxetcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"strconv"
	"time"
)

type Register struct {
	etcdConfig  *EtcdConfig
	svcInfo     *ServiceInfo
	svcTTL      int64
	etcdCli     *clientv3.Client
	leaseId     clientv3.LeaseID
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse
}

func NewRegister(etcdConfig *EtcdConfig) (*Register, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdConfig.Endpoints,
		DialTimeout: etcdConfig.DialTimeout,
	})
	if err != nil {
		return nil, err
	}
	return &Register{etcdConfig: etcdConfig, etcdCli: cli}, nil
}

func (r *Register) ServiceRegister(svcInfo *ServiceInfo, ttl int64) error {
	r.svcInfo = svcInfo
	r.svcTTL = ttl
	err := r.register()
	if err != nil {
		return err
	}
	go r.keepAlive()
	return nil
}

func (r *Register) register() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := r.etcdCli.Grant(ctx, r.svcTTL)
	if err != nil {
		return err
	}
	r.keepAliveCh, err = r.etcdCli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}
	serviceValue := net.JoinHostPort(r.svcInfo.SvcIp, strconv.Itoa(r.svcInfo.SvcPort))
	serviceKey := GetPrefix(schemeName, r.svcInfo.svcName) + "/" + serviceValue
	if _, err = r.etcdCli.Put(ctx, serviceKey, serviceValue, clientv3.WithLease(resp.ID)); err != nil {
		return err
	}
	if r.leaseId > 0 {
		_, _ = r.etcdCli.Revoke(ctx, r.leaseId)
	}
	r.leaseId = resp.ID
	return nil
}

func (r *Register) keepAlive() {
	t := time.NewTicker(time.Duration(r.svcTTL/2) * time.Second)
	for {
		select {
		case _, ok := <-r.keepAliveCh:
			if !ok {
				if err := r.register(); err != nil {
					continue
				}
			}
		case <-t.C:
			if r.keepAliveCh == nil {
				if err := r.register(); err != nil {
					continue
				}
			}
		}
	}

}

func (r *Register) ServiceDeregister() error {
	serviceValue := net.JoinHostPort(r.svcInfo.SvcIp, strconv.Itoa(r.svcInfo.SvcPort))
	serviceKey := GetPrefix(schemeName, r.svcInfo.svcName) + "/" + serviceValue
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := r.etcdCli.Delete(ctx, serviceKey, clientv3.WithLease(r.leaseId)); err != nil {
		return err
	}
	if _, err := r.etcdCli.Revoke(context.Background(), r.leaseId); err != nil {
		return err
	}
	return nil
}

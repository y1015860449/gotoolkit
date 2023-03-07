package hxetcd

import (
	"context"
	"errors"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"strings"
	"sync"
	"time"
)

type Resolver struct {
	etcdConfig         *EtcdConfig
	svcName            string
	cc                 resolver.ClientConn
	grpcClientConn     *grpc.ClientConn
	cli                *clientv3.Client
	watchStartRevision int64
}

var (
	nameResolver        = make(map[string]*Resolver)
	rwNameResolverMutex sync.RWMutex
)

func NewResolver(etcdConfig *EtcdConfig, svcName string) (*Resolver, error) {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdConfig.Endpoints,
		DialTimeout: etcdConfig.DialTimeout,
	})
	if err != nil {
		return nil, err
	}

	var r Resolver
	r.etcdConfig = etcdConfig
	r.svcName = svcName
	r.cli = etcdCli
	resolver.Register(&r)

	conn, err := grpc.Dial(
		GetPrefix(schemeName, svcName),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithInsecure(),
		grpc.WithTimeout(time.Duration(5)*time.Second),
	)
	if err == nil {
		r.grpcClientConn = conn
	}
	return &r, err
}

func (r *Resolver) ResolveNow(rn resolver.ResolveNowOptions) {
}

func (r *Resolver) Close() {
}

func GetConn(etcdConfig *EtcdConfig, svcName string) *grpc.ClientConn {
	rwNameResolverMutex.RLock()
	r, ok := nameResolver[schemeName+svcName]
	rwNameResolverMutex.RUnlock()
	if ok {
		return r.grpcClientConn
	}

	rwNameResolverMutex.Lock()
	r, ok = nameResolver[schemeName+svcName]
	if ok {
		rwNameResolverMutex.Unlock()
		return r.grpcClientConn
	}

	r, err := NewResolver(etcdConfig, svcName)
	if err != nil {
		rwNameResolverMutex.Unlock()
		return nil
	}

	nameResolver[schemeName+svcName] = r
	rwNameResolverMutex.Unlock()
	return r.grpcClientConn
}

func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	if r.cli == nil {
		return nil, errors.New("etcd clientv3 client failed")
	}
	r.cc = cc

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	//     "%s:///%s"
	prefix := GetPrefix(schemeName, r.svcName)
	// get key first
	resp, err := r.cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err == nil {
		var addrList []resolver.Address
		for i := range resp.Kvs {
			addrList = append(addrList, resolver.Address{Addr: string(resp.Kvs[i].Value)})
		}
		r.cc.UpdateState(resolver.State{Addresses: addrList})
		r.watchStartRevision = resp.Header.Revision + 1
		go r.watch(prefix, addrList)
	} else {
		return nil, err
	}

	return r, nil
}

func (r *Resolver) Scheme() string {
	return schemeName
}

func exists(addrList []resolver.Address, addr string) bool {
	for _, v := range addrList {
		if v.Addr == addr {
			return true
		}
	}
	return false
}

func remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}

func (r *Resolver) watch(prefix string, addrList []resolver.Address) {
	rch := r.cli.Watch(context.Background(), prefix, clientv3.WithPrefix(), clientv3.WithPrefix())
	for n := range rch {
		flag := 0
		for _, ev := range n.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				if !exists(addrList, string(ev.Kv.Value)) {
					flag = 1
					addrList = append(addrList, resolver.Address{Addr: string(ev.Kv.Value)})
				}
			case clientv3.EventTypeDelete:
				i := strings.LastIndexAny(string(ev.Kv.Key), "/")
				if i < 0 {
					return
				}
				t := string(ev.Kv.Key)[i+1:]
				if s, ok := remove(addrList, t); ok {
					flag = 1
					addrList = s
				}
			}
		}

		if flag == 1 {
			r.cc.UpdateState(resolver.State{Addresses: addrList})
		}
	}
}

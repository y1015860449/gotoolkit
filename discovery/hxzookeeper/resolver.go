package hxzookeeper

import (
	"errors"
	"fmt"
	"github.com/go-zookeeper/zk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"gotoolkit/discovery/balancer/hash"
	"time"
)

type Resolver struct {
	conn           *zk.Conn
	conf           *ZkConfig
	svcName        string
	cc             resolver.ClientConn
	grpcClientConn *grpc.ClientConn
}

func NewResolver(zkConfig *ZkConfig, svcName string) (*Resolver, error) {
	conn, _, err := zk.Connect(zkConfig.Urls, zkConfig.Timeout)
	if err != nil {
		return nil, err
	}
	r := &Resolver{
		conn:    conn,
		conf:    zkConfig,
		svcName: svcName,
	}
	resolver.Register(r)
	grpcConn, err := grpc.Dial(
		fmt.Sprintf("%s:///%s", schemeName, svcName),
		grpc.WithResolvers(r),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, hash.Name)),
		grpc.WithTimeout(time.Duration(5)*time.Second),
	)
	if err == nil {
		r.grpcClientConn = grpcConn
	}
	return r, nil
}

func (rlv *Resolver) ResolveNow(rn resolver.ResolveNowOptions) {
	_ = rn
}

func (rlv *Resolver) Close() {
}

func (rlv *Resolver) GetConn() *grpc.ClientConn {
	return rlv.grpcClientConn
}

func (rlv *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	_ = target
	_ = opts
	if rlv.conn == nil {
		return nil, errors.New("zookeeper client failed")
	}
	rlv.cc = cc
	prefix := fmt.Sprintf("/%s/%s", schemeName, rlv.svcName)
	nodes, _, err := rlv.conn.Children(prefix)
	if err == nil {
		var addrList []resolver.Address
		for _, node := range nodes {
			addrList = append(addrList, resolver.Address{Addr: node})
		}
		_ = rlv.cc.UpdateState(resolver.State{Addresses: addrList})
		go rlv.watch(prefix, addrList)
	} else {
		return nil, err
	}

	return rlv, nil
}

func (rlv *Resolver) Scheme() string {
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

func (rlv *Resolver) watch(prefix string, addrList []resolver.Address) {
	for {
		flag := 0
		snapshot, _, ch, err := rlv.conn.ChildrenW(prefix)
		if err != nil {
			return
		}
		select {
		case e := <-ch:
			switch e.Type {
			case zk.EventNodeDeleted:
				for _, node := range snapshot {
					if s, ok := remove(addrList, node); ok {
						flag = 1
						addrList = s
					}
				}
			case zk.EventNodeCreated:
				for _, node := range snapshot {
					if !exists(addrList, node) {
						flag = 1
						addrList = append(addrList, resolver.Address{Addr: node})
					}
				}
			case zk.EventNodeChildrenChanged:
				snapshot, _, err = rlv.conn.Children(prefix)
				if err == nil {
					flag = 1
					addrList = addrList[0:0]
					for _, node := range snapshot {
						addrList = append(addrList, resolver.Address{Addr: node})
					}
				}
			}

		}
		// 更新
		if flag == 1 {
			_ = rlv.cc.UpdateState(resolver.State{Addresses: addrList})
		}
	}
}

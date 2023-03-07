package hxconsul

import (
	"errors"
	"fmt"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/y1015860449/gotoolkit/discovery/balancer/hash"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"net"
	"strconv"
	"time"
)

type Resolver struct {
	client         *consulApi.Client
	consulConf     *ConsulConfig
	SvcName        string
	Tag            string
	cc             resolver.ClientConn
	grpcClientConn *grpc.ClientConn
	lastIndex      uint64
}

func NewResolver(consulConf *ConsulConfig, svcName, tag string) (*Resolver, error) {
	c := consulApi.DefaultConfig()
	c.Address = consulConf.Address
	cli, err := consulApi.NewClient(c)
	if err != nil {
		return nil, err
	}
	r := &Resolver{
		client:     cli,
		consulConf: consulConf,
		SvcName:    svcName,
		lastIndex:  0,
	}
	resolver.Register(r)

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:///%s", schemeName, svcName),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, hash.Name)),
		grpc.WithInsecure(),
		grpc.WithTimeout(time.Duration(5)*time.Second),
	)
	if err == nil {
		r.grpcClientConn = conn
	}
	return r, nil
}

func (rlv *Resolver) GetConn() *grpc.ClientConn {
	return rlv.grpcClientConn
}

func (rlv *Resolver) Scheme() string {
	return schemeName
}

func (rlv *Resolver) ResolveNow(options resolver.ResolveNowOptions) {
}

func (rlv *Resolver) Close() {
}

func (rlv *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	if rlv.client == nil {
		return nil, errors.New("consul client failed")
	}
	rlv.cc = cc
	addrList, err := rlv.serviceList()
	if err != nil {
		return nil, err
	}
	_ = rlv.cc.UpdateState(resolver.State{Addresses: addrList})
	go rlv.watch(addrList)
	return rlv, nil
}

func (rlv *Resolver) serviceList() ([]resolver.Address, error) {
	resp, mate, err := rlv.client.Health().Service(rlv.SvcName, "", true, &consulApi.QueryOptions{
		WaitIndex: rlv.lastIndex,
	})
	if err != nil {
		return nil, err
	}
	var addrList []resolver.Address
	for _, svc := range resp {
		addrList = append(addrList, resolver.Address{Addr: net.JoinHostPort(svc.Service.Address, strconv.Itoa(svc.Service.Port))})
	}
	rlv.lastIndex = mate.LastIndex
	return addrList, nil
}

func (rlv *Resolver) watch(addrList []resolver.Address) {
	ticker := time.NewTicker(1 * time.Second)
	var err error
	for {
		select {
		case <-ticker.C:
			addrList, err = rlv.serviceList()
			if err != nil {
				continue
			}
			if err = rlv.cc.UpdateState(resolver.State{Addresses: addrList}); err != nil {
				continue
			}
		}
	}
}

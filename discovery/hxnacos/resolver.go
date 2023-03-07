package hxnacos

import (
	"context"
	"errors"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"gotoolkit/discovery/balancer/hash"
	"log"
	"net"
	"sort"
	"time"
)

type Resolver struct {
	namingClient   naming_client.INamingClient
	nacosConfig    *NacosConfig
	SvcName        string
	groupName      string
	cc             resolver.ClientConn
	grpcClientConn *grpc.ClientConn
	cancelFunc     context.CancelFunc
}

func NewResolver(config *NacosConfig, svcName, groupName string) (*Resolver, error) {
	serverConfig, clientConfig, err := getNacosSdkConfig(config)
	if err != nil {
		return nil, err
	}
	param := vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfig,
	}
	namingClient, err := clients.NewNamingClient(param)
	if err != nil {
		return nil, err
	}
	r := &Resolver{
		namingClient: namingClient,
		nacosConfig:  config,
		SvcName:      svcName,
		groupName:    groupName,
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
	rlv.cancelFunc()
}

func (rlv *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	if rlv.namingClient == nil {
		return nil, errors.New("nacos client failed")
	}
	rlv.cc = cc
	addrList, err := rlv.serviceList()
	if err != nil {
		return nil, err
	}
	_ = rlv.cc.UpdateState(resolver.State{Addresses: addrList})

	param := &vo.SubscribeParam{
		ServiceName:       rlv.SvcName,
		SubscribeCallback: rlv.CallBackHandle,
	}
	if len(rlv.groupName) > 0 {
		param.GroupName = rlv.groupName
	}
	go rlv.namingClient.Subscribe(param)
	return rlv, nil
}

func (rlv *Resolver) serviceList() ([]resolver.Address, error) {
	instances, err := rlv.namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: rlv.SvcName,
		GroupName:   rlv.groupName,
	})
	if err != nil {
		return nil, err
	}
	var addrList []resolver.Address
	for _, instance := range instances {
		if instance.Healthy && instance.Enable {
			addrList = append(addrList, resolver.Address{Addr: net.JoinHostPort(instance.Ip, fmt.Sprintf("%d", instance.Port))})
		}
	}
	if len(addrList) > 0 {
		sort.Sort(byAddressString(addrList))
	}
	return addrList, nil
}

func (rlv *Resolver) CallBackHandle(services []model.Instance, err error) {
	if err != nil {
		log.Printf("[Nacos resolver] watcher call back handle error:%v", err)
		return
	}
	addrList, _ := rlv.serviceList()
	_ = rlv.cc.UpdateState(resolver.State{Addresses: addrList})
}

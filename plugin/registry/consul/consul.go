package consul

import (
	"errors"
	consulApi "github.com/hashicorp/consul/api"
)

type ConsulConfig struct {
	Address string
}

type HealthConfig struct {
	HealthType                     string // 健康检查类型 http tcp udp grpc
	HealthUri                      string // 健康检查访问uri
	CheckID                        string // 健康检查唯一id
	Interval                       string // 时间间隔 5s
	Timeout                        string // 超时时间 1s
	DeregisterCriticalServiceAfter string // 超过多少时间注销服务
}

type RegisterConfig struct {
	SvcId   string        // 服务id
	SvcName string        // 服务名称
	Address string        // 服务ip
	Port    int           // 服务端口
	health  *HealthConfig // 健康检查配置
}

type ServiceInfo struct {
	Ip   string
	Port int
}

type ConsulClient struct {
	client       *consulApi.Client
	consulConf   *ConsulConfig
	registerConf *RegisterConfig
}

func InitConsul(c *ConsulConfig) (*ConsulClient, error) {
	config := consulApi.DefaultConfig()
	config.Address = c.Address
	cli, err := consulApi.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ConsulClient{
		client:     cli,
		consulConf: c,
	}, nil
}

// 服务注册
func (cli *ConsulClient) ServiceRegister(c *RegisterConfig) error {
	if c == nil || c.health == nil {
		return errors.New("params is exception")
	}
	reg := &consulApi.AgentServiceRegistration{
		ID:      c.SvcId,
		Name:    c.SvcName,
		Port:    c.Port,
		Address: c.Address,
		Check: &consulApi.AgentServiceCheck{
			CheckID:                        c.health.CheckID,
			Interval:                       c.health.Interval,
			Timeout:                        c.health.Timeout,
			DeregisterCriticalServiceAfter: c.health.DeregisterCriticalServiceAfter,
		},
	}
	if c.health.HealthType == "tcp" {
		reg.Check.TCP = c.health.HealthUri
	} else if c.health.HealthType == "udp" {
		reg.Check.UDP = c.health.HealthUri
	} else if c.health.HealthType == "grpc" {
		reg.Check.GRPC = c.health.HealthUri
	} else {
		reg.Check.HTTP = c.health.HealthUri
	}
	err := cli.client.Agent().ServiceRegister(reg)
	if err != nil {
		return err
	}
	cli.registerConf = c
	return nil
}

// 服务注销
func (cli *ConsulClient) ServiceDeregister() error {
	return cli.client.Agent().ServiceDeregister(cli.registerConf.SvcId)
}

func (cli *ConsulClient) Deregister(svcId string) error {
	return cli.client.Agent().ServiceDeregister(svcId)
}

// 获取服务列表
func (cli *ConsulClient) ServiceList(svcName string) ([]ServiceInfo, error) {
	services, _, err := cli.client.Health().Service(svcName, "", true, nil)
	if err != nil {
		return nil, err
	}
	var infos []ServiceInfo
	for _, svc := range services {
		infos = append(infos, ServiceInfo{
			Ip:   svc.Service.Address,
			Port: svc.Service.Port,
		})
	}
	return infos, nil
}

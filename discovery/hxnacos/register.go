package hxnacos

import (
	"errors"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type Register struct {
	namingClient   naming_client.INamingClient
	nacosConfig    *NacosConfig
	registerConfig *RegisterConfig
}

func NewRegister(config *NacosConfig) (*Register, error) {
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
	return &Register{
		namingClient: namingClient,
		nacosConfig:  config,
	}, nil
}

func (register *Register) ServiceRegister(config *RegisterConfig) error {
	if config == nil || len(config.SvcName) <= 0 || config.SvcPort <= 0 {
		return errors.New("config is err")
	}
	param := vo.RegisterInstanceParam{
		Ip:          config.SvcIp,
		Port:        config.SvcPort,
		Weight:      10,
		Enable:      config.Enable,
		Healthy:     config.Healthy,
		Metadata:    config.Metadata,
		ServiceName: config.SvcName,
		GroupName:   config.GroupName,
		Ephemeral:   config.Ephemeral,
	}
	_, err := register.namingClient.RegisterInstance(param)
	return err
}

func (register *Register) ServiceDeregister() error {
	_, err := register.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          register.registerConfig.SvcIp,
		Port:        register.registerConfig.SvcPort,
		ServiceName: register.registerConfig.SvcName,
		GroupName:   register.registerConfig.GroupName,
		Ephemeral:   register.registerConfig.Ephemeral,
	})
	return err
}

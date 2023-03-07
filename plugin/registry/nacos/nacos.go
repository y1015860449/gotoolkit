package nacos

import (
	"errors"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"strconv"
	"strings"
)

type NacosConfig struct {
	Host        string // 地址 例:"http://127.0.0.1:8848;127.0.0.1:8849;127.0.0.1:8850"
	NameSpaceId string // 命名空间

	// 以下基本使用默认值，调用DefaultNacosConfig
	NotLoadCacheAtStart bool
	TimeoutMs           uint64 // 超时时间 毫秒
	CacheDir            string // 缓存地址
	LogDir              string // 日志地址
	LogLevel            string // 日志等级
}

func DefaultNacosConfig() *NacosConfig {
	return &NacosConfig{
		Host:                "http://127.0.0.1:8848",
		NameSpaceId:         "public",
		NotLoadCacheAtStart: false,
		TimeoutMs:           5000,
		CacheDir:            "./nacos/cache",
		LogDir:              "./nacos/log",
		LogLevel:            "warn",
	}
}

type NacosRegisterConfig struct {
	SvcIp     string            // 注册服务的ip
	SvcPort   uint64            // 注册服务的端口
	SvcName   string            // 注册服务名称
	GroupName string            // 服务组名
	Metadata  map[string]string // 数据

	// 以下基本使用默认值，调用DefaultNacosRegisterConfig
	Healthy   bool // 健康检查
	Enable    bool // 服务启用
	Ephemeral bool
}

func DefaultNacosRegisterConfig() *NacosRegisterConfig {
	return &NacosRegisterConfig{
		Healthy:   true,
		Enable:    true,
		Ephemeral: true,
	}
}

type NacosClient struct {
	configClient        config_client.IConfigClient
	namingClient        naming_client.INamingClient
	nacosConfig         *NacosConfig
	nacosRegisterConfig *NacosRegisterConfig
}

func InitNacos(config *NacosConfig) (*NacosClient, error) {
	serverConfig, clientConfig, err := GetNacosSdkConfig(config)
	if err != nil {
		return nil, err
	}
	param := vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfig,
	}
	configClient, err := clients.NewConfigClient(param)
	if err != nil {
		return nil, err
	}
	namingClient, err := clients.NewNamingClient(param)
	if err != nil {
		return nil, err
	}
	return &NacosClient{
		configClient: configClient,
		namingClient: namingClient,
		nacosConfig:  config,
	}, nil
}

func GetNacosSdkConfig(config *NacosConfig) ([]constant.ServerConfig, *constant.ClientConfig, error) {
	tmp := strings.Split(config.Host, "://")
	if len(tmp) <= 1 {
		return nil, nil, errors.New("config is err")
	}
	scheme := tmp[0]
	addrs := strings.Split(tmp[1], ";")
	serverConfig := make([]constant.ServerConfig, 0)
	for _, addr := range addrs {
		str := strings.Split(addr, ":")
		if len(str) <= 1 {
			return nil, nil, errors.New("config is err")
		}
		ip := str[0]
		port, _ := strconv.ParseUint(str[1], 10, 64)
		serverConfig = append(serverConfig, constant.ServerConfig{
			Scheme: scheme,
			IpAddr: ip,
			Port:   port,
		})
	}

	clientConfig := &constant.ClientConfig{
		NamespaceId:         config.NameSpaceId,
		NotLoadCacheAtStart: config.NotLoadCacheAtStart,
		TimeoutMs:           config.TimeoutMs,
		CacheDir:            config.CacheDir,
		LogDir:              config.LogDir,
		LogLevel:            config.LogLevel,
	}
	return serverConfig, clientConfig, nil
}

////////////////////////////////////////////////
// 注册发现中心使用
////////////////////////////////////////////////

func (client *NacosClient) RegisterInstance(config *NacosRegisterConfig) (bool, error) {
	if config == nil || len(config.SvcName) <= 0 || config.SvcPort <= 0 {
		return false, errors.New("config is err")
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
	success, err := client.namingClient.RegisterInstance(param)
	return success, err
}

func (client *NacosClient) DeregisterInstance() (bool, error) {
	if client.nacosRegisterConfig == nil {
		return false, nil
	}
	success, err := client.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          client.nacosRegisterConfig.SvcIp,
		Port:        client.nacosRegisterConfig.SvcPort,
		ServiceName: client.nacosRegisterConfig.SvcName,
		GroupName:   client.nacosRegisterConfig.GroupName,
		Ephemeral:   client.nacosRegisterConfig.Ephemeral,
	})
	return success, err
}

func (client *NacosClient) Deregister(svcIp string, svcPort uint64, svcName string) (bool, error) {
	success, err := client.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          svcIp,
		Port:        svcPort,
		ServiceName: svcName,
		Ephemeral:   true,
	})
	return success, err
}

func (client *NacosClient) UpdateMetaData(metadata map[string]string) (bool, error) {
	param := vo.UpdateInstanceParam{
		Ip:          client.nacosRegisterConfig.SvcIp,
		Port:        client.nacosRegisterConfig.SvcPort,
		Weight:      10,
		Enable:      client.nacosRegisterConfig.Enable,
		Metadata:    metadata,
		ServiceName: client.nacosRegisterConfig.SvcName,
		GroupName:   client.nacosRegisterConfig.GroupName,
		Ephemeral:   client.nacosRegisterConfig.Ephemeral,
	}
	success, err := client.namingClient.UpdateInstance(param)
	return success, err
}

func (client *NacosClient) SelectAllInstances(svcName, groupName string) ([]model.Instance, error) {
	if len(svcName) <= 0 {
		return nil, errors.New("param is err")
	}
	instances, err := client.namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: svcName,
		GroupName:   groupName,
	})
	return instances, err
}

func (client *NacosClient) SelectInstances(svcName, groupName string) ([]model.Instance, error) {
	if len(svcName) <= 0 {
		return nil, errors.New("param is err")
	}
	instances, err := client.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: svcName,
		GroupName:   groupName,
		HealthyOnly: true,
	})
	return instances, err
}

func (client *NacosClient) SelectOneHealthyInstance(svcName, groupName string) (*model.Instance, error) {
	if len(svcName) <= 0 {
		return nil, errors.New("param is err")
	}
	instances, err := client.namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: svcName,
		GroupName:   groupName,
	})
	return instances, err
}

func (client *NacosClient) Subscribe(svcName, groupName string, cb func(services []model.Instance, err error)) error {
	if len(svcName) <= 0 || cb == nil {
		return errors.New("param is err")
	}
	err := client.namingClient.Subscribe(&vo.SubscribeParam{
		ServiceName:       svcName,
		GroupName:         groupName,
		SubscribeCallback: cb,
	})
	return err
}

func (client *NacosClient) Unsubscribe(svcName, groupName string, cb func(services []model.Instance, err error)) error {
	if len(svcName) <= 0 {
		return errors.New("param is err")
	}
	err := client.namingClient.Unsubscribe(&vo.SubscribeParam{
		ServiceName:       svcName,
		GroupName:         groupName,
		SubscribeCallback: cb,
	})
	return err
}

////////////////////////////////////////////////
// 配置中心使用
////////////////////////////////////////////////

func (client *NacosClient) GetConfig(dataId, group string) (string, error) {
	if len(dataId) <= 0 || len(group) <= 0 {
		return "", errors.New("param is err")
	}
	return client.configClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
}

func (client *NacosClient) PublishConfig(dataId, group, content string) (bool, error) {
	if len(dataId) <= 0 || len(group) <= 0 || len(content) <= 0 {
		return false, errors.New("param is err")
	}
	return client.configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   group,
		Content: content,
	})
}

func (client *NacosClient) DeleteConfig(dataId, group string) (bool, error) {
	if len(dataId) <= 0 || len(group) <= 0 {
		return false, errors.New("param is err")
	}
	return client.configClient.DeleteConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
}

func (client *NacosClient) ListenConfig(dataId, group string, changeCb func(namespace, group, dataId, data string)) error {
	if len(dataId) <= 0 || len(group) <= 0 || changeCb == nil {
		return errors.New("param is err")
	}
	return client.configClient.ListenConfig(vo.ConfigParam{
		DataId:   dataId,
		Group:    group,
		OnChange: changeCb,
	})
}

func (client *NacosClient) CancelListenConfig(dataId, group string) error {
	if len(dataId) <= 0 || len(group) <= 0 {
		return errors.New("param is err")
	}
	return client.configClient.CancelListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
}

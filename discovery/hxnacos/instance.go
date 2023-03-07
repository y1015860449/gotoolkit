package hxnacos

import (
	"errors"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"google.golang.org/grpc/resolver"
	"strconv"
	"strings"
)

const schemeName = "nacos"

type NacosConfig struct {
	Host        string // 地址 例:"http://127.0.0.1:8848;127.0.0.1:8849;127.0.0.1:8850"
	NameSpaceId string // 命名空间

	// 以下基本使用默认值，调用DefaultNacosConfig
	NotLoadCacheAtStart bool
	TimeoutMs           uint64 // 超时时间 毫秒
	BeatInterval        int64  // 心跳时间
	CacheDir            string // 缓存地址
	LogDir              string // 日志地址
	LogLevel            string // 日志等级
}

func DefaultNacosConfig() *NacosConfig {
	return &NacosConfig{
		Host:                "http://127.0.0.1:8848",
		NameSpaceId:         "public",
		NotLoadCacheAtStart: true,
		TimeoutMs:           5000,
		BeatInterval:        2000,
		CacheDir:            "./nacos/cache",
		LogDir:              "./nacos/log",
		LogLevel:            "warn",
	}
}

type RegisterConfig struct {
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

func DefaultRegisterConfig() *RegisterConfig {
	return &RegisterConfig{
		Healthy:   true,
		Enable:    true,
		Ephemeral: true,
	}
}

func getNacosSdkConfig(config *NacosConfig) ([]constant.ServerConfig, *constant.ClientConfig, error) {
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
			Scheme:      scheme,
			ContextPath: constant.DEFAULT_CONTEXT_PATH,
			IpAddr:      ip,
			Port:        port,
		})
	}

	clientConfig := &constant.ClientConfig{
		NamespaceId:         config.NameSpaceId,
		NotLoadCacheAtStart: config.NotLoadCacheAtStart,
		TimeoutMs:           config.TimeoutMs,
		BeatInterval:        config.BeatInterval,
		CacheDir:            config.CacheDir,
		LogDir:              config.LogDir,
		LogLevel:            config.LogLevel,
	}
	return serverConfig, clientConfig, nil
}

type byAddressString []resolver.Address

func (p byAddressString) Len() int           { return len(p) }
func (p byAddressString) Less(i, j int) bool { return p[i].Addr < p[j].Addr }
func (p byAddressString) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

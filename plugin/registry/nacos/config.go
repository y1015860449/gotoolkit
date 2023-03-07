package nacos

import (
	v2 "gopkg.in/yaml.v2"
)

type ExternalConfig struct {
	ListenOn        string           `yaml:"listenOn"`
	Redis           *Redis           `yaml:"redis"`
	Mongodb         *Mongodb         `yaml:"mongodb"`
	Kafka           *Kafka           `yaml:"kafka"`
	NacosConf       *NacosConf       `yaml:"nacosConf"`
	NacosServicName *NacosServicName `yaml:"nacosServicName"`
	KeysPath        *KeysPath        `yaml:"keysPath"`
	Log             *Log             `yaml:"log"`
}

type Redis struct {
	Addr         []string `yaml:"addr"`         // 地址
	Type         string   `yaml:"type"`         // node/cluster
	Pwd          string   `yaml:"pwd"`          // 密码
	MaxRetries   int      `yaml:"maxRetries"`   // 最大尝试次数, 默认3
	MinIdleConns int      `yaml:"minIdleConns"` // 最初连接数， 默认8
}

type Mongodb struct {
	Uri         string `yaml:"uri"`
	MaxPoolSize uint64 `yaml:"maxPoolSize"`
	DbName      string `yaml:"dbName"`
}

type Kafka struct {
	Topic            string   `yaml:"topic"`
	TopicPush        string   `yaml:"topicPush"`
	TopicSessionTime string   `yaml:"topicSessionTime"`
	Group            string   `yaml:"group"`
	Brokers          []string `yaml:"brokers"`
}

type NacosConf struct {
	Host        string `yaml:"host"`
	GroupName   string `yaml:"groupName"`
	NameSpaceId string `yaml:"nameSpaceId"`
	ServiceName string `yaml:"serviceName"`
	ServiceIp   string `yaml:"serviceIp"`
	ServicePort uint64 `yaml:"servicePort"`
}

type NacosServicName struct {
	AuthServiceName string `yaml:"authServiceName"`
	UserServiceName string `yaml:"userServiceName"`
	ImcServiceName  string `yaml:"imcServiceName"`
	DmsServiceName  string `yaml:"dmsServiceName"`
}

type Log struct {
	LogPath    string `yaml:"logPath"`    // 日志文件路径
	LogLevel   string `yaml:"logLevel"`   //日志级别 debug/info/warn/error
	MaxBackups int    `yaml:"maxBackups"` // 保存的文件个数
	MaxAge     int    `yaml:"maxAge"`     // 保存的天数， 没有的话不删除
	ServerName string `yaml:"serverName"` // 服务名称
}

type KeysPath struct {
	Path string `yaml:"path"`
}

var (
	Config = ExternalConfig{}
)

func InitConfig(data []byte) (*ExternalConfig, error) {
	err := v2.Unmarshal(data, &Config)
	return &Config, err
}

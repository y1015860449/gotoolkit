package hxetcd

import (
	"fmt"
	"time"
)

const schemeName = "etcd"

type EtcdConfig struct {
	Endpoints   []string      `json:"endpoints"`
	DialTimeout time.Duration `json:"dialTimeout"`
}

type ServiceInfo struct {
	svcName string
	SvcIp   string
	SvcPort int
}

func GetPrefix(schema, serviceName string) string {
	return fmt.Sprintf("%s:///%s", schema, serviceName)
}

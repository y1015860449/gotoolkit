package hxzookeeper

import (
	"time"
)

const schemeName = "zookeeper"

type ZkConfig struct {
	Urls    []string
	Timeout time.Duration
}

type ServiceInfo struct {
	svcName string
	SvcIp   string
	SvcPort int
}

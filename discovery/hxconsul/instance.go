package hxconsul

import "time"

const schemeName = "consul"

type ConsulConfig struct {
	Address string
	Ttl     int
}

type RegisterConfig struct {
	SvcName        string        // 服务名称
	Address        string        // 服务ip
	Port           int           // 服务端口
	UpdateInterval time.Duration // 健康检查时间
	Tag            string        // 服务版本号,非必填
}

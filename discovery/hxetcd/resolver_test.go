package hxetcd

import (
	"context"
	"github.com/y1015860449/gotoolkit/discovery/pb/hello"
	"github.com/y1015860449/gotoolkit/log/zaplog"
	"testing"
	"time"
)

func TestGetConn(t *testing.T) {
	logConfig := zaplog.DefaultConfig()
	logConfig.LogPath = "./logs/resolver.log"
	zaplog.InitLogger(logConfig)
	conn := GetConn(&EtcdConfig{
		Endpoints:   []string{"192.168.20.99:2379"},
		DialTimeout: 5 * time.Second,
	}, "etcd_hello")
	cli := hello.NewHelloClient(conn)

	for {
		if resp, err := cli.SayHello(context.Background(), &hello.Request{Text: "hello"}); err != nil {
			zaplog.ZapLog.Errorf("err(%+v)", err)
		} else {
			zaplog.ZapLog.Infof("resp(%+v)", resp)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

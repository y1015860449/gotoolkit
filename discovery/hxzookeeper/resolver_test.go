package hxzookeeper

import (
	"context"
	"github.com/y1015860449/gotoolkit/discovery/balancer/hash"
	"github.com/y1015860449/gotoolkit/discovery/pb/hello"
	"github.com/y1015860449/gotoolkit/log/zaplog"
	"testing"
	"time"
)

func TestGetConn(t *testing.T) {
	logConfig := zaplog.DefaultConfig()
	logConfig.LogPath = "./logs/resolver.log"
	zaplog.InitLogger(logConfig)
	c := &ZkConfig{Urls: []string{"192.168.20.99:2181"}, Timeout: 5 * time.Second}
	rlv, err := NewResolver(c, "zk_hello")
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}
	conn := rlv.GetConn()
	cli := hello.NewHelloClient(conn)

	for {
		ctx := context.WithValue(context.Background(), hash.HashKey, "test")
		if resp, err := cli.SayHello(ctx, &hello.Request{Text: "hello"}); err != nil {
			zaplog.ZapLog.Errorf("err(%+v)", err)
		} else {
			zaplog.ZapLog.Infof("resp(%+v)", resp)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

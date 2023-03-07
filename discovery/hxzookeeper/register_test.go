package hxzookeeper

import (
	"context"
	"fmt"
	"github.com/y1015860449/gotoolkit/discovery/pb/hello"
	"github.com/y1015860449/gotoolkit/log/zaplog"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

type server struct {
	hello.UnimplementedHelloServer
}

func (s *server) SayHello(ctx context.Context, request *hello.Request) (*hello.Response, error) {
	return &hello.Response{Result: fmt.Sprintf("response: %s", request.Text)}, nil
}

func TestRegister_RegisterNacos(t *testing.T) {
	zaplog.InitLogger(nil)
	c := &ZkConfig{Urls: []string{"192.168.20.99:2181"}, Timeout: 5 * time.Second}
	reg, err := NewRegister(c)
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}

	lis, err := net.Listen("tcp", ":8868")
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}

	rc := &ServiceInfo{
		svcName: "zk_hello",
		SvcIp:   "192.168.166.125",
		SvcPort: 8868,
	}
	if err = reg.ServiceRegister(rc); err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}
	defer reg.ServiceDeregister()

	s := grpc.NewServer()
	hello.RegisterHelloServer(s, &server{})
	if err = s.Serve(lis); err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}
}

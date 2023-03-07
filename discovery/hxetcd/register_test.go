package hxetcd

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

func TestRegister_RegisterEtcd(t *testing.T) {
	zaplog.InitLogger(nil)
	reg, err := NewRegister(&EtcdConfig{
		Endpoints:   []string{"192.168.20.99:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}

	lis, err := net.Listen("tcp", ":8868")
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}

	if err = reg.ServiceRegister(&ServiceInfo{
		svcName: "etcd_hello",
		SvcIp:   "192.168.166.125",
		SvcPort: 8868,
	}, 5); err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}

	s := grpc.NewServer()
	hello.RegisterHelloServer(s, &server{})
	if err = s.Serve(lis); err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}
}

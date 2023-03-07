package hxconsul

import (
	"context"
	"fmt"
	"github.com/y1015860449/gotoolkit/discovery/pb/hello"
	"github.com/y1015860449/gotoolkit/log/zaplog"
	"google.golang.org/grpc"
	"net"
	"testing"
)

type server struct {
	hello.UnimplementedHelloServer
}

func (s *server) SayHello(ctx context.Context, request *hello.Request) (*hello.Response, error) {
	return &hello.Response{Result: fmt.Sprintf("response: %s", request.Text)}, nil
}

func TestRegister_RegisterConsul(t *testing.T) {
	zaplog.InitLogger(nil)
	reg, err := NewRegister(&ConsulConfig{
		Address: "192.168.20.99:8500",
		Ttl:     8,
	})
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}

	lis, err := net.Listen("tcp", ":8868")
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}

	if err = reg.ServiceRegister(&RegisterConfig{
		SvcName:        "consul_hello",
		Address:        "192.168.166.125",
		Port:           8868,
		UpdateInterval: 2,
	}); err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}
	defer reg.ServiceDeregister()

	s := grpc.NewServer()
	hello.RegisterHelloServer(s, &server{})
	if err = s.Serve(lis); err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}
}

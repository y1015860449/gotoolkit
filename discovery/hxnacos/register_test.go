package hxnacos

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

func TestRegister_RegisterNacos(t *testing.T) {
	zaplog.InitLogger(nil)
	c := DefaultNacosConfig()
	c.Host = "http://192.168.20.99:8848"
	c.NameSpaceId = "public"
	reg, err := NewRegister(c)
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}

	lis, err := net.Listen("tcp", ":8868")
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}

	rc := DefaultRegisterConfig()
	rc.SvcName = "nacos_hello"
	rc.SvcIp = "192.168.166.125"
	rc.SvcPort = 8868
	rc.Metadata = map[string]string{"preserved.register.source": "go-grpc"}
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

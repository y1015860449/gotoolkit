package nacos

import (
	"github.com/y1015860449/gotoolkit/log/zaplog"
	"testing"
)

func TestNacosRegister(t *testing.T) {
	zaplog.InitLogger(nil)
	c := DefaultNacosConfig()
	c.Host = "http://192.168.188.200:8848"
	cli, err := InitNacos(c)
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}
	cli.Deregister("192.168.166.125", 8868, "nacos_hello")
	//rc := DefaultNacosRegisterConfig()
	//rc.SvcIp = "127.0.0.1"
	//rc.SvcPort = 8898
	//rc.SvcName = "test.nacos"
	//_, err = cli.RegisterInstance(rc)
	//if err != nil {
	//	zaplog.ZapLog.Panicf("err(%+v)", err)
	//}
}

package zookeeper

import (
	"github.com/go-zookeeper/zk"
	"github.com/y1015860449/gotoolkit/log/zaplog"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	zaplog.InitLogger(nil)
	cli, err := InitZookeeper(&ZkConfig{
		Urls:    []string{"192.168.20.99:2181"},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}
	defer cli.Close()
	cli.Create("/zk/test/create", int32(zk.FlagEphemeral), []byte("create"))

}

// 测试分布式锁
func TestLocker(t *testing.T) {
	zaplog.InitLogger(nil)
	cli, err := InitZookeeper(&ZkConfig{
		Urls:    []string{"192.168.20.99:2181"},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}
	defer cli.Close()

	go func() {
		rest, err := cli.Lock("/zk/test1")
		if err != nil {
			return
		}
		zaplog.ZapLog.Infof("test1 success")
		time.Sleep(8 * time.Second)
		cli.Unlock(rest)
	}()

	go func() {
		rest, err := cli.Lock("/zk/test1")
		if err != nil {
			return
		}
		zaplog.ZapLog.Infof("test2 success")
		time.Sleep(8 * time.Second)
		cli.Unlock(rest)
	}()

	go func() {
		rest, err := cli.Lock("/zk/test3")
		if err != nil {
			return
		}
		zaplog.ZapLog.Infof("test3 success")
		time.Sleep(8 * time.Second)
		cli.Unlock(rest)
	}()

	for {
		rest, err := cli.Lock("/zk/test1")
		if err != nil {
			return
		}
		zaplog.ZapLog.Infof("test4 success")
		time.Sleep(8 * time.Second)
		cli.Unlock(rest)
	}
}

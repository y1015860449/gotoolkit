package etcd

import (
	"github.com/y1015860449/gotoolkit/log/zaplog"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	zaplog.InitLogger(nil)
	cli, err := InitEtcdClient(&EtcdConfig{
		Endpoints:   []string{"192.168.20.99:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}
	defer cli.Close()
	cli.Put("/etcd/test/put", "put", 5*time.Second)

}

// 测试分布式锁
func TestLocker(t *testing.T) {
	zaplog.InitLogger(nil)
	cli, err := InitEtcdClient(&EtcdConfig{
		Endpoints:   []string{"192.168.20.99:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		zaplog.ZapLog.Panicf("err(%+v)", err)
	}
	defer cli.Close()

	go func() {
		rest, err := cli.Lock("test1", "test1", 10)
		if err != nil {
			return
		}
		if !rest.IsLock {
			zaplog.ZapLog.Infof("test1 fail")
			return
		}
		zaplog.ZapLog.Infof("test1 success")
		time.Sleep(8 * time.Second)
		cli.Unlock(rest)
	}()

	go func() {
		rest, err := cli.Lock("test1", "test2", 10)
		if err != nil {
			return
		}
		if !rest.IsLock {
			zaplog.ZapLog.Infof("test2 fail")
			return
		}
		zaplog.ZapLog.Infof("test2 success")
		time.Sleep(8 * time.Second)
		cli.Unlock(rest)
	}()

	go func() {
		rest, err := cli.Lock("test3", "test3", 10)
		if err != nil {
			return
		}
		if !rest.IsLock {
			zaplog.ZapLog.Infof("test3 fail")
			return
		}
		zaplog.ZapLog.Infof("test3 success")
		time.Sleep(8 * time.Second)
		cli.Unlock(rest)
	}()

	for {
		rest, err := cli.Lock("test3", "test4", 10)
		if err != nil {
			return
		}
		if !rest.IsLock {
			zaplog.ZapLog.Infof("test4 fail")
			time.Sleep(2 * time.Second)
			continue
		}
		zaplog.ZapLog.Infof("test4 success")
		time.Sleep(8 * time.Second)
		cli.Unlock(rest)
	}
}

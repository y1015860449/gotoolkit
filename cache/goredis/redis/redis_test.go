package redis

import (
	"testing"
	"time"
)

var (
	cli *GoRedis
	err error
)

func InitEpoRedis() error {
	cli, err = InitRedis(&RedisConfig{
		Addr: []string{"192.168.220.128:6379"},
		Type: "node",
	})
	return err
}

func TestEpoRedis_SetNx(t *testing.T) {
	if err := InitEpoRedis(); err != nil {
		t.Errorf("err(%v)", err)
	}
	cli.Set("1", 1)
	value, _ := cli.Get("1")
	t.Logf("%s", value)
	cli.SetEx("2", 2, 10)
	value, _ = cli.Get("2")
	t.Logf("%s", value)
	cli.SetNx("3", 3)
	value, _ = cli.Get("3")
	t.Logf("%s", value)
	cli.SetNxEx("4", 4, 10)
	value, _ = cli.Get("4")
	t.Logf("%s", value)
	cli.Incr("5")
	cli.Incr("5")
	value, _ = cli.Get("5")
	t.Logf("%s", value)
	cli.IncrBy("5", 2)
	value, _ = cli.Get("5")
	t.Logf("%s", value)
	cli.Decr("5")
	value, _ = cli.Get("5")
	t.Logf("%s", value)
	cli.DecrBy("5", 2)
	value, _ = cli.Get("5")
	t.Logf("%s", value)

	time.Sleep(time.Duration(2) * time.Second)

	cli.MSet(map[string]interface{}{"1": "1", "2": "2", "3": 3, "6": 4.2, "7": "ssss"})
	values, _ := cli.MGet([]string{"1", "2", "4", "5", "3"})
	t.Logf("%v", values)
	values, _ = cli.MGet([]string{"1", "2", "4", "6", "5", "3"})
	t.Logf("%v", values)
}

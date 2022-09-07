package redisLock

import (
	"github.com/y1015860449/gotoolkit/cache/goredis/redis"
	"log"
	"math/rand"
	"time"
)

var rdCli *redis.GoRedis

func InitRedis(cli *redis.GoRedis) {
	rdCli = cli
}

func KeyLock(key, value string, expiry int, timeout time.Duration) error {
	for {
		if rest, err := rdCli.SetNxEx(key, value, expiry); err != nil {
			log.Printf("keyLock SetNxEx err(%+v)", err)
			return err
		} else {
			if !rest {
				log.Printf("keyLock wait! key(%v) value(%v)", key, value)
				time.Sleep(time.Duration(rand.Intn(10)+5) * time.Millisecond)
				continue
			} else {
				log.Printf("keyLock success! key(%v) value(%v)", key, value)
				return nil
			}
		}
	}
}

func KeyUnLock(key, value string) error {
	script := `
		if redis.call('get',KEYS[1])==ARGV[1]
		then 
			return redis.call('del',KEYS[1])
		else 
			return 0
		end
	`
	rest, err := rdCli.Eval(script, []string{key}, []interface{}{value})
	if err != nil {
		log.Printf("keyUnLock Eval err(%+v)", err)
		return err
	} else {
		if rest.(int64) == 1 {
			log.Printf("keyUnLock success! key(%v) value(%v)", key, value)
		}
	}
	return nil
}

// AddKeyExpiry 续时间
func AddKeyExpiry(key, value string, expiry int) error {
	select {
	case <-time.After(time.Second*time.Duration(expiry) - time.Millisecond*50):
		script := `
			if redis.call('get', KEYS[1]) == ARGV[1]
			then
				return redis.call('setex', KEYS[1], ARGV[2], ARGV[1])
			else
				return 'no'
		`
		rest, err := rdCli.Eval(script, []string{key}, []interface{}{expiry, value})
		if err != nil {
			log.Printf("addKeyExpiry Eval err(%+v)", err)
			return err
		} else {
			if rest.(string) == "OK" {
				log.Printf("addKeyExpiry success")
				go AddKeyExpiry(key, value, expiry)
			}
		}

	}
	return nil
}

package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

type RedisConfig struct {
	Addrs        string
	Password     string
	MaxIdleConns int //最初的连接数量
	MaxOpenConns int //连接池最大连接数量,不确定可以用0(0表示自动定义，按需分配)
	MaxLifeTime  int //连接关闭时间100秒(100秒不使用将关闭)
}

type Redigo struct {
	pool *redis.Pool
}

func InitRedis(conf *RedisConfig) (*Redigo, error) {

	pool := &redis.Pool{
		Dial: func() (conn redis.Conn, err error) {
			c, err := redis.Dial("tcp", conf.Addrs)
			if err != nil {
				return nil, err
			}
			if conf.Password != "" { // 有可能没有密码
				if _, err := c.Do("AUTH", conf.Password); err != nil {
					_ = c.Close()
					return nil, err
				}
			}
			return c, nil
		},
		MaxIdle:     conf.MaxIdleConns,                             //最初的连接数量
		MaxActive:   conf.MaxOpenConns,                             //连接池最大连接数量,不确定可以用0(0表示自动定义，按需分配)
		IdleTimeout: time.Duration(conf.MaxLifeTime) * time.Second, //连接关闭时间100秒(100秒不使用将关闭)
		Wait:        true,                                          //超过最大连接，是报错，还是等待
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				return fmt.Errorf("ping redis error: %s", err)
			}
			return nil
		},
	}
	return &Redigo{pool: pool}, nil
}

func (hy *Redigo) Get(key string) (string, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	return redis.String(conn.Do("GET", key))
}

func (hy *Redigo) Set(key string, value string) error {
	conn := hy.pool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, value)
	return err
}

func (hy *Redigo) SetEx(key string, value string, seconds int) error {
	conn := hy.pool.Get()
	defer conn.Close()
	_, err := conn.Do("SETEX", key, seconds, value)
	return err
}

func (hy *Redigo) SetNx(key string, value string) (bool, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	if res, err := redis.Int(conn.Do("SETNX", key, value)); err != nil {
		return false, err
	} else {
		return res >= 1, err
	}
}

func (hy *Redigo) MGet(keys ...string) ([]string, error) {

	conn := hy.pool.Get()
	defer conn.Close()
	var values []string
	var err error
	if values, err = redis.Strings(conn.Do("MGET", convertSlice(keys)...)); err != nil {
		return nil, err
	}
	if values != nil && len(values) > 0 {
		for _, s := range values {
			if s != "" {
				return values, nil
			}
		}
	}
	return nil, nil
}

func (hy *Redigo) MSet(keyValues map[string]string) error {
	conn := hy.pool.Get()
	defer conn.Close()
	args := redis.Args{}.AddFlat(keyValues)
	_, err := conn.Do("MSET", args...)
	return err
}

func (hy *Redigo) Incr(key string) (int, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("INCR", key))
}

func (hy *Redigo) IncrBy(key string, incr int) (int, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("INCRBY", key, incr))
}

func (hy *Redigo) HGet(key string, field string) (string, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	return redis.String(conn.Do("HGET", key, field))
}

func (hy *Redigo) HGetAll(key string) (map[string]string, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	return redis.StringMap(conn.Do("HGETALL", key))
}

func (hy *Redigo) HSet(key string, field string, value string) error {
	conn := hy.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HSET", key, field, value)
	return err
}

func (hy *Redigo) HMGet(key string, fields ...string) ([]string, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	args := redis.Args{}.Add(key).AddFlat(convertSlice(fields))
	var values []string
	var err error
	if values, err = redis.Strings(conn.Do("HMGET", args...)); err != nil {
		return nil, err
	}
	if values != nil && len(values) > 0 {
		for _, s := range values {
			if s != "" {
				return values, nil
			}
		}
	}
	return nil, nil
}

func (hy *Redigo) HMSet(key string, fieldValues map[string]string) error {
	conn := hy.pool.Get()
	defer conn.Close()
	args := redis.Args{}.Add(key).AddFlat(fieldValues)
	_, err := conn.Do("HMSET", args...)
	return err
}

func (hy *Redigo) SAdd(key string, members ...string) error {
	conn := hy.pool.Get()
	defer conn.Close()
	args := redis.Args{}.Add(key).AddFlat(convertSlice(members))
	_, err := conn.Do("SADD", args...)
	return err
}

func (hy *Redigo) SMembers(key string) ([]string, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	return redis.Strings(conn.Do("SMEMBERS", key))
}

func (hy *Redigo) SRem(key string, members ...string) error {
	conn := hy.pool.Get()
	defer conn.Close()
	args := redis.Args{}.Add(key).AddFlat(convertSlice(members))
	_, err := conn.Do("SREM", args...)
	return err
}

func (hy *Redigo) SIsMember(key string, member string) (bool, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	do, err := redis.Int(conn.Do("SISMEMBER", key, member))
	if err != nil {
		return false, err
	}
	return do == 1, nil
}

func (hy *Redigo) ZAdd(key string, values map[string]int) error {
	conn := hy.pool.Get()
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for k, v := range values {
		args = args.Add(v).Add(k)
	}
	_, err := conn.Do("ZADD", args...)
	return err
}

func (hy *Redigo) ZRange(key string, start, end int) ([]string, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	return redis.Strings(conn.Do("ZRANGE", key, start, end))
}

func (hy *Redigo) ZRangeWithScores(key string, start, end int) (map[string]int, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	return redis.IntMap(conn.Do("ZRANGE", key, start, end, "withscores"))
}

func (hy *Redigo) ZRangeByScore(key string, min, max interface{}) ([]string, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	return redis.Strings(conn.Do("ZRANGEBYSCORE", key, min, max))
}

func (hy *Redigo) ZRangeByScoreWithScores(key string, min, max interface{}) (map[string]int, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	return redis.IntMap(conn.Do("ZRANGEBYSCORE", key, min, max, "withscores"))
}

func (hy *Redigo) ZRem(key string, members ...string) error {
	conn := hy.pool.Get()
	defer conn.Close()
	args := redis.Args{}.Add(key).AddFlat(convertSlice(members))
	_, err := conn.Do("ZREM", args...)
	return err
}

func (hy *Redigo) ZScore(key string, member string) (int, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("ZSCORE", key, member))
}

func (hy *Redigo) Del(keys ...string) error {
	conn := hy.pool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", convertSlice(keys)...)
	return err
}

func (hy *Redigo) Exists(key string) (bool, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	if res, err := redis.Int(conn.Do("EXISTS", key)); err != nil {
		return false, err
	} else {
		return res == 1, err
	}
}

func (hy *Redigo) Expire(key string, seconds int) (bool, error) {
	conn := hy.pool.Get()
	defer conn.Close()
	if res, err := redis.Int(conn.Do("EXPIRE", key, seconds)); err != nil {
		return false, err
	} else {
		return res == 1, err
	}
}

func (hy *Redigo) GetConn() redis.Conn {
	return hy.pool.Get()
}

func convertSlice(keys []string) []interface{} {
	var ks []interface{}
	for _, key := range keys {
		ks = append(ks, key)
	}
	return ks
}

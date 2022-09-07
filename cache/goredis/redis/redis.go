package redis

import (
	"errors"
	"github.com/go-redis/redis"
	"time"
)

type RedisConfig struct {
	Addr         []string // 地址
	Type         string   // node/cluster
	Pwd          string   // 密码
	MaxRetries   int      // 最大尝试次数, 默认3
	MinIdleConns int      // 最初连接数， 默认8
}

type GoRedis struct {
	redCli redis.Cmdable
}

func checkConfig(conf *RedisConfig) {
	if conf.MinIdleConns == 0 {
		conf.MinIdleConns = 8
	}
	if conf.MaxRetries == 0 {
		conf.MaxRetries = 3
	}
	if len(conf.Type) <= 0 {
		conf.Type = "node"
	}
}

func InitRedis(conf *RedisConfig) (*GoRedis, error) {
	checkConfig(conf)
	if len(conf.Addr) <= 0 {
		return nil, errors.New("config addr error")
	}
	if conf.Type == "node" {
		cli := redis.NewClient(&redis.Options{
			Addr:         conf.Addr[0],
			Password:     conf.Pwd,
			DB:           0,
			MaxRetries:   conf.MaxRetries,
			MinIdleConns: conf.MinIdleConns,
		})
		if err := cli.Ping().Err(); err != nil {
			return nil, err
		}
		return &GoRedis{redCli: cli}, nil
	} else if conf.Type == "cluster" {
		cli := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        conf.Addr,
			Password:     conf.Pwd,
			MaxRetries:   conf.MaxRetries,
			MinIdleConns: conf.MinIdleConns,
		})
		if err := cli.Ping().Err(); err != nil {
			return nil, err
		}
		return &GoRedis{redCli: cli}, nil
	} else {
		return nil, errors.New("config type error")
	}
}

func (cli *GoRedis) Exists(keys []string) (int64, error) {
	return cli.redCli.Exists(keys...).Result()
}

func (cli *GoRedis) Del(keys []string) error {
	err := cli.redCli.Del(keys...).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	return nil
}

func (cli *GoRedis) Expire(key string, expiration int) (bool, error) {
	rest, err := cli.redCli.Expire(key, time.Duration(expiration)*time.Second).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	return rest, nil
}

func (cli *GoRedis) Persist(key string) (bool, error) {
	rest, err := cli.redCli.Persist(key).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	return rest, nil
}

func (cli *GoRedis) RenameNX(key, newKey string) (bool, error) {
	rest, err := cli.redCli.RenameNX(key, newKey).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	return rest, nil
}

////////////////////////////////////////
// string
////////////////////////////////////////
func (cli *GoRedis) Incr(key string) (int64, error) {
	return cli.redCli.Incr(key).Result()
}

func (cli *GoRedis) IncrBy(key string, value int64) (int64, error) {
	return cli.redCli.IncrBy(key, value).Result()
}

func (cli *GoRedis) Decr(key string) (int64, error) {
	return cli.redCli.Decr(key).Result()
}

func (cli *GoRedis) DecrBy(key string, value int64) (int64, error) {
	return cli.redCli.DecrBy(key, value).Result()
}

func (cli *GoRedis) Set(key string, value interface{}) error {
	return cli.redCli.Set(key, value, 0).Err()
}

func (cli *GoRedis) SetEx(key string, value interface{}, expiration int) error {
	return cli.redCli.Set(key, value, time.Duration(expiration)*time.Second).Err()
}

func (cli *GoRedis) SetNx(key string, value interface{}) error {
	return cli.redCli.SetNX(key, value, 0).Err()
}

func (cli *GoRedis) SetNxEx(key string, value interface{}, expiration int) (bool, error) {
	return cli.redCli.SetNX(key, value, time.Duration(expiration)*time.Second).Result()
}

func (cli *GoRedis) Get(key string) (string, error) {
	value, err := cli.redCli.Get(key).Result()
	if err != nil && err != redis.Nil {
		return "", err
	}
	return value, nil
}

func (cli *GoRedis) MSet(keyValues map[string]interface{}) error {
	var pairs []interface{}
	for k, v := range keyValues {
		pairs = append(pairs, k, v)
	}
	_, err := cli.redCli.MSet(pairs...).Result()
	return err
}

func (cli *GoRedis) MSetNx(keyValues map[string]interface{}) error {
	var pairs []interface{}
	for k, v := range keyValues {
		pairs = append(pairs, k, v)
	}
	_, err := cli.redCli.MSetNX(pairs...).Result()
	return err
}

func (cli *GoRedis) MGet(keys []string) (map[string]interface{}, error) {
	rest, err := cli.redCli.MGet(keys...).Result()
	if err != nil {
		return nil, err
	}
	values := make(map[string]interface{}, len(keys))
	for i, v := range keys {
		values[v] = rest[i]
	}
	return values, err
}

////////////////////////////////////////
// hash
////////////////////////////////////////
func (cli *GoRedis) HSet(key, field string, value interface{}) error {
	return cli.redCli.HSet(key, field, value).Err()
}

func (cli *GoRedis) HSetNX(key, field string, value interface{}) error {
	return cli.redCli.HSetNX(key, field, value).Err()
}

func (cli *GoRedis) HMSet(key string, fields map[string]interface{}) error {
	return cli.redCli.HMSet(key, fields).Err()
}

func (cli *GoRedis) HGet(key, field string) (string, error) {
	return cli.redCli.HGet(key, field).Result()
}

func (cli *GoRedis) HMGet(key string, fields []string) ([]interface{}, error) {
	return cli.redCli.HMGet(key, fields...).Result()
}

func (cli *GoRedis) HGetAll(key string) (map[string]string, error) {
	return cli.redCli.HGetAll(key).Result()
}

func (cli *GoRedis) HKeys(key string) ([]string, error) {
	return cli.redCli.HKeys(key).Result()
}

func (cli *GoRedis) HVals(key string) ([]string, error) {
	return cli.redCli.HVals(key).Result()
}

func (cli *GoRedis) HDel(key string, fields []string) error {
	return cli.redCli.HDel(key, fields...).Err()
}

func (cli *GoRedis) HExists(key, field string) (bool, error) {
	return cli.redCli.HExists(key, field).Result()
}

func (cli *GoRedis) HIncrBy(key, field string, incr int64) (int64, error) {
	return cli.redCli.HIncrBy(key, field, incr).Result()
}

////////////////////////////////////////
// set
////////////////////////////////////////
func (cli *GoRedis) SAdd(key string, member interface{}) (int64, error) {
	return cli.redCli.SAdd(key, member).Result()
}

func (cli *GoRedis) SAdds(key string, members []interface{}) (int64, error) {
	return cli.redCli.SAdd(key, members...).Result()
}

func (cli *GoRedis) SCard(key string) (int64, error) {
	return cli.redCli.SCard(key).Result()
}

func (cli *GoRedis) SMembers(key string) ([]string, error) {
	return cli.redCli.SMembers(key).Result()
}

func (cli *GoRedis) SIsMember(key string, member interface{}) (bool, error) {
	return cli.redCli.SIsMember(key, member).Result()
}

func (cli *GoRedis) SRem(key string, member interface{}) (int64, error) {
	return cli.redCli.SRem(key, member).Result()
}

func (cli *GoRedis) SRems(key string, members []interface{}) (int64, error) {
	return cli.redCli.SRem(key, members...).Result()
}

func (cli *GoRedis) SInter(keys []string) ([]string, error) {
	return cli.redCli.SInter(keys...).Result()
}

func (cli *GoRedis) SDiff(keys []string) ([]string, error) {
	return cli.redCli.SDiff(keys...).Result()
}

func (cli *GoRedis) SUnion(keys []string) ([]string, error) {
	return cli.redCli.SUnion(keys...).Result()
}

////////////////////////////////////////
// sorted set
////////////////////////////////////////
func (cli *GoRedis) ZAdd(key string, value string, score int64) (int64, error) {
	return cli.redCli.ZAdd(key, redis.Z{Member: value, Score: float64(score)}).Result()
}

func (cli *GoRedis) ZAdds(key string, valueScore map[string]int64) (int64, error) {
	var zScore []redis.Z
	for k, v := range valueScore {
		zScore = append(zScore, redis.Z{Member: k, Score: float64(v)})
	}
	return cli.redCli.ZAdd(key, zScore...).Result()
}

func (cli *GoRedis) ZAddNX(key string, value string, score int64) (int64, error) {
	return cli.redCli.ZAddNX(key, redis.Z{Member: value, Score: float64(score)}).Result()
}

func (cli *GoRedis) ZAddsNX(key string, valueScore map[string]int64) (int64, error) {
	var zScore []redis.Z
	for k, v := range valueScore {
		zScore = append(zScore, redis.Z{Member: k, Score: float64(v)})
	}
	return cli.redCli.ZAddNX(key, zScore...).Result()
}

func (cli *GoRedis) ZRangeWithScores(key string, start, end int64) (map[string]int64, error) {
	rest, err := cli.redCli.ZRangeWithScores(key, start, end).Result()
	if err != nil {
		return nil, err
	}
	valueScore := make(map[string]int64, 0)
	for _, v := range rest {
		valueScore[v.Member.(string)] = int64(v.Score)
	}
	return valueScore, nil
}

func (cli *GoRedis) ZScore(key string, value string) (int64, error) {
	rest, err := cli.redCli.ZScore(key, value).Result()
	if err != nil {
		return 0, err
	}
	return int64(rest), nil
}

////////////////////////////////////////
// bitmap
////////////////////////////////////////
func (cli *GoRedis) SetBit(key string, offset int64, value int) (int64, error) {
	return cli.redCli.SetBit(key, offset, value).Result()
}

func (cli *GoRedis) GetBit(key string, offset int64) (int64, error) {
	return cli.redCli.GetBit(key, offset).Result()
}

func (cli *GoRedis) BitCount(key string, start, end int64) (int64, error) {

	return cli.redCli.BitCount(key, &redis.BitCount{
		Start: start,
		End:   end,
	}).Result()
}

////////////////////////////////////////
// 其他高级属性
////////////////////////////////////////
func (cli *GoRedis) Pipeline() redis.Pipeliner {
	return cli.redCli.Pipeline()
}

func (cli *GoRedis) Pipelined(fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return cli.redCli.Pipelined(fn)
}

func (cli *GoRedis) Eval(script string, keys []string, args []interface{}) (interface{}, error) {
	return cli.redCli.Eval(script, keys, args...).Result()
}

func (cli *GoRedis) EvalSha(script string, keys []string, args []interface{}) (interface{}, error) {
	return cli.redCli.EvalSha(script, keys, args...).Result()
}

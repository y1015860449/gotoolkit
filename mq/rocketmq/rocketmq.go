package rocketmq

import (
	"github.com/y1015860449/gotoolkit/plugin/mq/broker"
	"github.com/y1015860449/gotoolkit/plugin/mq/rocketmq"
)

type RocketMqConf struct {
	Addrs             []string
	Retry             int
	MaxReconsumeTimes int32
	FromWhere         rocketmq.ConsumeFromWhere // mq的开始未知 0 最新位置  1 最早位置
	ConsumerMode      rocketmq.MessageModel     // 消费者模式 0 BroadCasting 1 Clustering
}

func DefaultRocketMqConfig() *RocketMqConf {
	return &RocketMqConf{
		Addrs:             []string{"127.0.0.1:6789"},
		Retry:             3,
		MaxReconsumeTimes: 5,
		FromWhere:         rocketmq.ConsumeFromLastOffset,
		ConsumerMode:      rocketmq.Clustering,
	}
}

func NewRocketMqBroker(conf *RocketMqConf) (broker.Broker, error) {
	opts := make([]broker.Option, 0)
	if len(conf.Addrs) > 0 {
		opts = append(opts, broker.Addrs(conf.Addrs...))
	}
	opts = append(opts, rocketmq.WithRetry(conf.Retry))
	if conf.MaxReconsumeTimes > 0 {
		opts = append(opts, rocketmq.WithMaxReconsumeTimes(conf.MaxReconsumeTimes))
	}
	opts = append(opts, rocketmq.WithConsumeFromWhere(conf.FromWhere))
	opts = append(opts, rocketmq.WithConsumeMode(conf.ConsumerMode))
	b := rocketmq.NewBroker(opts...)
	if err := b.Connect(); err != nil {
		return nil, err
	}
	return b, nil
}

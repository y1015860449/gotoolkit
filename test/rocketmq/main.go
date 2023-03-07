package main

import (
	"github.com/y1015860449/gotoolkit/mq/rocketmq"
	"github.com/y1015860449/gotoolkit/plugin/mq/broker"
	rmq "github.com/y1015860449/gotoolkit/plugin/mq/rocketmq"
	"log"
)

func main() {
	var (
		bk  broker.Broker
		err error
	)

	// 构建broker
	if bk, err = rocketmq.NewRocketMqBroker(&rocketmq.RocketMqConf{
		Addrs:        []string{"127.0.0.1:6789"},
		FromWhere:    rmq.ConsumeFromLastOffset,
		ConsumerMode: rmq.BroadCasting,
	}); err != nil {
		log.Fatalf("init rocketmq broker err(%v)", err)
	}
	// 订阅
	if _, err = bk.Subscribe("testTopic", func(event broker.Event) error {
		// todo 消息内容
		log.Printf("head(%+v) body(%+v)", event.Message().Header, event.Message().Body)
		return nil
	}); err != nil {
		log.Fatalf("register subscriber err(%v)", err)
	}

	// 发布
	if err = bk.Publish("testTopic", &broker.Message{Header: nil, Body: []byte("test")}); err != nil {
		log.Fatalf("publish err(%v)", err)
	}
}

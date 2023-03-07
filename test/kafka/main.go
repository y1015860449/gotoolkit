package main

import (
	"github.com/y1015860449/gotoolkit/mq/kafka"
	"github.com/y1015860449/gotoolkit/plugin/mq/broker"
	"log"
)

func main() {
	var (
		bk  broker.Broker
		err error
	)
	// 构建broker
	if bk, err = kafka.NewKafkaBroker([]string{"127.0.0.1:9092"}, kafka.PttHash, kafka.CnmNewest); err != nil {
		log.Fatalf("init kafka broker err(%v)", err)
	}
	// 订阅
	if _, err = bk.Subscribe("testTopic", func(event broker.Event) error {
		// todo 消息内容
		log.Printf("head(%+v) body(%+v)", event.Message().Header, event.Message().Body)
		return nil
	}, func(opt *broker.SubscribeOptions) {
		log.Print("broker connect success")
		opt.Queue = "testGroup" // 使用相同的groupID确保消息不重复消费
	}); err != nil {
		log.Fatalf("register subscriber err(%v)", err)
	}

	// 发布
	if err = bk.Publish("testTopic", &broker.Message{Header: nil, Body: []byte("test")}); err != nil {
		log.Fatalf("publish err(%v)", err)
	}
}

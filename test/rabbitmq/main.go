package main

import (
	"fmt"
	"github.com/y1015860449/gotoolkit/log/zaplog"
	"github.com/y1015860449/gotoolkit/mq/rabbitmq"
	"github.com/y1015860449/gotoolkit/plugin/mq/broker"
	"time"
)

type TestCus1 struct {
}

func (e *TestCus1) Handler(event broker.Event) error {
	defer event.Ack()
	msg := event.Message()
	zaplog.ZapLog.Infof("recv1111 %s", string(msg.Body))
	return nil
}

type TestCus2 struct {
}

func (e *TestCus2) Handler(event broker.Event) error {
	defer event.Ack()
	msg := event.Message()
	zaplog.ZapLog.Infof("recv2222 %s", string(msg.Body))
	return nil
}

const (
	url   = "amqp://admin:admin@192.168.20.99:5672"
	topic = "test.topic"
)

func main() {
	zaplog.InitLogger(nil)
	bk, err := rabbitmq.NewRabbitMqBroker([]string{url}, "test1")
	if err != nil {
		zaplog.ZapLog.Panicf("new broker err(%+v)", err)
	}
	defer bk.Disconnect()
	h := &TestCus1{}
	opts := rabbitmq.SubscribeOption(false, nil, nil)
	opts = append(opts, func(options *broker.SubscribeOptions) {
		options.AutoAck = false
		options.Queue = "test.rabbit"
	})
	sub1, err := bk.Subscribe(topic, h.Handler, opts...)
	if err != nil {
		zaplog.ZapLog.Panicf("subscriber err(%+v)", err)
	}
	defer sub1.Unsubscribe()

	h2 := &TestCus2{}
	sub2, err := bk.Subscribe(topic, h2.Handler, opts...)
	if err != nil {
		zaplog.ZapLog.Panicf("subscriber err(%+v)", err)
	}
	defer sub2.Unsubscribe()

	for i := 0; i < 10000; i++ {
		bk.Publish(topic, &broker.Message{
			Header: nil,
			Body:   []byte(fmt.Sprintf("test %d", i)),
		}, rabbitmq.PublishOption(2, "text/plain")...)
		time.Sleep(500 * time.Millisecond)
	}
}

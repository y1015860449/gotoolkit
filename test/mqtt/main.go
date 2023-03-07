package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/y1015860449/gotoolkit/log/zaplog"
	"github.com/y1015860449/gotoolkit/mq/hxmqtt"
	"time"
)

type testMqttData struct {
	Topic   string `json:"topic"`
	Produce string `json:"produce"`
	Msg     string `json:"msg"`
	Time    int64  `json:"time"`
}

const (
	topicName  = "test-mqtt-topic"
	topicName2 = "test2-mqtt2-topic2"
	topicName3 = "test3-mqtt3-topic3"
	testMsg    = "test mqtt tick %d"
)

func Subscriber1(cli mqtt.Client, msg mqtt.Message) {
	var rcvData testMqttData
	_ = json.Unmarshal(msg.Payload(), &rcvData)
	id := msg.MessageID()
	topic := msg.Topic()
	qos := msg.Qos()

	zaplog.ZapLog.Infof("Subscriber111111111 data(%+v) id(%v) topic(%v) qos(%v)", &rcvData, id, topic, qos)
}

func Subscriber2(cli mqtt.Client, msg mqtt.Message) {
	var rcvData testMqttData
	_ = json.Unmarshal(msg.Payload(), &rcvData)
	id := msg.MessageID()
	topic := msg.Topic()
	qos := msg.Qos()

	zaplog.ZapLog.Infof("Subscriber22222222 data(%+v) id(%v) topic(%v) qos(%v)", &rcvData, id, topic, qos)
}

func SubscribeMultiple(cli mqtt.Client, msg mqtt.Message) {
	var rcvData testMqttData
	_ = json.Unmarshal(msg.Payload(), &rcvData)
	id := msg.MessageID()
	topic := msg.Topic()
	qos := msg.Qos()

	zaplog.ZapLog.Infof("SubscribeMultiple data(%+v) id(%v) topic(%v) qos(%v)", &rcvData, id, topic, qos)
}

func NewMqttClient() *hxmqtt.MqttClient {
	c := hxmqtt.DefaultMqttConf()
	c.Uri = "tcp://192.168.20.99:1883"
	c.Username = "admin"
	c.Password = "admin123"
	mqttCli, err := hxmqtt.NewMqtt(c)
	if err != nil {
		zaplog.Panicf("new mqtt err(%+v)", err)
	}
	zaplog.ZapLog.Infof("success client %s", c.ClientId)
	return mqttCli
}

func main() {
	zaplog.InitLogger(nil)
	mqttCli := NewMqttClient()
	mqttCli1 := NewMqttClient()
	mqttCli2 := NewMqttClient()
	go func() {
		_ = mqttCli1.Subscribe(topicName, 1, Subscriber1)
	}()

	go func() {
		_ = mqttCli2.Subscribe(topicName2, 1, Subscriber2)
	}()

	go func() {
		_ = mqttCli.SubscribeMultiple(map[string]byte{topicName: 1, topicName2: 1, topicName3: 1}, SubscribeMultiple)
	}()

	go func() {
		tick := int64(0)
		for {
			msg := fmt.Sprintf(testMsg, tick)
			tick++
			testData := &testMqttData{
				Topic:   topicName,
				Produce: "client11",
				Msg:     msg,
				Time:    time.Now().UnixMilli(),
			}
			data, _ := json.Marshal(testData)
			_ = mqttCli.Publish(topicName, 0, data)
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	go func() {
		tick := int64(0)
		for {
			msg := fmt.Sprintf(testMsg, tick)
			tick++
			testData := &testMqttData{
				Topic:   topicName2,
				Produce: "client22",
				Msg:     msg,
				Time:    time.Now().UnixMilli(),
			}
			data, _ := json.Marshal(testData)
			_ = mqttCli1.Publish(topicName2, 0, data)
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	tick := int64(0)
	for {
		msg := fmt.Sprintf(testMsg, tick)
		tick++
		testData := &testMqttData{
			Topic:   topicName3,
			Produce: "client33",
			Msg:     msg,
			Time:    time.Now().UnixMilli(),
		}
		data, _ := json.Marshal(testData)
		_, _ = mqttCli2.PublishTimeout(topicName3, 0, data, 3*time.Second)
		time.Sleep(200 * time.Millisecond)
	}
}

package hxmqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/y1015860449/gotoolkit/log/zaplog"
	"github.com/y1015860449/gotoolkit/utils"
	"time"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	zaplog.ZapLog.Infof("Received message: %s from topic: %s", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	zaplog.ZapLog.Infof("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	zaplog.ZapLog.Infof("Connect lost: %+v", err)
}

type MqttConf struct {
	Uri      string // 格式"tcp://[ip]:[port]"
	Username string // 用户名
	Password string // 密码

	ConnectedFunc      mqtt.OnConnectHandler      // 连接成功回调函数
	LostConnFunc       mqtt.ConnectionLostHandler // 断开连接回调函数
	DefaultPublishFunc mqtt.MessageHandler        // 默认订阅回调函数
	ClientId           string                     // 客户端id（客户端唯一标识）
	AutoAckDisabled    bool                       // 禁用自动ack
	AutoReconnect      bool
	ConnectTimeout     time.Duration
}

func DefaultMqttConf() *MqttConf {
	return &MqttConf{
		Uri:                "tcp://127.0.0.1:1883",
		ClientId:           utils.GetUUID()[:22],
		DefaultPublishFunc: messagePubHandler,
		ConnectedFunc:      connectHandler,
		LostConnFunc:       connectLostHandler,
		AutoAckDisabled:    false,
		AutoReconnect:      true,
		ConnectTimeout:     5 * time.Second,
	}
}

type MqttClient struct {
	cli mqtt.Client
	c   *MqttConf
}

func NewMqtt(config *MqttConf) (*MqttClient, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Uri)
	opts.SetClientID(config.ClientId)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	opts.SetDefaultPublishHandler(config.DefaultPublishFunc)
	opts.SetOnConnectHandler(config.ConnectedFunc)
	opts.SetConnectionLostHandler(config.LostConnFunc)
	opts.SetAutoAckDisabled(config.AutoAckDisabled)
	opts.SetAutoReconnect(config.AutoReconnect)
	opts.SetConnectTimeout(config.ConnectTimeout)
	cli := mqtt.NewClient(opts)
	if token := cli.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return &MqttClient{cli: cli, c: config}, nil
}

// 订阅单个主题
func (p *MqttClient) Subscribe(topic string, qos byte, subCallback mqtt.MessageHandler) error {
	if token := p.cli.Subscribe(topic, qos, subCallback); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// 订阅多个主题，同一个回调处理函数
func (p *MqttClient) SubscribeMultiple(filters map[string]byte, subCallback mqtt.MessageHandler) error {
	if token := p.cli.SubscribeMultiple(filters, subCallback); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (p *MqttClient) Unsubscribe(topics []string) error {
	if token := p.cli.Unsubscribe(topics...); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (p *MqttClient) Publish(topic string, qos byte, data []byte) error {
	if token := p.cli.Publish(topic, qos, false, data); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (p *MqttClient) PublishTimeout(topic string, qos byte, data []byte, timeout time.Duration) (bool, error) {
	token := p.cli.Publish(topic, qos, false, data)
	success := token.WaitTimeout(timeout)
	if token.Error() != nil {
		return success, token.Error()
	}
	return success, nil
}

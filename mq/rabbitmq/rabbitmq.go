package rabbitmq

import (
	"github.com/y1015860449/gotoolkit/plugin/mq/broker"
	"github.com/y1015860449/gotoolkit/plugin/mq/rabbitmq"
)

func NewRabbitMqBroker(addrs []string, exchangeName string) (broker.Broker, error) {
	var options []broker.Option
	options = append(options, broker.Addrs(addrs...), rabbitmq.DurableExchange(), rabbitmq.ExchangeName(exchangeName))
	bk := rabbitmq.NewBroker(options...)
	if err := bk.Connect(); err != nil {
		return nil, err
	}
	return bk, nil
}

func PublishOption(deliveryMode uint8, contentType string) []broker.PublishOption {
	var options []broker.PublishOption
	options = append(options, rabbitmq.DeliveryMode(deliveryMode), rabbitmq.ContentType(contentType))
	return options
}

func SubscribeOption(autoAck bool, headers, args map[string]interface{}) []broker.SubscribeOption {
	var options []broker.SubscribeOption
	options = append(options, rabbitmq.DurableQueue())
	if autoAck {
		options = append(options, rabbitmq.AckOnSuccess())
	}
	if len(headers) > 0 {
		options = append(options, rabbitmq.Headers(headers))
	}
	if len(args) > 0 {
		options = append(options, rabbitmq.QueueArguments(args))
	}
	return options
}

package kafka

import (
	"errors"
	"github.com/Shopify/sarama"
	"github.com/y1015860449/gotoolkit/plugin/mq/broker"
	"github.com/y1015860449/gotoolkit/plugin/mq/kafka"
	"time"
)

const connectTimeout = time.Second * 3

type PartitionerMode int8

const (
	PttHash       PartitionerMode = 1 // hash
	PttManual     PartitionerMode = 2 // 手工
	PttRandom     PartitionerMode = 3 // 随机
	PttRoundRobin PartitionerMode = 4 // 轮询调度
)

type ConsumerMode int8

const (
	CnmNewest ConsumerMode = 1 // 最新
	CnmOldest ConsumerMode = 2 // 最老
)

func NewKafkaBroker(addrs []string, pttMode PartitionerMode, cnmMode ConsumerMode) (broker.Broker, error) {
	var options []broker.Option
	options = append(options, broker.Addrs(addrs...), producerConfig(pttMode), consumerConfig(cnmMode))
	bk := kafka.NewBroker(options...)

	errCh := make(chan error, 1)
	defer close(errCh)

	go func() {
		errCh <- bk.Connect()
	}()

	select {
	case <-time.After(connectTimeout):
	case e, ok := <-errCh:
		if ok {
			return bk, e
		}
	}
	return nil, errors.New("connect timeout")
}

func defaultProducerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = sarama.V2_0_0_0
	return config
}

func producerConfig(pttMode PartitionerMode) broker.Option {
	config := defaultProducerConfig()
	switch pttMode {
	case PttHash:
		config.Producer.Partitioner = sarama.NewHashPartitioner
	case PttManual:
		config.Producer.Partitioner = sarama.NewManualPartitioner
	case PttRandom:
		config.Producer.Partitioner = sarama.NewRandomPartitioner
	default:
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	}
	return kafka.BrokerConfig(config)
}

func defaultConsumerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = true
	return config
}

func consumerConfig(cnmMode ConsumerMode) broker.Option {
	config := defaultConsumerConfig()
	if cnmMode == CnmOldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	} else {
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	}
	return kafka.ClusterConfig(config)
}

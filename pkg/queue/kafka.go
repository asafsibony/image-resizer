package queue

import (
	"time"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type Producer struct {
	Producer *kafka.Producer
}

type Consumer struct {
	Consumer *kafka.Consumer
}

func NewKafkaConsumer(kafkaBootstrapServers string, kafkaConsumerGroupID string) (*Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  kafkaBootstrapServers,
		"group.id":           kafkaConsumerGroupID,
		"auto.offset.reset":  "latest",
		"enable.auto.commit": "false",
	})
	if err != nil {
		return nil, err
	}
	return &Consumer{Consumer: consumer}, nil
}

func NewKafkaProducer(kafkaBootstrapServers string) (*Producer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaBootstrapServers})
	if err != nil {
		return nil, err
	}
	return &Producer{Producer: producer}, nil
}

func (p *Producer) ProduceMessage(topic string, key string, message []byte) error {
	return p.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
		Timestamp:      time.Now(),
		TimestampType:  kafka.TimestampCreateTime,
		Key:            []byte(key),
	}, nil)
}

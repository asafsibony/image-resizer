package queue

import (
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

func NewKafkaConsumer(kafkaBootstrapServers string, kafkaConsumerGroupID string) (*kafka.Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  kafkaBootstrapServers,
		"group.id":           kafkaConsumerGroupID,
		"auto.offset.reset":  "latest",
		"enable.auto.commit": "false",
	})
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

func NewKafkaProducer(kafkaBootstrapServers string) (*kafka.Producer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaBootstrapServers})
	if err != nil {
		return nil, err
	}
	return producer, nil
}

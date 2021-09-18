package main

import (
	"context"
	"os"

	"github.com/asafsibony/image-resizer/cmd/image-resizer-ms/resizer"
	"github.com/asafsibony/image-resizer/pkg/cache"
	"github.com/asafsibony/image-resizer/pkg/persistency"
	"github.com/asafsibony/image-resizer/pkg/queue"
	"github.com/sirupsen/logrus"
)

func main() {
	// TODO: replace by env vars
	LOG_LEVEL := "debug"
	REDIS_HOST := "127.0.0.1:6379"
	REDIS_PASSWORD := ""
	REDIS_DB := 0
	KAFKA_SERVERS := "127.0.0.1:29092"
	KAFKA_IMAGE_RESIZE_TOPIC := "image-resize"
	KAFKA_CONSUMER_GROUP := "image-resizer-ms"
	POSTGRES_HOST := "127.0.0.1"
	POSTGRES_PORT := "5432"
	POSTGRES_DATABASE := "imageresizer"
	POSTGRES_USER := "postgres"
	POSTGRES_PASSWORD := "\"\""
	POSTGRES_OPTIONS := ""

	// initialize the logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	logLevel, err := logrus.ParseLevel(LOG_LEVEL)
	if err != nil {
		logLevel = logrus.WarnLevel
		logger.Warn("Faild to parse the log level, setting the log level to WARN.")
	}
	logger.SetLevel(logLevel)

	logger.Debug("Intializing the Kafka consumer...")
	kafkaConsumer, err := queue.NewKafkaConsumer(KAFKA_SERVERS, KAFKA_CONSUMER_GROUP)
	defer kafkaConsumer.Consumer.Close()
	if err != nil {
		logger.Error(err)
	}

	logger.Debug("Intializing the Redis client")
	redisClient := cache.NewRedisClient(context.Background(), logger, REDIS_HOST, REDIS_PASSWORD, REDIS_DB)

	logger.Debug("Intializing the Postgres client...")
	connectionInfo := &persistency.ConnectionInfo{
		Host:     POSTGRES_HOST,
		Port:     POSTGRES_PORT,
		Database: POSTGRES_DATABASE,
		User:     POSTGRES_USER,
		Password: POSTGRES_PASSWORD,
		Options:  POSTGRES_OPTIONS,
	}
	psqlClient, err := persistency.NewClient(logger, connectionInfo, false)
	if err != nil {
		logger.Error(err.Error())
	}
	err = psqlClient.Start()
	if err != nil {
		logger.Error(err.Error())
	}

	logger.Debug("Services Initialize done. starting the app.")
	resizer := resizer.NewResizer(logger, redisClient, psqlClient, kafkaConsumer, KAFKA_IMAGE_RESIZE_TOPIC)
	resizer.Start()
}

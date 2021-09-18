package main

import (
	"context"
	"os"

	"github.com/asafsibony/image-resizer/cmd/api-gateway/router"
	"github.com/asafsibony/image-resizer/pkg/cache"
	httpServer "github.com/asafsibony/image-resizer/pkg/http"
	"github.com/asafsibony/image-resizer/pkg/persistency"
	"github.com/asafsibony/image-resizer/pkg/queue"
	"github.com/sirupsen/logrus"
)

// type Config struct {
// 	httpServer *httpServer.Server
// 	logger *logrus.Logger
// }

func main() {
	LOG_LEVEL := "debug"
	HTTP_SERVER_PORT := "8080"
	REDIS_HOST := "127.0.0.1:6379"
	REDIS_PASSWORD := ""
	REDIS_DB := 0

	KAFKA_SERVERS := "127.0.0.1:29092"
	KAFKA_IMAGE_RESIZE_TOPIC := "image_resize"
	POSTGRES_HOST := "127.0.0.1"
	POSTGRES_PORT := "5432"
	POSTGRES_DATABASE := "imageresizer"
	POSTGRES_USER := "postgres"
	POSTGRES_PASSWORD := "\"\""
	POSTGRES_OPTIONS := ""

	logger := logrus.New()
	// init logger
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	logLevel, err := logrus.ParseLevel(LOG_LEVEL)
	if err != nil {
		logLevel = logrus.WarnLevel
		logger.Warn("Faild to parse the log level, setting the log level to WARN.")
	}
	logger.SetLevel(logLevel)

	logger.Debug("Image resizer application started.")

	// kafkaConsumer, err := queue.NewKafkaConsumer(KAFKA_SERVERS, CONSUMER_GROUP)
	// defer kafkaConsumer.Close()
	// if err != nil {
	// 	logger.Error(err)
	// }

	KafkaProducer, err := queue.NewKafkaProducer(KAFKA_SERVERS)
	defer KafkaProducer.Producer.Close()
	if err != nil {
		logger.Error(err.Error())
	}

	redisClient := cache.NewRedisClient(context.Background(), logger, REDIS_HOST, REDIS_PASSWORD, REDIS_DB)

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

	httpServer := httpServer.NewServer(logger, HTTP_SERVER_PORT)
	router := router.NewRouter(httpServer, logger, redisClient, psqlClient, KafkaProducer, KAFKA_IMAGE_RESIZE_TOPIC)
	router.InitRoutes()

	httpServer.Start()
}

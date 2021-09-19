package main

import (
	"context"
	"fmt"
	"os"

	"github.com/asafsibony/image-resizer/cmd/api-gateway/router"
	"github.com/asafsibony/image-resizer/pkg/cache"
	httpServer "github.com/asafsibony/image-resizer/pkg/http"
	"github.com/asafsibony/image-resizer/pkg/persistency"
	"github.com/asafsibony/image-resizer/pkg/queue"
	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"
)

type Config struct {
	LogLevel              string `env:"APP_LOG_LEVEL" envDefault:"debug"`
	HttpServerPort        string `env:"HTTP_SERVER_PORT" envDefault:"8080"`
	RedisHost             string `env:"REDIS_HOST" envDefault:"127.0.0.1:6379"`
	RedisPassword         string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB               int    `env:"REDIS_DB" envDefault:"0"`
	KafkaServers          string `env:"KAFKA_SERVERS" envDefault:"127.0.0.1:29092"`
	KafkaImageResizeTopic string `env:"KAFKA_IMAGE_RESIZE_TOPIC" envDefault:"image-resize"`
	PostgresHost          string `env:"POSTGRES_HOST" envDefault:"127.0.0.1"`
	PostgresPort          string `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresDB            string `env:"POSTGRES_DATABASE" envDefault:"imageresizer"`
	PostgresUser          string `env:"POSTGRES_USER" envDefault:"postgres"`
	PostgresPassword      string `env:"POSTGRES_PASSWORD" envDefault:""`
	PostgresOptions       string `env:"POSTGRES_OPTIONS" envDefault:""`
}

func main() {
	c := &Config{}
	if err := env.Parse(c); err != nil {
		fmt.Println(err.Error())
	}

	// init logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	logLevel, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		logLevel = logrus.WarnLevel
		logger.Warn("Faild to parse the log level, setting the log level to WARN.")
	}
	logger.SetLevel(logLevel)

	logger.Debug("Intializing the Kafka producer...")
	KafkaProducer, err := queue.NewKafkaProducer(c.KafkaServers)
	defer KafkaProducer.Producer.Close()
	if err != nil {
		logger.Error(err.Error())
	}

	logger.Debug("Intializing the Redis client")
	redisClient := cache.NewRedisClient(context.Background(), logger, c.RedisHost, c.RedisPassword, c.RedisDB)

	logger.Debug("Intializing the Postgres client...")
	connectionInfo := &persistency.ConnectionInfo{
		Host:     c.PostgresHost,
		Port:     c.PostgresPort,
		Database: c.PostgresDB,
		User:     c.PostgresUser,
		Password: c.PostgresPassword,
		Options:  c.PostgresOptions,
	}
	psqlClient, err := persistency.NewClient(logger, connectionInfo, false)
	if err != nil {
		logger.Error(err.Error())
	}
	err = psqlClient.Start()
	if err != nil {
		logger.Error(err.Error())
	}

	logger.Debug("Intializing the HTTP...")
	httpServer := httpServer.NewServer(logger, c.HttpServerPort)
	router := router.NewRouter(httpServer, logger, redisClient, psqlClient, KafkaProducer, c.KafkaImageResizeTopic)
	router.InitRoutes()

	logger.Debug("Services Initialize done. starting the app.")
	httpServer.Start()
}

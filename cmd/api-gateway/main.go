package main

import (
	"os"

	"github.com/asafsibony/image-resizer/pkg/api"
	"github.com/sirupsen/logrus"
)

func main() {
	LOG_LEVEL := "debug"
	log := logrus.New()
	// init logger
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)

	logLevel, err := logrus.ParseLevel(LOG_LEVEL)
	if err != nil {
		logLevel = logrus.WarnLevel
		log.Warn("Faild to parse the log level, setting the log level to WARN.")
	}
	log.SetLevel(logLevel)

	log.Debug("Image resizer application started.")

	httpServer, err := api.NewServer(log)
	if err != nil {
		//TODO: handle
	}
	httpServer.Start()
}

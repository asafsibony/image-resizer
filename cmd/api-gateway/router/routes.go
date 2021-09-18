package router

import (
	"github.com/asafsibony/image-resizer/pkg/cache"
	httpServer "github.com/asafsibony/image-resizer/pkg/http"
	"github.com/asafsibony/image-resizer/pkg/persistency"
	"github.com/asafsibony/image-resizer/pkg/queue"
	"github.com/sirupsen/logrus"
)

type Router struct {
	httpServer       *httpServer.Server
	logger           *logrus.Logger
	redisClient      *cache.RedisClient
	psqlClient       *persistency.Client
	kafkaProducer    *queue.Producer
	imageResizeTopic string
}

func NewRouter(httpServer *httpServer.Server,
	logger *logrus.Logger,
	redisClient *cache.RedisClient,
	psqlClient *persistency.Client,
	kafkaProducer *queue.Producer,
	imageResizeTopic string) *Router {
	return &Router{
		httpServer:       httpServer,
		logger:           logger,
		redisClient:      redisClient,
		psqlClient:       psqlClient,
		kafkaProducer:    kafkaProducer,
		imageResizeTopic: imageResizeTopic,
	}
}

// Init the HTTP routes
func (r *Router) InitRoutes() {
	r.httpServer.Router.HandleFunc("/upload", r.uploadImage).Methods("POST")
	r.httpServer.Router.HandleFunc("/status/{uuid}", r.getStatus).Methods("GET")
	r.httpServer.Router.HandleFunc("/download/{uuid}", r.downloadImage).Methods("GET")
}

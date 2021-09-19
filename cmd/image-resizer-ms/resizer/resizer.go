package resizer

import (
	"encoding/json"

	"github.com/asafsibony/image-resizer/pkg/cache"
	"github.com/asafsibony/image-resizer/pkg/persistency"
	"github.com/asafsibony/image-resizer/pkg/queue"
	"github.com/asafsibony/image-resizer/pkg/resources"
	"github.com/asafsibony/image-resizer/pkg/utils"
	"github.com/sirupsen/logrus"
)

type Resizer struct {
	logger           *logrus.Logger
	redisClient      *cache.RedisClient
	psqlClient       *persistency.Client
	kafkaConsumer    *queue.Consumer
	imageResizeTopic string
}

func NewResizer(logger *logrus.Logger,
	redisClient *cache.RedisClient,
	psqlClient *persistency.Client,
	kafkaConsumer *queue.Consumer,
	imageResizeTopic string) *Resizer {
	return &Resizer{
		logger:           logger,
		redisClient:      redisClient,
		psqlClient:       psqlClient,
		kafkaConsumer:    kafkaConsumer,
		imageResizeTopic: imageResizeTopic,
	}
}

// Consumes the requested images for resizing, and perform it.
func (r *Resizer) Start() {
	c := r.kafkaConsumer.Consumer
	c.SubscribeTopics([]string{r.imageResizeTopic}, nil)

	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			r.logger.Debug("Recived imaged resize request: ", string(msg.Key))
			image := &resources.Image{}
			err := json.Unmarshal(msg.Value, image)
			if err != nil {
				r.logger.Error(err)
				continue
			}

			// Resize the image:
			r.resizeImage(image)
		} else {
			// The client will automatically try to recover from all errors.
			r.logger.Error("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

func (r *Resizer) resizeImage(image *resources.Image) {
	defer func() {
		if re := recover(); re != nil {
			err := re.(error)
			r.logger.Error(err.Error())
		}
	}()

	resizedImage, err := utils.ResizeImage(image.Image, uint(image.Dimensions.Width), uint(image.Dimensions.Height))
	status := ""
	if err != nil {
		r.logger.Error("Failed to resize the image: ", image.UUID.String(), " Error: ", err.Error())
		status = resources.Failed
	} else {
		r.logger.Debug("Image resize finished succesfully. image UUID: ", image.UUID.String())
		status = resources.Done
	}

	// Updating Cache and DB with the resize results:
	r.updateResultInCache(image.UUID.String(), status, resizedImage, image.Name)
	r.updateResultInDB(image.UUID, status, resizedImage, image.Name)
}

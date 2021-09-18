package router

import (
	"encoding/json"
	"time"

	"github.com/asafsibony/image-resizer/pkg/resources"
	"github.com/google/uuid"
)

// Save resize request to cache
func (r *Router) saveRequestToCache(imageUuidStr string) {
	r.redisClient.Set(imageUuidStr+":status", resources.Processing)
}

// Save resize request to DB
func (r *Router) saveRequestToDB(imageUuid uuid.UUID) error {
	result := r.psqlClient.Database.Table("requests").Create(&resources.Request{
		ImageUUID: imageUuid,
		Status:    resources.Processing,
		CreatedAt: time.Now()})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Send the resize request to Kafka queue
func (r *Router) sendRequestToQueue(imageUuid uuid.UUID, imageBytes []byte, targetDimensions *resources.TargetDimensions) error {

	image := &resources.Image{
		UUID:       imageUuid,
		Image:      imageBytes,
		Dimensions: targetDimensions,
	}
	imageJson, err := json.Marshal(image)
	if err != nil {
		return err
	}

	err = r.kafkaProducer.ProduceMessage(r.imageResizeTopic, imageUuid.String(), imageJson)
	if err != nil {
		return err
	}

	return nil
}

// Get Resize request status: processing/done/failed (Trying first from the cache and then from postgres as a fallback).
func (r *Router) getRequestStatus(imageUuidStr string) (string, error) {
	// get request status from cache
	status, err := r.redisClient.Get(imageUuidStr + ":status")

	if err == nil && status != "" {
		return status, nil
	}

	// if not exist in cache get from DB
	imageUuid, err := uuid.Parse(imageUuidStr)
	if err != nil {
		r.logger.Error(err)
		return "", err
	}

	request := &resources.Request{}
	result := r.psqlClient.Database.Table("requests").Where("image_uuid = ?", imageUuid).Last(request)
	if result.Error != nil {
		r.logger.Error(err)
		return "", err
	}

	// Update the cache with the fetch status from the DB
	r.redisClient.Set(imageUuidStr+":status", request.Status)

	return request.Status, nil
}

// Get Resized image (Trying first from the cache and then from postgres as a fallback).
func (r *Router) getResizedImage(imageUuidStr string) ([]byte, error) {
	// get resized image from cache
	resizedImage, err := r.redisClient.Get(imageUuidStr + ":resized_image")
	if err == nil && resizedImage != "" {
		return []byte(resizedImage), nil
	}

	// if not exist in cache get from DB
	imageUuid, err := uuid.Parse(imageUuidStr)
	if err != nil {
		r.logger.Error(err)
		return []byte{}, err
	}

	resized_image := &resources.Image{}
	result := r.psqlClient.Database.Table("images").Last(resized_image, "uuid = ?", imageUuid)
	if result.Error != nil {
		r.logger.Error(err)
		return []byte{}, err
	}

	// Update the cache
	r.redisClient.Set(imageUuidStr+":resized_image", resized_image.Image)

	return resized_image.Image, nil
}

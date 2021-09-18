package resizer

import (
	"time"

	"github.com/asafsibony/image-resizer/pkg/resources"
	"github.com/google/uuid"
)

// Update the result in Cache (Redis)
func (r *Resizer) updateResultInCache(imageUuid string, status string, resizedImage []byte) error {
	_, err := r.redisClient.Set(imageUuid+":status", status)
	if err != nil {
		return err
	}

	if status == resources.Done {
		_, err := r.redisClient.Set(imageUuid+":resized_image", resizedImage)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update the result in persistency (Postgres)
func (r *Resizer) updateResultInDB(imageUuid uuid.UUID, status string, resizedImage []byte) error {
	result := r.psqlClient.Database.Table("requests").Create(&resources.Request{ImageUUID: imageUuid, CreatedAt: time.Now(), Status: status})
	if result.Error != nil {
		return result.Error
	}

	if status == resources.Done {
		result := r.psqlClient.Database.Table("images").Create(&resources.Image{UUID: imageUuid, CreatedAt: time.Now(), Image: resizedImage})
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

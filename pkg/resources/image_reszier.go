package resources

import (
	"time"

	"github.com/google/uuid"
)

type TargetDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Image struct {
	CreatedAt  time.Time         `json:"created_at,omitempty" gorm:"column:inserted_at"`
	UUID       uuid.UUID         `json:"uuid" gorm:"column:uuid"`
	Image      []byte            `json:"image,omitempty" gorm:"column:resized_image"`
	Dimensions *TargetDimensions `json:"dimensions" gorm:"-"`
}

type Request struct {
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"column:inserted_at"`
	UUID      uuid.UUID `json:"uuid" gorm:"column:uuid"`
	Status    string    `json:"status" gorm:"column:status"`
}

// Status if the image resize request
const (
	Processing string = "Processing"
	Done       string = "Done"
	Failed     string = "Failed"
)

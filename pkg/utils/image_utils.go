package utils

import (
	"errors"
	"image"
	"mime/multipart"
	"strconv"
)

// Validating the image dimensions and the image format
// TODO: decode more formats other than JPEG and PNG
func ValidateImageResolution(file *multipart.File, maxSize int64) error {
	imageConfig, _, err := image.DecodeConfig(*file)
	if err != nil {
		return err
	}

	if int64(imageConfig.Width*imageConfig.Height) > maxSize {
		return errors.New("Max file resolution allowd is: " + strconv.FormatInt(maxSize/1024/1024, 10) + " MP.")
	}
	return nil

}

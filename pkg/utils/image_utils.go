package utils

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"mime/multipart"
	"strconv"

	"github.com/nfnt/resize"
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

func ResizeImage(imageBytes []byte, width uint, height uint) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}

	m := resize.Resize(width, height, img, resize.Lanczos3)

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, m, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

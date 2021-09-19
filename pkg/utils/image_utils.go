package utils

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"

	"io"
	"mime/multipart"
	"strconv"

	"github.com/nfnt/resize"
)

// Validating the image dimensions and the image format
// TODO: decode more formats other than JPEG and PNG
func ValidateImageResolution(file multipart.File, maxSize int64) error {
	imageConfig, _, err := image.DecodeConfig(file)
	if err != nil {
		return err
	}

	if int64(imageConfig.Width*imageConfig.Height) > maxSize {
		return errors.New("Max file resolution allowd is: " + strconv.FormatInt(maxSize/1024/1024, 10) + " MP.")
	}

	// reset the file reader
	file.Seek(0, io.SeekStart)

	return nil

}

func ResizeImage(imageBytes []byte, width uint, height uint) ([]byte, error) {
	img, format, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}

	m := resize.Resize(width, height, img, resize.Lanczos3)

	buf := new(bytes.Buffer)
	if format == "jpeg" {
		err = jpeg.Encode(buf, m, nil)
	} else if format == "png" {
		err = png.Encode(buf, m)
	} else {
		return []byte{}, errors.New("Ivalid image format, Allowed formats are: jpeg and png")
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

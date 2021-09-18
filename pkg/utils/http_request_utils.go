package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func ValidateRequestSize(w http.ResponseWriter, r *http.Request, maxSize int64) error {
	sizeExceededError := fmt.Sprintf("Max allowd file size is: %s MB", strconv.FormatInt(maxSize/1024/1024, 10))

	if r.ContentLength > maxSize {
		return errors.New(sizeExceededError)
	}

	// Limiting the number of bytes read from the request (The content length is not set for chunked request bodies)
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		return errors.New(sizeExceededError)
	}

	return nil
}

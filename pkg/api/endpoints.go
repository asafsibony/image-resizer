package api

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

var MAX_UPLOAD_SIZE int64 = 10 * 1024 * 1024     //10MB
var MAX_RESOLUTION_SIZE int64 = 10 * 1024 * 1024 //10MP

type ResizeDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ResizeRequest struct {
	ID           uuid.UUID
	Width        int
	Height       int
	Status       string
	ResizedImage []byte
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	s.logger.Debug(r.Context())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("asaf"))
}

func (s *Server) uploadImage(w http.ResponseWriter, r *http.Request) {
	// Check max file constraints
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		http.Error(w, "Max file size allowd is: "+strconv.FormatInt(MAX_UPLOAD_SIZE/1024/1024, 10)+" MB.", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// TODO: decode more formats other than JPEG and PNG
	imageConfig, _, err := image.DecodeConfig(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if int64(imageConfig.Width*imageConfig.Height) > MAX_RESOLUTION_SIZE {
		http.Error(w, "Max file resolution allowd is: "+strconv.FormatInt(MAX_RESOLUTION_SIZE/1024/1024, 10)+" MP.", http.StatusBadRequest)
		return
	}

	s.logger.Debug("dimensions: ", r.FormValue("dimensions"))
	resizeDimensions := ResizeDimensions{}
	err = json.Unmarshal([]byte(r.FormValue("dimensions")), &resizeDimensions)
	if err != nil {
		http.Error(w, "Failed to parse the json with the desires resize dimensions.", http.StatusBadRequest)
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	imageBytes := StreamToByte(file)

	resizeRequest := ResizeRequest{}
	resizeRequest.Height = resizeDimensions.Height
	resizeRequest.Width = resizeDimensions.Width
	resizeRequest.ID = uuid.New()
	resizeRequest.ResizedImage = imageBytes

	// send to kafka
	// save to redis
	// save to postgres

	w.WriteHeader(http.StatusOK)
	return
}

func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

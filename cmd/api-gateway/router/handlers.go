package router

import (
	"encoding/json"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/asafsibony/image-resizer/pkg/resources"
	"github.com/asafsibony/image-resizer/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	MAX_UPLOAD_SIZE     int64 = 10 * 1024 * 1024 // 10MB
	MAX_RESOLUTION_SIZE int64 = 10 * 1024 * 1024 // 10MP
)

func (router *Router) uploadImage(w http.ResponseWriter, r *http.Request) {
	// TODO: Break to smaller functions
	// Validating image size constraint before start reading
	if r.ContentLength > MAX_UPLOAD_SIZE {
		http.Error(w, "Max allowd file size is: "+strconv.FormatInt(MAX_UPLOAD_SIZE/1024/1024, 10)+" MB.", http.StatusBadRequest)
		return
	}

	// Limiting the number of bytes read from the request (The content length is not set for chunked request bodies)
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		http.Error(w, "Max allowd file size is: "+strconv.FormatInt(MAX_UPLOAD_SIZE/1024/1024, 10)+" MB.", http.StatusBadRequest)
		return
	}

	// Getting the image stream from the request
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate image format and resolution
	if utils.ValidateImageResolution(&file, MAX_RESOLUTION_SIZE) != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Getting the dimensions JSON from the request
	targetDimensions := &resources.TargetDimensions{}
	err = json.Unmarshal([]byte(r.FormValue("dimensions")), targetDimensions)
	if err != nil {
		http.Error(w, "Failed to parse the json with the desires resize dimensions.", http.StatusBadRequest)
		return
	}

	// reset the file reader
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	imageBytes := utils.StreamToByte(file)

	uuid := uuid.New()
	imageRequest := &resources.Request{
		UUID:      uuid,
		Status:    resources.Processing,
		CreatedAt: time.Now(),
	}

	// save to redis, keep going even if save failed
	router.redisClient.Set(uuid.String()+":status", imageRequest.Status)

	// save to postgres
	result := router.psqlClient.Database.Table("requests").Create(imageRequest)
	if result.Error != nil {
		router.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// send to kafka queue
	image := &resources.Image{
		UUID:       uuid,
		Image:      imageBytes,
		Dimensions: targetDimensions,
	}

	imageJson, err := json.Marshal(image)
	if err != nil {
		router.logger.Error(err)
		return
	}
	err = router.kafkaProducer.ProduceMessage(router.imageResizeTopic, uuid.String(), imageJson)
	if err != nil {
		router.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(imageRequest.UUID.String()))
	// TODO: Change all responses to json

	// w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// w.WriteHeader(http.StatusOK)
	// if err := json.NewEncoder(w).Encode(todos); err != nil {
	//     panic(err)
	// }

	return
}

func (router *Router) getStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reqUUID := vars["uuid"]
	status, err := router.getRequestStatus(reqUUID)
	if err != nil {
		http.Error(w, "The rquested uuid not found.", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(status))
}

func (router *Router) downloadImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reqUUID := vars["uuid"]
	status, err := router.getRequestStatus(reqUUID)
	if err != nil {
		http.Error(w, "The rquested image uuid not found.", http.StatusNotFound)
		return
	}

	if status == resources.Processing {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("The requested resized image is still in process, please try again later."))
		return
	} else if status == resources.Failed {
		http.Error(w, "The rquested image failed to resize.", http.StatusInternalServerError)
		return
	} else if status == resources.Done {
		resizedImage, err := router.getResizedImage(reqUUID)
		if err != nil {
			http.Error(w, "Failed to download the requested image.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resizedImage))
		return
	}

	// Requested image status undefined
	http.Error(w, "OOPS something went wrong. please try to re-upload your image.", http.StatusInternalServerError)
}

// -------------------------------------
// Helpers:
// --------
// --------
func (r *Router) getRequestStatus(reqUUID string) (string, error) {
	// get request status from cache
	status, err := r.redisClient.Get(reqUUID + ":status")
	if err == nil {
		return status, nil
	}

	// if not exist in cache get from DB
	uuid, err := uuid.Parse(reqUUID)
	if err != nil {
		r.logger.Error(err)
		return "", err
	}

	request := &resources.Request{}
	result := r.psqlClient.Database.Table("requests").First(request, "uuid = ?", uuid)
	if result.Error != nil {
		r.logger.Error(err)
		return "", err
	}

	// Update the cache with the fetch status from the DB
	r.redisClient.Set(reqUUID+":status", request.Status)

	return status, nil
}

func (r *Router) getResizedImage(reqUUID string) ([]byte, error) {
	// get resized image from cache
	resizedImage, err := r.redisClient.Get(reqUUID + ":resized_image")
	if err == nil {
		return []byte(resizedImage), nil
	}

	// if not exist in cache get from DB
	uuid, err := uuid.Parse(reqUUID)
	if err != nil {
		r.logger.Error(err)
		return []byte{}, err
	}

	resized_image := &resources.Image{}
	result := r.psqlClient.Database.Table("images").First(resized_image, "uuid = ?", uuid)
	if result.Error != nil {
		r.logger.Error(err)
		return []byte{}, err
	}

	// Update the cache
	r.redisClient.Set(reqUUID+":resized_image", resized_image.Image)

	return resized_image.Image, nil
}

package router

import (
	"encoding/json"
	"net/http"

	"github.com/asafsibony/image-resizer/pkg/resources"
	"github.com/asafsibony/image-resizer/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	MAX_UPLOAD_SIZE     int64 = 10 * 1024 * 1024 // 10MB
	MAX_RESOLUTION_SIZE int64 = 10 * 1024 * 1024 // 10MP
)

// upload endpoint
func (router *Router) uploadImage(w http.ResponseWriter, r *http.Request) {
	// Validating image size constraint
	err := utils.ValidateRequestSize(w, r, MAX_UPLOAD_SIZE)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Getting the image file reader from the request
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		router.logger.Error(err.Error())
		http.Error(w, "Bad request.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate image format and resolution
	if utils.ValidateImageResolution(file, MAX_RESOLUTION_SIZE) != nil {
		router.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Getting the target dimensions JSON from the request
	targetDimensions := &resources.TargetDimensions{}
	err = json.Unmarshal([]byte(r.FormValue("dimensions")), targetDimensions)
	if err != nil {
		router.logger.Error(err.Error())
		http.Error(w, "Failed to parse the dimensions json.", http.StatusBadRequest)
		return
	}

	imageUuid := uuid.New()

	// save to redis, keep going even if save failed
	router.saveRequestToCache(imageUuid.String())

	// save to postgres
	err = router.saveRequestToDB(imageUuid)
	if err != nil {
		router.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// send to kafka queue
	imageBytes := utils.StreamToByte(file)
	err = router.sendRequestToQueue(imageUuid, imageBytes, targetDimensions, fileHeader.Filename)
	if err != nil {
		router.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(imageUuid.String()))
	return
}

// Status endpoint
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

// Download endpoint
func (router *Router) downloadImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageUuid := vars["uuid"]
	status, err := router.getRequestStatus(imageUuid)
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
		resizedImage, err := router.getResizedImage(imageUuid)
		if err != nil {
			http.Error(w, "Failed to download the requested image.", http.StatusInternalServerError)
			return
		}

		ContentDisposition := "attachment; filename=" + resizedImage.Name
		w.Header().Set("Content-Disposition", ContentDisposition)
		w.Header().Set("Content-Type", "image/jpg")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resizedImage.Image))
		return
	}

	// Requested image status undefined
	http.Error(w, "OOPS something went wrong. please try to re-upload your image.", http.StatusInternalServerError)
}

package handlers

import (
	"encoding/json"
	"net/http"

	"VidVendor/config"
	"VidVendor/services"
)

func UploadVideoHandler(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
	var requestBody struct {
		URL string `json:"url"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	// just add the URL to the queue -> the goroutine will download the video and generate UUID asynchronously
	if err := services.AddVideoURL(requestBody.URL); err != nil {
		http.Error(w, "Failed to add video URL", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetNextVideoHandler(w http.ResponseWriter, r *http.Request) {
	nextVideoUUID := services.GetNextVideo()
	videoPath := config.GetConfig().OutputDirectory + "/" + nextVideoUUID + ".mp4"
	w.Header().Set("Content-Type", "application/mp4")
	http.ServeFile(w, r, videoPath)
	services.ScheduleVideoForDeletion(nextVideoUUID)
}

func StopStreamHandler(w http.ResponseWriter, r *http.Request) {
	services.EndStream()
	w.WriteHeader(http.StatusOK)
}

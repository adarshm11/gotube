package services

import (
	"VidVendor/config"
)

var (
	URLQueue      chan string // Stores the video URLs to be downloaded
	PlaybackQueue chan string // Stores the UUIDs of downloaded videos ready for playback
	DeletionQueue chan string // Stores the UUIDs of videos to be deleted
)

func InitQueues(cfg *config.Config) {
	URLQueue = make(chan string, cfg.URLQueueBufferSize)
	PlaybackQueue = make(chan string, cfg.PlaybackQueueBufferSize)
	DeletionQueue = make(chan string, cfg.DeleteQueueBufferSize)
}

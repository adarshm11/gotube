package services

import (
	"log"
	"os"

	"VidVendor/config"
)

// Adds a new URL to the URLQueue to be downloaded
func AddVideoURL(url string) error {
	URLQueue <- url
	log.Printf("Video URL added to queue: %s", url)
	return nil
}

// Retrieves the UUID of the next video to be played from the PlaybackQueue
func GetNextVideo() string {
	uuid, ok := <-PlaybackQueue
	if !ok {
		log.Printf("PlaybackQueue is closed, exiting...")
		return ""
	}

	log.Printf("Next video ID: %s\n", uuid)
	return uuid
}

// Goroutine that continuously listens to the DeletionQueue and deletes videos
func VideoCleanup(cfg *config.Config, sigchan chan os.Signal) {
	for uuid := range DeletionQueue {
		select {
		case <-sigchan:
			log.Printf("Received shutdown signal, exiting VideoCleanup goroutine...")
			return
		default:
			videoPath := cfg.OutputDirectory + "/" + uuid + ".mp4"
			if err := os.Remove(videoPath); err != nil {
				log.Printf("Failed to delete video %s: %v", uuid, err)
			} else {
				log.Printf("Video %s deleted successfully", uuid)
			}
		}
	}

}

func ScheduleVideoForDeletion(uuid string) {
	DeletionQueue <- uuid
}

// Empties the PlaybackQueue and schedules all videos for deletion
func EndStream() {
	for len(PlaybackQueue) > 0 {
		DeletionQueue <- <-PlaybackQueue
	}
}

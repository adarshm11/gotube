package services

import (
	"io"
	"log"
	"os"

	"github.com/kkdai/youtube/v2"

	"VidVendor/config"
	"VidVendor/utils"
)

var client = youtube.Client{}

// Extracts URLs to be downloaded from the URLQueue, downloads the videos, saves them with a generated UUID, and adds the UUID to the PlaybackQueue
func DownloadVideo(cfg *config.Config, sigchan chan os.Signal) error {
	for {
		select {
		case <-sigchan:
			log.Printf("Received shutdown signal, exiting DownloadVideo goroutine...")
			return nil
		default:
			url, ok := <-URLQueue
			if !ok {
				log.Printf("URLQueue is closed, exiting...")
				return nil
			}
			video, err := client.GetVideo(url)
			if err != nil {
				log.Printf("Failed to get video: %v", err)
				continue
			}
			formats := video.Formats.WithAudioChannels() // get only formats with audio
			stream, _, err := client.GetStream(video, &formats[0])
			if err != nil {
				log.Printf("Failed to get video stream: %v", err)
				continue
			}
			defer stream.Close()

			// create the video file
			videoId := utils.GenerateUUID()
			videoPath := cfg.OutputDirectory + "/" + videoId + ".mp4"
			videoFile, err := os.Create(videoPath)
			if err != nil {
				log.Printf("Failed to create video file: %v", err)
				continue
			}
			defer videoFile.Close()

			// save the video stream to the file
			_, err = io.Copy(videoFile, stream)
			if err != nil {
				log.Printf("Failed to save video stream: %v", err)
				os.Remove(videoPath)
				continue
			}

			PlaybackQueue <- videoId
			log.Printf("Video %s downloaded successfully\n", videoId)
		}
	}
}

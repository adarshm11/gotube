package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"VidVendor/config"
	"VidVendor/handlers"
	"VidVendor/services"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	configPath := flag.String("config", "config.yml", "Path to the config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	r := http.NewServeMux()

	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "pong"}`))
	})

	r.HandleFunc("/upload", handlers.UploadVideoHandler)
	r.HandleFunc("/next", handlers.GetNextVideoHandler)
	r.HandleFunc("/stop", handlers.StopStreamHandler)

	services.InitQueues(cfg)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)

	go services.DownloadVideo(cfg, sigchan)
	go services.VideoCleanup(cfg, sigchan)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: corsMiddleware(r),
	}

	go func() {
		log.Println("server running on port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v\n", err)
		}
	}()

	<-sigchan
	log.Println("Shutting down server...")
	services.EndStream()
	close(services.DeletionQueue)
	close(services.PlaybackQueue)
	close(services.URLQueue)
	server.Shutdown(context.TODO())
	log.Println("Server stopped")
}

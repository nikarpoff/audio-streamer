package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikarpoff/audio-streamer/internal/audio"
	"github.com/nikarpoff/audio-streamer/internal/config"
)

func main() {
	cfg := config.DefaultConfig()

	// Create capture object
	capture, err := audio.NewCapture(cfg)
	if err != nil {
		log.Fatal("Failed to create capture:", err)
	}
	defer capture.Stop() // defer call closing

	// Create playback object
	playback, err := audio.NewPlayback(cfg)
	if err != nil {
		log.Fatal("Failed to create playback:", err)
	}
	defer playback.Stop() // defer call closing

	if err := capture.Start(); err != nil {
		log.Fatal("Failed to start capture:", err)
	}

	log.Println("Loopback test started - you should hear your microphone")
	log.Println("Press Ctrl+C to stop")

	// Handling interruption
	sigChan := make(chan os.Signal, 1) // buffer with one signal
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	running := true
	for running {
		select {
		// Try to recieve data from capture channel
		case data, ok := <-capture.Buffer:
			if !ok {
				running = false
				break
			}

			// If recieved then write into playback channel
			playback.Write(data)

		case <-sigChan:
			running = false
		}
	}

	close(sigChan)
	log.Println("Loopback test stopped")
}

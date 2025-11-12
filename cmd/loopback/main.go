package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikarpoff/audio-streamer/internal/audio"
	"github.com/nikarpoff/audio-streamer/internal/config"
)

func main() {
	cfg := config.DefaultConfig()

	// Create capture object
	capture, err := audio.NewCapture(cfg, true)
	if err != nil {
		log.Fatal("Failed to create capture:", err)
	}
	defer capture.Stop() // defer call closing

	loopPerfomance := audio.Metric{}
	playPerfomance := audio.Metric{}
	var (
		lastLoop time.Time
		lastPlay time.Time

		dtLoop time.Duration
		dtPlay time.Duration
	)

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

	go audio.Run(&loopPerfomance, "Loop", 5)
	go audio.Run(&playPerfomance, "Playback Write Delay", 5)

	// Handling interruption
	sigChan := make(chan os.Signal, 1) // buffer with one signal
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	running := true
	lastLoop = time.Now()
	for running {
		select {
		// Try to recieve data from capture channel
		case data, ok := <-capture.Buffer:
			if !ok {
				running = false
				break
			}

			// If recieved then write into playback channel
			lastPlay = time.Now()
			playback.Write(data)
			dtPlay = time.Since(lastPlay)
			playPerfomance.Add(dtPlay)

			dtLoop = time.Since(lastLoop)
			loopPerfomance.Add(dtLoop)
			lastLoop = time.Now()
		case <-sigChan:
			running = false
		}
	}

	close(sigChan)
	log.Println("Loopback test stopped")
}

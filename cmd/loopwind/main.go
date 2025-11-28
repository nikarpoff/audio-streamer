package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikarpoff/audio-streamer/internal/audio"
	"github.com/nikarpoff/audio-streamer/internal/config"
)

func main() {
	cfg := config.DefaultConfig()

	if err := audio.InitializePortaudio(); err != nil {
		log.Fatal("Initialization PortAudio error!:", err)
	}
	devices := audio.GetDevices()

	// Show all devices
	for _, device := range devices {
		fmt.Printf("%s Device: %s\n", device.HostApi.Type, device.Name)
	}

	audioStream, err := audio.NewAudioStream(cfg, devices[3], devices[8])
	if err != nil {
		log.Fatal("Failed to create audio stream:", err)
	}
	defer audioStream.Stop() // defer call closing

	// Try to start record/playback
	if err := audioStream.Start(); err != nil {
		log.Fatal("Failed to start record/playback:", err)
	}

	log.Println("Loopback test started - you should hear your microphone")
	log.Println("Press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1) // buffer with one signal
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	running := true
	for running {
		select {
		// Try to recieve data from capture channel
		case data, ok := <-audioStream.InputBuffer:
			if !ok {
				log.Println("Recieved not ok status from InputBuffer! Stop recording!")
				running = false
				break
			}

			// If recieved then write into playback channel
			audioStream.OutputBuffer <- data
		case <-sigChan:
			audioStream.StopCapture <- os.Interrupt
			audioStream.StopPlayback <- os.Interrupt
			running = false
		}
	}

	close(sigChan)
}

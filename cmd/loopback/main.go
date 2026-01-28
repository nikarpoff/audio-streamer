package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikarpoff/audio-streamer/internal/audio"
	"github.com/nikarpoff/audio-streamer/internal/config"
	"github.com/nikarpoff/audio-streamer/internal/utils"
)

func main() {
	utils.PrintWelcome("Test Loopback")

	cfg := config.DefaultConfig()

	if err := audio.InitializePortaudio(); err != nil {
		log.Fatal("Initialization PortAudio error!:", err)
	}

	// Select input and output devices
	utils.ShowHosts()
	devices := audio.GetDevices()
	utils.ShowDevices(devices)
	inputDevice, outputDevice := utils.SelectDevice(devices)

	audioStream, err := audio.NewAudioStream(cfg, inputDevice, outputDevice)
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

	// Loopback
	go func() {
		for {
			data, ok := <-audioStream.InputBuffer
			if !ok {
				log.Println("Recieved not ok status from InputBuffer! Stop recording!")
				return
			}

			audioStream.OutputBuffer <- data
		}
	}()

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
	log.Println("Shutting down...")
	close(interrupt)
}

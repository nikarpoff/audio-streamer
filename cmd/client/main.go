package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/nikarpoff/audio-streamer/internal/audio"
	"github.com/nikarpoff/audio-streamer/internal/config"
	"github.com/nikarpoff/audio-streamer/internal/utils"
)

func main() {
	utils.PrintWelcome("Client")

	if err := audio.InitializePortaudio(); err != nil {
		log.Fatal("Initialization PortAudio error!:", err)
	}

	var socketAddress string
	fmt.Printf("Please, select server's socket address. By default 'ws://kbks:7001/ws' [enter for default]: ")
	fmt.Scan(&socketAddress)
	fmt.Println()

	isValidAddress := utils.IsValidWebSocketURL(socketAddress)

	if !isValidAddress {
		fmt.Printf("You provide invalid server's socket address... Please, try again")
		return
	}

	cfg := config.DefaultConfig()

	// Select input and output devices
	utils.ShowHosts()
	devices := audio.GetDevices()
	utils.ShowDevices(devices)
	inputDevice, outputDevice := utils.SelectDevice(devices)

	// Create audio capture and playback
	audioStream, err := audio.NewAudioStream(cfg, inputDevice, outputDevice)
	if err != nil {
		log.Fatal("Failed to create audio stream:", err)
	}
	defer audioStream.Stop() // defer call closing

	// Parse server URL
	u, err := url.Parse(socketAddress)
	if err != nil {
		log.Fatal("Failed to parse server address:", err)
	}

	log.Printf("\nConnecting to %s", u.String())

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}
	defer conn.Close()

	log.Println("Connected to server. Starting audio streams...")

	// Try to start record/playback
	if err := audioStream.Start(); err != nil {
		log.Fatal("Failed to start record/playback:", err)
	}

	log.Println("You joined to server! Other clients can hear your microphone!")
	log.Println("Press Ctrl+C to stop")

	// Handle incoming audio messages
	go func() {
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			if messageType == websocket.BinaryMessage {
				// Convert bytes back to int16 samples
				samples := make([]int16, len(data)/2)
				for i := 0; i < len(samples); i++ {
					samples[i] = int16(data[i*2]) | (int16(data[i*2+1]) << 8)
				}
				audioStream.OutputBuffer <- samples
			}
		}
	}()

	// Send captured audio
	go func() {
		for {
			data, ok := <-audioStream.InputBuffer
			if !ok {
				return
			}

			// Convert int16 samples to bytes
			byteData := make([]byte, len(data)*2)
			for i, sample := range data {
				byteData[i*2] = byte(sample)
				byteData[i*2+1] = byte(sample >> 8)
			}

			err := conn.WriteMessage(websocket.BinaryMessage, byteData)
			if err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}()

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
	log.Println("Shutting down...")
	close(interrupt)
}

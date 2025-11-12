package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/nikarpoff/audio-streamer/internal/audio"
	"github.com/nikarpoff/audio-streamer/internal/config"
)

var (
	serverAddr = flag.String("server", "ws://localhost:8080/ws", "WebSocket server address")
)

func main() {
	flag.Parse()

	log.SetFlags(0)

	cfg := config.DefaultConfig()

	// Create audio capture and playback
	capture, err := audio.NewCapture(cfg, false)
	if err != nil {
		log.Fatal("Failed to create capture:", err)
	}
	defer capture.Stop()

	playback, err := audio.NewPlayback(cfg)
	if err != nil {
		log.Fatal("Failed to create playback:", err)
	}
	defer playback.Stop()

	// Parse server URL
	u, err := url.Parse(*serverAddr)
	if err != nil {
		log.Fatal("Failed to parse server address:", err)
	}

	log.Printf("Connecting to %s", u.String())

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}
	defer conn.Close()

	log.Println("Connected to server. Starting audio streams...")

	// Start capture
	if err := capture.Start(); err != nil {
		log.Fatal("Failed to start capture:", err)
	}

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
				playback.Write(samples)
			}
		}
	}()

	// Send captured audio
	go func() {
		for {
			data, ok := <-capture.Buffer
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
}

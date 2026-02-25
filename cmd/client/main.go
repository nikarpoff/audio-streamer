package main

import (
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/nikarpoff/audio-streamer/internal/audio"
	"github.com/nikarpoff/audio-streamer/internal/config"
	"github.com/nikarpoff/audio-streamer/internal/protocol"
	"github.com/nikarpoff/audio-streamer/internal/utils"
)

const udpSocketBufferSize = 1 << 20 // 1 MiB

func main() {
	utils.PrintWelcome("Client")

	if err := audio.InitializePortaudio(); err != nil {
		log.Fatal("Initialization PortAudio error!:", err)
	}

	serverAddress := utils.SelectAddress()

	cfg := config.DefaultConfig()

	// Select audio backend
	hostApis := audio.GetAPIS()
	utils.ShowHosts(hostApis)
	hostApi := utils.SelectHostAPI(hostApis)

	// Select input and output devices
	devices := audio.GetDevices(hostApi)
	utils.ShowDevices(devices)
	inputDevice, outputDevice := utils.SelectDevice(devices)

	// Show configuration
	utils.ShowAudioParams(cfg, int(inputDevice.DefaultSampleRate), inputDevice.MaxInputChannels, outputDevice.MaxOutputChannels)

	// Select optimal parameters
	cfg = utils.SelectConfig(inputDevice.MaxInputChannels, outputDevice.MaxOutputChannels)
	utils.ShowAudioParams(cfg, int(cfg.SampleRate), cfg.InputChannels, cfg.OutputChannels)

	// Create audio capture and playback
	audioStream, err := audio.NewAudioStream(cfg, inputDevice, outputDevice)
	if err != nil {
		log.Fatal("Failed to create audio stream:", err)
	}
	defer audioStream.Stop() // defer call closing

	log.Printf("\nConnecting to %s", serverAddress)

	serverAddr, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil {
		log.Fatal("Failed to resolve UDP address:", err)
	}

	// Connect to server
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatal("Failed to connect to UDP server:", err)
	}
	defer conn.Close()

	_ = conn.SetReadBuffer(udpSocketBufferSize)
	_ = conn.SetWriteBuffer(udpSocketBufferSize)

	log.Println("Connected to server. Starting audio streams...")

	// Try to start record/playback
	if err := audioStream.Start(); err != nil {
		log.Fatal("Failed to start record/playback:", err)
	}

	log.Println("You joined to server! Other clients can hear your microphone!")
	log.Println("Press Ctrl+C to stop")

	// Handle incoming audio messages
	go func() {
		buffer := make([]byte, 4096)

		var lastSequence uint32
		hasSequence := false

		for {
			n, err := conn.Read(buffer)
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			sequence, samples, err := protocol.DecodeAudioPacket(buffer[:n])
			if err != nil {
				continue
			}

			if hasSequence && !protocol.SequenceAhead(sequence, lastSequence) {
				continue
			}
			lastSequence = sequence
			hasSequence = true

			select {
			case audioStream.OutputBuffer <- samples:
			default:
				<-audioStream.OutputBuffer
				audioStream.OutputBuffer <- samples
			}
		}
	}()

	// Send captured audio
	go func() {
		sequence := uint32(1)

		for {
			data, ok := <-audioStream.InputBuffer
			if !ok {
				return
			}

			// Convert int16 samples to bytes
			packet := protocol.EncodeAudioPacket(sequence, data)
			sequence++

			_, err := conn.Write(packet)
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

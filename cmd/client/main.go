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
const maxConcealedPacketsPerGap = 2 // Packets number that can be replaced in gap

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
	log.Printf("Network mode: mono (capture mixed from %dch, playback expanded to %dch)", cfg.InputChannels, cfg.OutputChannels)

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

	// Start goroutines for reading and writing audio packets
	go readAudioPackets(conn, audioStream, cfg.OutputChannels)
	go writeAudioPackets(conn, audioStream, cfg.InputChannels)

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
	log.Println("Shutting down...")
	close(interrupt)
}

func readAudioPackets(conn *net.UDPConn, audioStream *audio.AudioStream, outputChannels int) {
	// Handle incoming audio messages
	buffer := make([]byte, 8192)

	var lastSequence uint32
	var lastSamples []int16
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

		// Conceal packet loss if sequence gap detected
		if hasSequence {
			gap := int(sequence - lastSequence - 1)
			if gap > 0 && len(lastSamples) > 0 {
				log.Printf("Packet loss detected: gap of %d packets (last sequence: %d, current: %d)", gap, lastSequence, sequence)
				concealPacketLoss(lastSamples, gap, audioStream)
			}
		}

		lastSequence = sequence
		hasSequence = true

		lastSamples = make([]int16, len(samples))
		copy(lastSamples, samples)

		playbackSamples := audio.MonoToInterleaved(samples, outputChannels)

		select {
		case audioStream.OutputBuffer <- playbackSamples:
		default:
			<-audioStream.OutputBuffer
			audioStream.OutputBuffer <- playbackSamples
		}
	}
}

func writeAudioPackets(conn *net.UDPConn, audioStream *audio.AudioStream, inputChannels int) {
	sequence := uint32(1)

	for {
		data, ok := <-audioStream.InputBuffer
		if !ok {
			return
		}

		// Convert int16 samples to bytes
		mono := audio.MixToMonoInterleaved(data, inputChannels)
		packet := protocol.EncodeAudioPacket(sequence, mono)
		sequence++

		_, err := conn.Write(packet)
		if err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

func concealPacketLoss(lastSamples []int16, gap int, audioStream *audio.AudioStream) {
	// Simple concealment strategy: repeat last received samples for a limited number of lost packets.
	if gap > maxConcealedPacketsPerGap {
		gap = maxConcealedPacketsPerGap
	}

	for i := 0; i < gap; i++ {
		concealed := make([]int16, len(lastSamples))
		copy(concealed, lastSamples)

		select {
		case audioStream.OutputBuffer <- concealed:
		default:
			<-audioStream.OutputBuffer
			audioStream.OutputBuffer <- concealed
		}
	}
}

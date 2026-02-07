package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gordonklaus/portaudio"
	"github.com/nikarpoff/audio-streamer/internal/config"
)

func readInt(message string, defaultValue int) int {
	fmt.Printf("%s (by default %d): ", message, defaultValue)
	var input string
	fmt.Scanln(&input)

	// Trim the newline character from the input
	input = strings.TrimSpace(input)

	// Check if the input is empty
	if input == "" {
		return defaultValue
	} else {
		// Attempt to convert the input string to an integer
		intValue, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("An error occurred while parsing input: ", err)
			return defaultValue
		}

		return intValue
	}
}

func ShowHosts(hostApis []*portaudio.HostApiInfo) {
	if hostApis != nil {
		fmt.Printf("Available API hosts:\n")
		for i, hostApi := range hostApis {
			fmt.Printf("%d: %s (%s)\n", i, hostApi.Type, hostApi.Name)
		}
	}
}

func ShowDevices(devices []*portaudio.DeviceInfo) {
	// Show all devices
	fmt.Printf("Available devices:\n")
	for i, device := range devices {
		fmt.Printf("%d: %s Device: %s\n", i, device.HostApi.Type, device.Name)
	}
}

func ShowAudioParams(cfg *config.AudioConfig, inputChannels int, outputChannels int) {
	// Show all devices
	fmt.Printf("\n==========================================================\n")
	fmt.Printf("Current audio stream params:\n")
	fmt.Printf("\t Capture channels: %d\n", inputChannels)
	fmt.Printf("\t Play channels: %d\n", outputChannels)
	fmt.Printf("\t Sample rate: %f\n", int(cfg.SampleRate))
	fmt.Printf("\t Buffer size: %d\n", cfg.BufferSize)
	fmt.Printf("\t Bit depth: %d\n", cfg.BitDepth)
	fmt.Printf("==========================================================\n\n")
}

func SelectHostAPI(hostApis []*portaudio.HostApiInfo) *portaudio.HostApiInfo {
	// Asks user to prefered audio Input/Output
	totalHosts := len(hostApis)
	var hostIdx int
	fmt.Printf("\nPlease, select prefered audio host API (for low-latency use ASIO/ALSA/Core Audio)\n")
	fmt.Printf("Provide index, e.g. >> 5\n")
	fmt.Printf(">> ")

	_, err := fmt.Scan(&hostIdx)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	if hostIdx >= totalHosts || hostIdx < 0 {
		log.Fatal("Invalid index! Please, select between 0 and ", totalHosts)
		return nil
	}

	return hostApis[hostIdx]
}

func SelectDevice(devices []*portaudio.DeviceInfo) (*portaudio.DeviceInfo, *portaudio.DeviceInfo) {
	// Asks user to prefered audio Input/Output
	totalDevices := len(devices)
	var captureDeviceIdx int
	var playbackDeviceIdx int
	fmt.Printf("\nPlease, select prefered Input and Output devices indexes, e.g. >> 0 1\n")
	fmt.Printf(">> ")

	_, err := fmt.Scan(&captureDeviceIdx, &playbackDeviceIdx)
	if err != nil {
		log.Fatal(err)
		return nil, nil
	}
	if playbackDeviceIdx >= totalDevices || playbackDeviceIdx < 0 {
		log.Fatal("Invalid output device index! Please, select between 0 and ", totalDevices)
		return nil, nil
	}
	if captureDeviceIdx >= totalDevices || captureDeviceIdx < 0 {
		log.Fatal("Invalid input device index! Please, select between 0 and ", totalDevices)
		return nil, nil
	}

	return devices[captureDeviceIdx], devices[playbackDeviceIdx]
}

func SelectAddress() string {
	defaultAddress := config.DefaultServerAddress()
	fmt.Printf("Please, select server's socket address. By default '%s' [enter for default]: ", defaultAddress)
	var input string
	fmt.Scanln(&input)
	socketAddress := strings.TrimSpace(input)

	if socketAddress == "" {
		socketAddress = defaultAddress
		fmt.Printf("Using default address: %s\n", socketAddress)
	} else {
		isValidAddress := IsValidServerAddress(socketAddress)

		if !isValidAddress {
			log.Fatal("You provide invalid server's socket address... Please, try again")
			return ""
		}
	}

	return socketAddress
}

func SelectConfig(maxInputChannels int, maxOutputChannels int) *config.AudioConfig {
	fmt.Printf("Please select the optimal settings or leave the default settings (Enter)...\n")
	defaultConfig := config.DefaultConfig()

	inputChannels := readInt("Input channels number", defaultConfig.InputChannels)
	if !IsValidBoundedInteger(inputChannels, 0, maxInputChannels) {
		fmt.Printf("Invalid channels number! Select between 0 and %d", maxInputChannels)
		inputChannels = defaultConfig.InputChannels
	}

	outputChannels := readInt("Output channels number", defaultConfig.OutputChannels)
	if !IsValidBoundedInteger(outputChannels, 0, maxOutputChannels) {
		fmt.Printf("Invalid channels number! Select between 0 and %d", maxOutputChannels)
		outputChannels = defaultConfig.OutputChannels
	}

	sampleRate := readInt("Sample rate in Hz", int(defaultConfig.SampleRate))
	if !IsValidBoundedInteger(outputChannels, 0, 192000) {
		fmt.Printf("Invalid sample rate! Select between 0 and %d Hz", 192000)
		sampleRate = int(defaultConfig.SampleRate)
	}

	bufferSize := readInt("Buffer size", defaultConfig.BufferSize)
	if !IsValidBoundedInteger(outputChannels, 0, 2048) {
		fmt.Printf("Invalid buffer size! Select between 0 and %d", 2048)
		bufferSize = defaultConfig.BufferSize
	}

	return &config.AudioConfig{
		InputChannels:  inputChannels,
		OutputChannels: outputChannels,
		SampleRate:     float64(sampleRate),
		BufferSize:     bufferSize,
		BitDepth:       defaultConfig.BitDepth,
	}
}

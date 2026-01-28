package utils

import (
	"fmt"
	"log"

	"github.com/gordonklaus/portaudio"
	"github.com/nikarpoff/audio-streamer/internal/audio"
)

func ShowHosts() {
	hostApis := audio.GetAPIS()

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

func SelectDevice(devices []*portaudio.DeviceInfo) (*portaudio.DeviceInfo, *portaudio.DeviceInfo) {
	// Asks user to prefered audio Input/Output
	totalDevices := len(devices)
	var captureDeviceIdx int
	var playbackDeviceIdx int
	fmt.Printf("\nPlease, select prefered Input and Output devices indexes, e.g. >> 0 1: ")

	_, err := fmt.Scan(&captureDeviceIdx, &playbackDeviceIdx)
	if err != nil {
		log.Fatal(err)
	}
	if captureDeviceIdx >= totalDevices || playbackDeviceIdx >= totalDevices {
		log.Fatal("Invalid indices! Max index is ", totalDevices)
	}

	return devices[captureDeviceIdx], devices[playbackDeviceIdx]
}

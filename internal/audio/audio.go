package audio

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/gordonklaus/portaudio"
	"github.com/nikarpoff/audio-streamer/internal/config"
)

const (
	audioQueueSize = 6
)

type AudioStream struct {
	stream       *portaudio.Stream   // PortAudio Stream
	config       *config.AudioConfig // Config (sr, buffer size)
	inputBuffer  []int16             // Internal portaudio buffer
	outputBuffer []int16             // Internal portaudio buffer
	InputBuffer  chan []int16        // External buffer for audio stream (streamer <-> socket)
	OutputBuffer chan []int16        // External buffer for audio stream (streamer <-> socket)

	// Cancelation channels for capture and playback goroutines
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
	stopped  atomic.Bool
}

func InitializePortaudio() error {
	return portaudio.Initialize()
}

func GetDevices(hostApi *portaudio.HostApiInfo) []*portaudio.DeviceInfo {
	devices, err := portaudio.Devices()
	if err != nil {
		log.Fatal("faied to get devices list:", err)
		return nil
	}

	requestedDevices := make([]*portaudio.DeviceInfo, 0, 8)
	for _, device := range devices {
		if device.HostApi.Name == hostApi.Name {
			requestedDevices = append(requestedDevices, device)
		}
	}

	return requestedDevices
}

func GetAPIS() []*portaudio.HostApiInfo {
	fmt.Println("Host APIs:")
	hostApis, err := portaudio.HostApis()
	if err != nil {
		log.Fatal("faied to get host APIs list:", err)
		return nil
	}
	return hostApis
}

func NewAudioStream(cfg *config.AudioConfig, captureDevice *portaudio.DeviceInfo, playbackDevice *portaudio.DeviceInfo) (*AudioStream, error) {
	// Portaudio Should be initialized!
	streamParams := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   captureDevice,
			Channels: cfg.InputChannels,
			Latency:  captureDevice.DefaultLowInputLatency,
		},
		Output: portaudio.StreamDeviceParameters{
			Device:   playbackDevice,
			Channels: cfg.OutputChannels,
			Latency:  playbackDevice.DefaultLowOutputLatency,
		},
		SampleRate:      cfg.SampleRate,
		FramesPerBuffer: cfg.BufferSize,
	}

	// Bufferized audiodata channels
	inputBuffer := make([]int16, cfg.BufferSize*cfg.InputChannels)
	outputBuffer := make([]int16, cfg.BufferSize*cfg.OutputChannels)

	// Create capture stream
	stream, err := portaudio.OpenStream(streamParams, &inputBuffer, &outputBuffer)

	if err != nil {
		return nil, fmt.Errorf("failed to open audio stream: %w", err)
	}

	return &AudioStream{
		stream:       stream,
		config:       cfg,
		inputBuffer:  inputBuffer,
		outputBuffer: outputBuffer,
		InputBuffer:  make(chan []int16, audioQueueSize),
		OutputBuffer: make(chan []int16, audioQueueSize),
		stopCh:       make(chan struct{}),
	}, nil
}

// Tries to start capturing audio
func (s *AudioStream) Start() error {
	if err := s.stream.Start(); err != nil {
		return fmt.Errorf("failed to start audio stream: %w", err)
	}

	log.Printf("Audio capture/playback started: %.0f Hz, frames/buffer: %d\n",
		s.config.SampleRate, s.config.BufferSize)

	s.wg.Add(2) // Wait group for capture and playback goroutines
	go s.startCapture()
	go s.startPlayback()

	return nil
}

func (s *AudioStream) startCapture() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopCh:
			log.Printf("Capturing stopped")
			return
		default:
			if err := s.stream.Read(); err != nil {
				log.Printf("Read (capture) error: %v", err)
				continue
			}

			// Copy data, cause channel can be reused
			data := make([]int16, len(s.inputBuffer))
			copy(data, s.inputBuffer)

			select {
			case s.InputBuffer <- data:
			default:
				// Drop oldest buffered capture block to avoid latency accumulation.
				<-s.InputBuffer
				s.InputBuffer <- data
			}
		}
	}
}

func (s *AudioStream) startPlayback() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopCh:
			log.Printf("Playing stopped")
			return
		case data, ok := <-s.OutputBuffer:
			if !ok {
				log.Println("Recieved not ok status from OutputBuffer! Skip data!")
				continue
			}

			copy(s.outputBuffer, data)

			if err := s.stream.Write(); err != nil {
				log.Printf("Writing (playback) error: %v", err)
				continue
			}
		}
	}
}

// Stops and closes portaudio stream, closes buffers and terminates portaudio!
func (s *AudioStream) Stop() error {
	s.stopOnce.Do(func() {
		s.stopped.Store(true)
		close(s.stopCh)
	})

	if s.stream != nil {
		if err := s.stream.Stop(); err != nil {
			return err
		}
		if err := s.stream.Close(); err != nil {
			return err
		}
	}

	// Wait for capture and playback goroutines to finish
	s.wg.Wait()

	if err := portaudio.Terminate(); err != nil {
		return err
	}

	close(s.InputBuffer)
	close(s.OutputBuffer)

	log.Println("Audio capture/playback stopped")
	return nil
}

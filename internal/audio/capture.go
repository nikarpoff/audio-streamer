package audio

import (
	"context"
	"fmt"
	"log"

	"github.com/gordonklaus/portaudio"
	"github.com/nikarpoff/audio-streamer/internal/config"
)

type Capture struct {
	stream *portaudio.Stream
	config *config.AudioConfig
	Buffer chan []int16
	ctx    context.Context
	cancel context.CancelFunc
}

func NewCapture(cfg *config.AudioConfig) (*Capture, error) {
	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize portaudio: %w", err)
	}

	buffer := make(chan []int16, 100) // Bufferized audiodata channel

	// Create capture stream
	stream, err := portaudio.OpenDefaultStream(
		1, // number of input channels
		0, // number of output channels
		cfg.SampleRate,
		cfg.BufferSize,
		func(in []int16) {
			// Copy data, cause channel can be reused
			data := make([]int16, len(in))
			copy(data, in)
			select {
			case buffer <- data:
				// Data was succesfully sent!
			default:
				// Skip data (backpressure)
				log.Println("Capture buffer full, dropping audio data")
			}
		},
	)
	if err != nil {
		portaudio.Terminate()
		return nil, fmt.Errorf("failed to open audio stream: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Capture{
		stream: stream,
		config: cfg,
		Buffer: buffer,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Tries to start capturing audio
func (c *Capture) Start() error {
	if err := c.stream.Start(); err != nil {
		return fmt.Errorf("failed to start capture stream: %w", err)
	}

	log.Printf("Audio capture started: %.0f Hz, %d channels, buffer: %d\n",
		c.config.SampleRate, c.config.Channels, c.config.BufferSize)

	return nil
}

// Stops and closes portaudio stream, closes buffers and terminates portaudio
func (c *Capture) Stop() error {
	if c.stream != nil {
		if err := c.stream.Stop(); err != nil {
			return err
		}
		if err := c.stream.Close(); err != nil {
			return err
		}
	}

	portaudio.Terminate()
	c.cancel()
	close(c.Buffer)

	log.Println("Audio capture stopped")
	return nil
}

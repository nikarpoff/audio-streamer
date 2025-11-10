package audio

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/ebitengine/oto/v3"
	"github.com/nikarpoff/audio-streamer/internal/config"
)

type Playback struct {
	context *oto.Context
	player  *oto.Player
	config  *config.AudioConfig
	Buffer  chan []int16
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewPlayback(cfg *config.AudioConfig) (*Playback, error) {
	// Create Oto Context
	op := &oto.NewContextOptions{
		SampleRate:   int(cfg.SampleRate),
		ChannelCount: cfg.Channels,
		Format:       oto.FormatSignedInt16LE,
	}

	otoCtx, ready, err := oto.NewContext(op)
	if err != nil {
		return nil, fmt.Errorf("failed to create oto context: %w", err)
	}
	<-ready // wait for ready!

	ctx, cancel := context.WithCancel(context.Background())
	buffer := make(chan []int16, 100)

	pb := &Playback{
		context: otoCtx,
		config:  cfg,
		Buffer:  buffer,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Create reader that reads data from buffer into io.Pipe
	reader := pb.createReader()

	// Create player with reader
	player := otoCtx.NewPlayer(reader)
	player.Play()

	pb.player = player

	log.Printf("Audio playback started: %.0f Hz, %d channels\n",
		cfg.SampleRate, cfg.Channels)

	return pb, nil
}

// createReader creates io.Reader, that reads data from buffer and sends them into io.Pipe
func (p *Playback) createReader() io.Reader {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		for {
			select {
			case data, ok := <-p.Buffer:
				if !ok {
					return
				}
				// Convert int16 into bytes (little-endian)
				byteData := make([]byte, len(data)*2)
				for i, sample := range data {
					byteData[i*2] = byte(sample)
					byteData[i*2+1] = byte(sample >> 8)
				}

				// Write data into Pipe
				if _, err := pw.Write(byteData); err != nil {
					if err != io.ErrClosedPipe {
						log.Printf("Pipe write error: %v", err)
					}
					return
				}

			case <-p.ctx.Done():
				return
			}
		}
	}()

	return pr
}

func (p *Playback) Write(data []int16) {
	select {
	case p.Buffer <- data:
		// Data was sent!
	default:
		// Drop audio 'cause channel is full
		log.Println("Playback buffer full, dropping audio data")
	}
}

func (p *Playback) Stop() {
	p.cancel()
	close(p.Buffer)
	log.Println("Audio playback stopped")
}

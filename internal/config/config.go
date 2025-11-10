package config

type AudioConfig struct {
	SampleRate float64
	Channels   int
	BufferSize int
	BitDepth   int
}

func DefaultConfig() *AudioConfig {
	return &AudioConfig{
		SampleRate: 44100,
		Channels:   1, // use mono for quick work
		BufferSize: 1024,
		BitDepth:   16,
	}
}

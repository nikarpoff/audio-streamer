package config

type AudioConfig struct {
	InputChannels  int
	OutputChannels int
	SampleRate     float64
	BufferSize     int
	BitDepth       int
}

func DefaultConfig() *AudioConfig {
	return &AudioConfig{
		InputChannels:  2,
		OutputChannels: 2,
		SampleRate:     48000,
		BufferSize:     512,
		BitDepth:       16,
	}
}

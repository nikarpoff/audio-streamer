package config

import (
	"os"
	"strings"
)

const (
	defaultServerAddress = "127.0.0.1:8080"
	serverEnvKey         = "AUDIO_STREAMER_SERVER_ADDR"
)

// DefaultServerAddress returns server address from process environment.
func DefaultServerAddress() string {
	if value, ok := os.LookupEnv(serverEnvKey); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}

	return defaultServerAddress
}

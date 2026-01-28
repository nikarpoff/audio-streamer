package utils

import (
	"net/url"
)

func IsValidWebSocketURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Check schema (ws:// or wss://)
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return false
	}

	// Check host
	if u.Host == "" {
		return false
	}

	return true
}

package utils

import (
	"net/url"
)

func IsValidServerAddress(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Check schema
	if u.Scheme == "" {
		return u.Host != "" || u.Path != ""
	}

	// Check host
	if u.Host == "" {
		return false
	}

	return u.Host != ""
}

func IsValidBoundedInteger(value int, minValue int, maxValue int) bool {
	if value < minValue || value > maxValue {
		return false
	}

	return true
}

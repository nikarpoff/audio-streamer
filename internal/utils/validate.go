package utils

func IsValidBoundedInteger(value int, minValue int, maxValue int) bool {
	if value < minValue || value > maxValue {
		return false
	}

	return true
}

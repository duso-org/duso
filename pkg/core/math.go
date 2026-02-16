package core

// IsInteger checks if a float64 is an integer value
func IsInteger(n float64) bool {
	return n == float64(int64(n))
}

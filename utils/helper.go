package utils

// FirstNonEmpty returns the first non-empty string from the provided values.
// If all values are empty, it returns an empty string.
func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

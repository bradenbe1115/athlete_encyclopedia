package utils

import (
	"fmt"
	"os"
)

// ReadTextFromFile reads text content from a file and returns it as a string.
func ReadTextFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(data), nil
}

package utils

import (
	"fmt"
	"os"
)

// GetConnURI returns env var value of CONN_URI or an error if var is not set
func GetConnURI() (string, error) {
	connURI := os.Getenv("CONN_URI")
	if connURI == "" {
		return "", fmt.Errorf("CONN_URI env var not set")
	}
	return connURI, nil
}

// GetEncHome returns env var value of ENC_HOME or an error if var is not set
func GetEncHome() (string, error) {
	encHome := os.Getenv("ENC_HOME")
	if encHome == "" {
		return "", fmt.Errorf("ENC_HOME env var not set")
	}
	return encHome, nil
}

// GetJobFilePath returns env var value of JOB_FILE_PATH or an error if var is not set
func GetJobFilePath() (string, error) {
	jobFilePath := os.Getenv("JOB_FILE_PATH")
	if jobFilePath == "" {
		return "", fmt.Errorf("JOB_FILE_PATH env var not set")
	}
	return jobFilePath, nil
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var ErrInvalidConfig = errors.New("invalid config format")

// validateConfig validates that config contains expected fields.
func validateConfig(cfg *JobConfig) error {
	if cfg.DestTableName == "" {
		return ErrInvalidConfig
	}

	if cfg.InputType == "" {
		return ErrInvalidConfig
	}

	if cfg.LoadMethod != "append" && cfg.LoadMethod != "" {
		return ErrInvalidConfig
	}

	return nil
}

// ReadConfigFromFile reads a JSON config file and returns the JobConfig.
func ReadConfigFromFile(filePath string) (*JobConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config JobConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, ErrInvalidConfig
	}

	err = validateConfig(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}


// loadParamsFromConfig creates LoadParams from JobConfig struct
func loadParamsFromConfig(cfg JobConfig) *LoadParams {

	params := LoadParams{
		StagingTableName: fmt.Sprintf("stg_%s", cfg.DestTableName),
		DestTableName:    cfg.DestTableName,
		RawFilePath:      os.Getenv("RawFilePath"),
		Query:            "",
		Mode:             "append",
	}
	return &params
}

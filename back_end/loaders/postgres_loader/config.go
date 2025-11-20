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

	if cfg.LoadMethod != "append" {
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

// loadQueryTextFromFile loads query text from a file.
func loadQueryTextFromFile(queryFilePath string) (string, error) {
	queryBytes, err := os.ReadFile("queries/my_query.sql")
	if err != nil {
		return "", fmt.Errorf("failed to read sql file: %v", err)
	}
	query := string(queryBytes)

	return query, nil
}

// LoadParamsFromConfig creates LoadParams from JobConfig struct
func LoadParamsFromConfig(cfg JobConfig) (*LoadParams, error) {
	importId := os.Getenv("IMPORT_ID")
	if importId == "" {
		return nil, fmt.Errorf("env var IMPORT_ID not set")
	}

	params := LoadParams{
		ImportID:         importId,
		StagingTableName: fmt.Sprintf("stg_%s", cfg.DestTableName),
		DestTableName:    cfg.DestTableName,
		RawFilePath:      os.Getenv("RawFilePath"),
		Query:            "",
		Mode:             "append",
	}
	return &params, nil
}

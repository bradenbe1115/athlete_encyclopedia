package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadConfigFromFile reads a JSON config file and returns the JobConfig.
// This function implements the ConfigReader type defined in commands.go.
func ReadConfigFromFile(filePath string) (*JobConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config JobConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
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

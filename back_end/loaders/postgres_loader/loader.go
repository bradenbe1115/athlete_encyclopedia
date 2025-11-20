package main

import (
	"context"
	"fmt"
	"log/slog"
)

type JobConfig struct {
	InputType     string `json:"inputType"`
	DestTableName string `json:"destTableName"`
	LoadMethod    string `json:"loadMethod,omitempty"`
}

// LoadParams holds all possible parameters for different loader types
type LoadParams struct {
	ImportID         string
	StagingTableName string
	DestTableName    string

	// JSON loader specific
	RawFilePath string

	// SQL loader specific
	Query string
	Mode  string
}

// Loader interface that all loaders implement
type Loader interface {
	Load(ctx context.Context, params LoadParams) error
}

// LoaderFactory creates the appropriate loader based on input type
type LoaderFactory interface {
	CreateLoader(ctx context.Context, inputType string) (Loader, error)
}

// ConfigReader reads job config from a file path
// specified by the input string.
type ConfigReader func(string) (*JobConfig, error)

type PostgresLoader struct {
	Logger        *slog.Logger
	ConfigReader  ConfigReader
	LoaderFactory LoaderFactory
}

// New creates a new PostgresLoader
func New(ctx context.Context, configReader ConfigReader, connURI string) *PostgresLoader {
	logger := slog.Default()
	return &PostgresLoader{Logger: logger, ConfigReader: configReader, LoaderFactory: &PostgresLoaderFactory{ConnURI: connURI}}
}

// loadData reads the config and loads the data to Postgres.
func (l *PostgresLoader) LoadData(ctx context.Context, jobFilePath string) error {
	configFilePath := jobFilePath + "/config.json"
	config, err := l.ConfigReader(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse job config: %w", err)
	}

	loader, err := l.LoaderFactory.CreateLoader(ctx, config.InputType)
	if err != nil {
		return fmt.Errorf("failed to create loader: %w", err)
	}

	params, err := LoadParamsFromConfig(*config)
	if err != nil {
		return fmt.Errorf("failed to create load params: %w", err)
	}

	if config.InputType == "sql" {
		queryFilePath := jobFilePath + "/query.sql"
		query, err := loadQueryTextFromFile(queryFilePath)
		if err != nil {
			return fmt.Errorf("failed to load query: %w", err)
		}
		params.Query = query
	}
	return loader.Load(ctx, *params)
}

package main

import (
	"context"
	"fmt"
	"log/slog"
)

type PostgresLoaderFactory struct {
	ConnURI string
}

// PostgresLoaderFactory implements LoaderFactory. Returns the appropriate Loader based on
// inputType
func (p *PostgresLoaderFactory) CreateLoader(ctx context.Context, inputType string) (Loader, error) {
	logger := slog.Default()
	switch inputType {
	case "json":
		c := DuckDBPostgresConnector{ConnURI: p.ConnURI}
		db, err := c.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to db: %w", err)
		}
		return &DuckDBJSONPostgresLoader{DB: db, Logger: logger}, nil
	case "sql":
		{
			c := PostgresConnector{ConnURI: p.ConnURI}
			db, err := c.Connect(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to db: %w", err)
			}
			return &PostgresSQLLoader{DB: db, Logger: logger}, nil
		}
	default:
		return nil, fmt.Errorf("non-implemented option: %s", inputType)
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()
	logger := slog.Default()
	pgConnURI := os.Getenv("PGConnURI")
	destTableName := os.Getenv("DestTableName")
	stagingTableName := fmt.Sprintf(`stg_%s`, destTableName)
	rawFilePath := os.Getenv("RawFilePath")

	c := DuckDBPostgresConnector{PGConnURI: pgConnURI}
	db, err := c.Connect(ctx)
	if err != nil {
		log.Fatalf("failed to make database connection: %v", err)
	}

	loader := DuckDBJSONPostgresLoader{DB: db, Logger: logger}
	err = loader.Run(ctx, rawFilePath, stagingTableName, destTableName)
	if err != nil {
		log.Fatalf("loader failed: %v", err)
	}
}

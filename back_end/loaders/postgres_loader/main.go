package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var ErrUnexpectedFileFormat = errors.New("file name not in expected format")

func ExtractImportIdFromFileName(rawFilePath string) (string, error) {
	base := filepath.Base(rawFilePath)

	// Remove the extension
	name := base[:len(base)-len(filepath.Ext(base))]
	if len(name) > 18 {
		return strings.ReplaceAll(name[len(name)-18:], "-", "/"), nil

	} else {
		return "", ErrUnexpectedFileFormat
	}
}

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

	importId, err := ExtractImportIdFromFileName(rawFilePath)
	if err != nil {
		log.Fatalf("failed to extract ImportId from file name: %v", err)
	}

	loader := DuckDBJSONPostgresLoader{DB: db, Logger: logger}
	err = loader.Run(ctx, importId, rawFilePath, stagingTableName, destTableName)
	if err != nil {
		log.Fatalf("loader failed: %v", err)
	}
}

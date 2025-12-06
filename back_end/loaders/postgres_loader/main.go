package main

import (
	"context"
	"loaders/utils"
	"log"
)

func main() {
	ctx := context.Background()
	connURI, err := utils.GetConnURI()
	if err != nil {
		log.Fatalf("connection URI not found: %v", err)
	}
	encHome, err := utils.GetEncHome()
	if err != nil {
		log.Fatalf("project home not found: %v", err)
	}
	jobFilePath, err := utils.GetJobFilePath()
	if err != nil {
		log.Fatalf("job file path not found: %v", err)
	}
	fullJobFilePath := encHome + "/" + jobFilePath
	l := New(ctx, ReadConfigFromFile, connURI)
	err = l.LoadData(ctx, fullJobFilePath)
	if err != nil {
		log.Fatalf("error loading data: %v", err)
	}
}

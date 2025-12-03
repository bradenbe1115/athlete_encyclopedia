package main

import (
	"context"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	connURI := os.Getenv("CONN_URI")
	encHome := os.Getenv("ENC_HOME")
	jobFilePath := os.Getenv("JOB_FILE_PATH")
	fullJobFilePath := encHome + "/" + jobFilePath
	l := New(ctx, ReadConfigFromFile, connURI)
	err := l.LoadData(ctx, fullJobFilePath)
	if err != nil {
		log.Fatalf("error loading data: %v", err)
	}
}

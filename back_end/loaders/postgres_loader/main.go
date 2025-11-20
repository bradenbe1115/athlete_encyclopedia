package main

import (
	"context"
	"os"
)

func main() {
	ctx := context.Background()
	connURI := os.Getenv("CONN_URI")
	jobFilePath := os.Getenv("JOB_FILE_PATH")
	l := New(ctx, ReadConfigFromFile, connURI)
	l.LoadData(ctx, jobFilePath)
}

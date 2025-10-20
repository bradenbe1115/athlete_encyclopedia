package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	router := gin.Default()

	err := godotenv.Load(`../../../../.env`)
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	p := PostgresConnector{
		Database: os.Getenv("POSTGRES_DB"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
	}

	h, err := New(ctx, &p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	router.GET("/athletes", h.GetAthletes)

	router.Run("localhost:8080")
}

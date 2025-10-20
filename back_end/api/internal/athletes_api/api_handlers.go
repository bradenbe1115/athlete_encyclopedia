package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Athlete struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	FullName  string    `db:"full_name"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
}

// DBConnector can read data from the database.
type DBConnector interface {
	Connect(context.Context) (*sql.DB, error)
}

// DBClient provides methods for reading data from a DB.
type DBClient interface {
	FetchAthletes(context.Context, string, []interface{}) ([]Athlete, error)
}

// APIHandler provides methods to handle API requests
type APIHandler struct {
	DB     DBClient
	Logger *slog.Logger
}

// New connects to the database using the DBConnector, and returns a *APIHandler
// with an appropriate default configuration.
func New(ctx context.Context, dbConnector DBConnector) (*APIHandler, error) {
	logger := slog.Default()
	db, err := dbConnector.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &APIHandler{
		DB:     &PostgresClient{db, logger},
		Logger: logger,
	}, nil
}

type GetAthletesParams struct {
	FullName *string
	Team     *string
}

// GetAthletes creates HTTP response for fetch athletes.
func (h *APIHandler) GetAthletes(c *gin.Context) {
	fullname := c.Query("fullname")
	team := c.Query("team")
	query, args := BuildFetchAthletesQuery(GetAthletesParams{FullName: &fullname, Team: &team})
	athletes, err := h.DB.FetchAthletes(c, query, args)
	if err != nil {
		h.Logger.Error("failed to fetch athletes", "error", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch athletes."})
		return
	}
	c.IndentedJSON(http.StatusOK, athletes)
}

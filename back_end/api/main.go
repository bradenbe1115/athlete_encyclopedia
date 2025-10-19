package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	logger := slog.Default()
	router := gin.Default()

	err := godotenv.Load(`../../.env`)
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

	db, err := p.Connect(ctx)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	pc := &PostgresClient{db, logger}
	h := &APIHandler{DB: pc, Logger: logger}

	router.GET("/athletes", h.GetAthletes)

	router.Run("localhost:8080")
}

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

type GetAthletesParams struct {
	FullName *string
	Team     *string
}

// BuildFetchAthletesQuery builds the templated query and args
// to fetch athletes based off request params
func BuildFetchAthletesQuery(p GetAthletesParams) (q string, a []interface{}) {
	var conditions []string
	var args []interface{}
	paramCount := 1
	if p.FullName != nil && *p.FullName != "" {
		conditions = append(conditions, "full_name = $"+strconv.Itoa(paramCount))
		args = append(args, *p.FullName)
		paramCount++
	}

	if p.Team != nil && *p.Team != "" {
		conditions = append(conditions, "team = $"+strconv.Itoa(paramCount))
		args = append(args, *p.Team)
		paramCount++
	}

	query := "select * from athletes"

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " and ")
	}
	return query, args
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

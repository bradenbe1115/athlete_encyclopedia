package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

var (
	ErrUnexpectedQueryResultsSchema = errors.New("unexpected query result schema")
)

// PostgresConnector implements DBConnector
type PostgresConnector struct {
	Database string
	User     string
	Password string
	Host     string
	Port     string
}

// Connect to a postgres database.
func (p *PostgresConnector) Connect(ctx context.Context) (*sql.DB, error) {
	connStr := fmt.Sprintf(`user=%s dbname=%s password=%s host=%s port=%s sslmode=disable`, p.User, p.Database, p.Password, p.Host, p.Port)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create Postgres connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Postgres: %w", err)
	}

	return db, nil

}

// PostgresClient holds a connection to postgres
type PostgresClient struct {
	DB     *sql.DB
	Logger *slog.Logger
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

	query := `select 
						a.*
					from athletes a
					LEFT JOIN athletes_teams_relations atr ON atr.athlete_id = a.id
					LEFT JOIN teams t ON t.id = atr.team_id`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " and ")
	}
	return query, args
}

// getAthlete responds with the list of all athletes as JSON.
func (pc *PostgresClient) FetchAthletes(ctx context.Context, query string, args []interface{}) ([]Athlete, error) {

	rows, err := pc.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("counld not retrieve data from db: %w", err)
	}
	defer rows.Close()

	var athletes []Athlete
	for rows.Next() {
		var athlete Athlete
		err := rows.Scan(&athlete.ID, &athlete.CreatedAt, &athlete.FullName, &athlete.FirstName, &athlete.LastName)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedQueryResultsSchema, err)
		}
		athletes = append(athletes, athlete)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return athletes, nil
}

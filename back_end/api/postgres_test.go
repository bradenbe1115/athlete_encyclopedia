package main

import (
	"context"
	"database/sql/driver"
	"errors"
	"log/slog"
	"regexp"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"

	"testing"
)

func toDriverValues(args []interface{}) []driver.Value {
	vals := make([]driver.Value, len(args))
	for i, v := range args {
		vals[i] = driver.Value(v)
	}
	return vals
}

func TestBuildFetchAthletesQuery(t *testing.T) {
	tt := []struct {
		desc          string
		team          string
		fullName      string
		expectedQuery string
		expectedArgs  []interface{}
	}{
		{
			desc:          "No params",
			team:          "",
			fullName:      "",
			expectedQuery: `select * from athletes`,
			expectedArgs:  nil,
		},
		{
			desc:          "Full Name param only",
			team:          "",
			fullName:      "Test",
			expectedQuery: `select * from athletes WHERE fullname = $1`,
			expectedArgs:  []interface{}{"Test"},
		},
		{
			desc:          "Team param only",
			team:          "Test Team",
			fullName:      "",
			expectedQuery: `select * from athletes WHERE team = $1`,
			expectedArgs:  []interface{}{"Test Team"},
		},
		{
			desc:          "Team and full name params",
			team:          "Test Team",
			fullName:      "Test",
			expectedQuery: `select * from athletes WHERE fullname = $1 and team = $2`,
			expectedArgs:  []interface{}{"Test", "Test Team"},
		},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			p := GetAthletesParams{
				FullName: &test.fullName,
				Team:     &test.team,
			}

			query, args := BuildFetchAthletesQuery(p)
			matched, err := regexp.MatchString(regexp.QuoteMeta(test.expectedQuery), query)
			if err != nil {
				t.Fatalf("invalid regex pattern: %v", err)
			}
			if !matched {
				t.Errorf("query does not match expected pattern.\nPattern: %s\nActual:  %s", test.expectedQuery, query)
			}

			if diff := cmp.Diff(test.expectedArgs, args); diff != "" {
				t.Errorf("expected -, actual +:\n%s", diff)
			}
		})
	}
}

func TestFetchAthletes(t *testing.T) {
	var fetchAthletesColumns = []string{"id", "createdat", "fullname", "firstname", "lastname"}
	parsedTime, err := time.Parse("2006-01-02 15:04:05.000 -0700", "2020-05-25 23:30:31.198 +0000")
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}
	tt := []struct {
		desc          string
		query         string
		args          []interface{}
		newRows       *sqlmock.Rows
		expected      []Athlete
		expectedError error
	}{
		{
			desc:  "No params successful fetch",
			query: `SELECT .+ FROM athletes`,
			args:  []interface{}{},
			newRows: sqlmock.NewRows(fetchAthletesColumns).
				AddRow("1", parsedTime, "The Athlete", "The", "Athlete"),
			expected: []Athlete{
				{
					ID:        "1",
					CreatedAt: parsedTime,
					FullName:  "The Athlete",
					FirstName: "The",
					LastName:  "Athlete",
				},
			},
			expectedError: nil,
		},
		{
			desc:  "One param successful fetch",
			query: `SELECT .+ FROM athletes WHERE fullname = ?`,
			args:  []interface{}{"The Athlete"},
			newRows: sqlmock.NewRows(fetchAthletesColumns).
				AddRow("1", parsedTime, "The Athlete", "The", "Athlete"),
			expected: []Athlete{
				{
					ID:        "1",
					CreatedAt: parsedTime,
					FullName:  "The Athlete",
					FirstName: "The",
					LastName:  "Athlete",
				},
			},
			expectedError: nil,
		},
		{
			desc:  "Multiple params successful fetch",
			query: `SELECT .+ FROM athletes WHERE fullname = ? AND team = ?`,
			args:  []interface{}{"The Athlete", "Eagles"},
			newRows: sqlmock.NewRows(fetchAthletesColumns).
				AddRow("1", parsedTime, "The Athlete", "The", "Athlete"),
			expected: []Athlete{
				{
					ID:        "1",
					CreatedAt: parsedTime,
					FullName:  "The Athlete",
					FirstName: "The",
					LastName:  "Athlete",
				},
			},
			expectedError: nil,
		},
		{
			desc:  "Unexpected query result leads to error",
			query: `SELECT .+ FROM athletes;`,
			args:  []interface{}{},
			newRows: sqlmock.NewRows(fetchAthletesColumns).
				AddRow(nil, parsedTime, "The Athlete", "The", "Athlete"),
			expected:      nil,
			expectedError: ErrUnexpectedQueryResultsSchema,
		},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mock.ExpectQuery(regexp.QuoteMeta(test.query)).WithArgs(toDriverValues(test.args)...).WillReturnRows(test.newRows)

			client := PostgresClient{DB: db, Logger: slog.New(slog.DiscardHandler)}

			athletes, err := client.FetchAthletes(context.Background(), test.query, test.args)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.expected, athletes); diff != "" {
				t.Errorf("expected -, actual +:\n%s", diff)
			}
		})
	}
}

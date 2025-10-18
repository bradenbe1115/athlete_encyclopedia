package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"

	"testing"
)

func TestFetchAthletes(t *testing.T) {
	var fetchAthletesColumns = []string{"id", "createdat", "fullname", "firstname", "lastname"}
	parsedTime, err := time.Parse("2006-01-02 15:04:05.000 -0700", "2020-05-25 23:30:31.198 +0000")
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}
	tt := []struct {
		desc          string
		expectedQuery string
		newRows       *sqlmock.Rows
		expected      []Athlete
		expectedError error
	}{
		{
			desc:          "Successful Fetch",
			expectedQuery: `SELECT .+ FROM athletes;`,
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
			desc:          "Successful Fetch",
			expectedQuery: `SELECT .+ FROM athletes;`,
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

			mock.ExpectQuery(test.expectedQuery).WillReturnRows(test.newRows)

			client := PostgresClient{DB: db, Logger: slog.New(slog.DiscardHandler)}

			athletes, err := client.FetchAthletes(context.Background())
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.expected, athletes); diff != "" {
				t.Errorf("expected -, actual +:\n%s", diff)
			}
		})
	}
}

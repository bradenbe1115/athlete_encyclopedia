package main

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateStagingTable(t *testing.T) {
	tt := []struct {
		desc             string
		query            string
		stagingTableName string
		expectedQuery    string
		expectedError    error
	}{
		{
			desc:             "Succesfully call create temporary table",
			query:            `SELECT name, date from table_of_fools`,
			stagingTableName: `new_table_of_fools`,
			expectedQuery: `CREATE TEMPORARY TABLE new_table_of_fools AS 
							SELECT name, date from table_of_fools`,
			expectedError: nil,
		},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mock.ExpectExec(regexp.QuoteMeta(test.expectedQuery)).WillReturnResult(sqlmock.NewResult(1, 1))
			loader := PostgresLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			err = loader.CreateStagingTable(context.Background(), test.query, test.stagingTableName)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddImportIdColumn(t *testing.T) {
	tt := []struct {
		desc          string
		table         string
		importId      string
		expectedQuery string
		expectedError error
	}{
		{
			desc:          "Successful call",
			table:         "a_fake_table",
			importId:      "all_data_file",
			expectedQuery: `ALTER TABLE a_fake_table ADD COLUMN _import_id VARCHAR DEFAULT 'all_data_file'`,
			expectedError: nil,
		},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mock.ExpectExec(regexp.QuoteMeta(test.expectedQuery)).WillReturnResult(sqlmock.NewResult(1, 1))
			loader := PostgresLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			err = loader.AddImportIDColumn(context.Background(), test.table, test.importId)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetColumnNamesFromTable(t *testing.T) {
	tt := []struct {
		desc            string
		table           string
		expectedQuery   string
		expectedColumns []string
		expectedError   error
	}{
		{
			desc:  "Successfully retrieve column names",
			table: "players",
			expectedQuery: `
	SELECT
		column_name
	FROM information_schema.columns where table_name = 'players'`,
			expectedColumns: []string{"id", "name", "team", "_import_id"},
			expectedError:   nil,
		},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			rows := sqlmock.NewRows([]string{"column_name"})
			for _, col := range test.expectedColumns {
				rows.AddRow(col)
			}

			mock.ExpectQuery(regexp.QuoteMeta(test.expectedQuery)).WillReturnRows(rows)
			loader := PostgresLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			columns, err := loader.GetColumnNamesFromTable(context.Background(), test.table)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}

			if len(columns) != len(test.expectedColumns) {
				t.Errorf("expected %d columns, got %d", len(test.expectedColumns), len(columns))
			}

			for i, col := range columns {
				if col != test.expectedColumns[i] {
					t.Errorf("expected column %s at position %d, got %s", test.expectedColumns[i], i, col)
				}
			}
		})
	}
}

func TestDeleteFromTable(t *testing.T) {
	tt := []struct {
		desc          string
		tableName     string
		importId      string
		expectedQuery string
		expectedError error
	}{
		{
			desc:      "Successfully delete data from table",
			tableName: "players",
			importId:  "batch_2024_01",
			expectedQuery: `
	DELETE FROM players WHERE import_id = batch_2024_01`,
			expectedError: nil,
		},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mock.ExpectExec(regexp.QuoteMeta(test.expectedQuery)).WillReturnResult(sqlmock.NewResult(1, 1))
			loader := PostgresLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			err = loader.DeleteFromTable(context.Background(), test.tableName, test.importId)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAppendStagingData(t *testing.T) {
	tt := []struct {
		desc             string
		stagingTableName string
		destTableName    string
		colNames         []string
		expectedQuery    string
		expectedError    error
	}{
		{
			desc:             "Successfully append data from staging to destination",
			stagingTableName: "staging_players",
			destTableName:    "players",
			colNames:         []string{"id", "name", "team", "_import_id"},
			expectedQuery: `
	INSERT INTO players (id, name, team, _import_id)
	SELECT
	id, name, team, _import_id
	FROM staging_players`,
			expectedError: nil,
		},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mock.ExpectExec(regexp.QuoteMeta(test.expectedQuery)).WillReturnResult(sqlmock.NewResult(1, 1))
			loader := PostgresLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			err = loader.AppendStagingData(context.Background(), test.stagingTableName, test.destTableName, test.colNames)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

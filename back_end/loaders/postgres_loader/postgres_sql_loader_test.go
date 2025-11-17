package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
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
			loader := PostgresSQLLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			err = loader.createStagingTable(context.Background(), test.query, test.stagingTableName)
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
			loader := PostgresSQLLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			err = loader.addImportIDColumn(context.Background(), test.table, test.importId)
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
			loader := PostgresSQLLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			columns, err := loader.getColumnNamesFromTable(context.Background(), test.table)
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
			loader := PostgresSQLLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			err = loader.deleteFromTable(context.Background(), test.tableName, test.importId)
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
			loader := PostgresSQLLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			err = loader.appendStagingData(context.Background(), test.stagingTableName, test.destTableName, test.colNames)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// setupLoadMockExpectations is a helper function to set up mock expectations for Load method
func setupLoadMockExpectations(mock sqlmock.Sqlmock, importId, query, mode, stagingTableName, destTableName string,
	destTableColumns []string) {
	// Expect CreateStagingTable
	mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(
		"CREATE TEMPORARY TABLE %s AS %s",
		stagingTableName, query))).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect AddImportIDColumn
	mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(
		"ALTER TABLE %s ADD COLUMN _import_id VARCHAR DEFAULT '%s'",
		stagingTableName, importId))).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect GetColumnNamesFromTable
	rows := sqlmock.NewRows([]string{"column_name"})
	for _, col := range destTableColumns {
		rows.AddRow(col)
	}
	mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf(
		"SELECT\n\t\tcolumn_name\n\tFROM information_schema.columns where table_name = '%s'",
		destTableName))).
		WillReturnRows(rows)

	// Expect DeleteFromTable (only in append mode)
	if mode == "append" {
		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(
			"DELETE FROM %s WHERE import_id = %s",
			destTableName, importId))).
			WillReturnResult(sqlmock.NewResult(1, 5))

		// Expect AppendStagingData
		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(
			"INSERT INTO %s (%s)\n\tSELECT\n\t%s\n\tFROM %s",
			destTableName,
			strings.Join(destTableColumns, ", "),
			strings.Join(destTableColumns, ", "),
			stagingTableName))).
			WillReturnResult(sqlmock.NewResult(1, 5))

		// Expect DropStagingTable
		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(
			"DROP TABLE IF EXISTS %s",
			stagingTableName))).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
}

func TestLoad(t *testing.T) {
	tt := []struct {
		desc             string
		importId         string
		query            string
		mode             string
		stagingTableName string
		destTableName    string
		destTableColumns []string
		expectedError    error
	}{
		{
			desc:             "Successfully load data in append mode",
			importId:         "2025/11/11/11",
			query:            "SELECT id, name from source_table",
			mode:             "append",
			stagingTableName: "staging_table",
			destTableName:    "dest_table",
			destTableColumns: []string{"id", "name"},
			expectedError:    nil,
		},
		{
			desc:             "Successfully load data in append mode",
			importId:         "2025/11/11/11",
			query:            "SELECT id, name from source_table",
			mode:             "append",
			stagingTableName: "staging_table",
			destTableName:    "dest_table",
			destTableColumns: []string{"id", "name"},
			expectedError:    nil,
		},
	}
	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			setupLoadMockExpectations(mock, test.importId,
				test.query, test.mode, test.stagingTableName, test.destTableName,
				test.destTableColumns)

			loader := PostgresSQLLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			err = loader.load(context.Background(), test.importId, test.query, test.mode, test.stagingTableName, test.destTableName)

			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

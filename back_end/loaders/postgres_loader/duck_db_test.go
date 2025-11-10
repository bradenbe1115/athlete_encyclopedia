package main

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
)

func TestMapDuckDBTypeToPostgres(t *testing.T) {
	tt := []struct {
		desc     string
		duckType string
		expected string
	}{
		{
			desc:     "varchar",
			duckType: "VARCHAR",
			expected: "TEXT",
		},

		{
			desc:     "default",
			duckType: "UNKNOWN",
			expected: "TEXT",
		},
	}
	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			result := MapDuckDBTypeToPostgres(test.duckType)
			if diff := cmp.Diff(test.expected, result); diff != "" {
				t.Errorf("expected -, actual +:\n%s", diff)
			}
		})
	}
}

func TestGetDuckDBTableSchema(t *testing.T) {
	tableInfoColumns := []string{"cid", "name", "type", "not_null", "default_value", "pk"}
	badTableInfoColumns := []string{"cid"}
	tt := []struct {
		desc          string
		tableName     string
		expectedQuery string
		newRows       *sqlmock.Rows
		expected      []DBCol
		expectedError error
	}{
		{
			desc:          "Valid response",
			tableName:     "a_table",
			expectedQuery: "PRAGMA table_info('a_table')",
			newRows: sqlmock.NewRows(tableInfoColumns).
				AddRow(1, "col_One", "VARCHAR", false, "", false).
				AddRow(2, "col two", "FLOAT", false, "", false),
			expected: []DBCol{
				{
					Name:           "col_One",
					DuckType:       "VARCHAR",
					PostgresType:   "TEXT",
					NormalizedName: "col_one",
				},
				{
					Name:           "col two",
					DuckType:       "FLOAT",
					PostgresType:   "DOUBLE PRECISION",
					NormalizedName: "col_two",
				},
			},
			expectedError: nil,
		},
		{
			desc:          "Invalid response",
			tableName:     "a_table",
			expectedQuery: "PRAGMA table_info('a_table')",
			newRows: sqlmock.NewRows(badTableInfoColumns).
				AddRow(1).
				AddRow(2),
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

			mock.ExpectQuery(regexp.QuoteMeta(test.expectedQuery)).
				WillReturnRows(test.newRows)

			loader := DuckDBJSONPostgresLoader{DB: db, Logger: slog.New(slog.DiscardHandler)}
			result, err := loader.GetDuckDBTableSchema(context.Background(), test.tableName)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.expected, result); diff != "" {
				t.Errorf("expected -, actual +:\n%s", diff)
			}
		})
	}
}

func TestCreateDestTable(t *testing.T) {
	tt := []struct {
		desc          string
		destTableName string
		dc            []DBCol
		expectedQuery string
		expectedError error
	}{
		{
			desc:          "Single column schema",
			destTableName: "dest_table",
			dc: []DBCol{
				{
					Name:         "col_one",
					DuckType:     "VARCHAR",
					PostgresType: "TEXT",
				},
			},
			expectedQuery: "CREATE TABLE IF NOT EXISTS dest_table (col_one TEXT)",
			expectedError: nil,
		},
		{
			desc:          "Multi column schema",
			destTableName: "dest_table",
			dc: []DBCol{
				{
					Name:         "col_one",
					DuckType:     "VARCHAR",
					PostgresType: "TEXT",
				},
				{
					Name:         "col_two",
					DuckType:     "INTEGER",
					PostgresType: "BIGINT",
				},
			},
			expectedQuery: "CREATE TABLE IF NOT EXISTS dest_table (col_one TEXT, col_two BIGINT)",
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

			mock.ExpectQuery(regexp.QuoteMeta(test.expectedQuery))
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoadFromStagingtoPostgres(t *testing.T) {
	tt := []struct {
		desc          string
		stagingTable  string
		destTableName string
		expectedQuery string
		dc            []DBCol
		expectedError error
	}{
		{
			desc:          "Single column schema",
			stagingTable:  "a_stg_table",
			destTableName: "a_dst_table",
			expectedQuery: `INSERT INTO a_dst_table
							SELECT
								col_one
							FROM a_stg_tables`,
			dc: []DBCol{
				{
					Name:         "col_one",
					DuckType:     "VARCHAR",
					PostgresType: "TEXT",
				},
			},
			expectedError: nil,
		},

		{
			desc:          "Multi column schema",
			stagingTable:  "a_stg_table",
			destTableName: "a_dst_table",
			expectedQuery: `INSERT INTO a_dst_table
							SELECT
								col_one,
								col_two
							FROM a_stg_tables`,
			dc: []DBCol{
				{
					Name:         "col_one",
					DuckType:     "VARCHAR",
					PostgresType: "TEXT",
				},
				{
					Name:         "col_two",
					DuckType:     "INTEGER",
					PostgresType: "BIGINT",
				},
			},
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

			mock.ExpectQuery(regexp.QuoteMeta(test.expectedQuery))
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
)

// PostgresConnector implements DBConnector
type PostgresConnector struct {
	ConnURI string
}

// Connect to a postgres database using a connection URI.
func (p *PostgresConnector) Connect(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("postgres", p.ConnURI)
	if err != nil {
		return nil, fmt.Errorf("failed to create Postgres connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Postgres: %w", err)
	}

	return db, nil
}

// PostgresLoader holds a connection to postgres
type PostgresSQLLoader struct {
	DB     *sql.DB
	Logger *slog.Logger
}

// ExecuteQuery executes a postgres query
func (d *PostgresSQLLoader) ExecuteQuery(ctx context.Context, query string) (sql.Result, error) {
	r, err := d.DB.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}

	return r, nil
}

// Creates a staging table from the query results.
func (d *PostgresSQLLoader) createStagingTable(ctx context.Context, query string, stagingTableName string) error {
	fullQuery := fmt.Sprintf(`CREATE TEMPORARY TABLE %s AS %s`,
		stagingTableName, query)

	d.Logger.With("table", stagingTableName).InfoContext(ctx, "creating temp table")
	_, err := d.ExecuteQuery(ctx, fullQuery)
	if err != nil {
		return fmt.Errorf("failed to create staging table: %w", err)
	}
	d.Logger.With("table", stagingTableName).InfoContext(ctx, "successfully created temp table")
	return nil
}

// AddImportIdColumn appends an import id column to the table and populates the value.
func (d *PostgresSQLLoader) addImportIDColumn(ctx context.Context, table string, importID string) error {
	query := fmt.Sprintf(`
    ALTER TABLE %s ADD COLUMN _import_id VARCHAR DEFAULT '%s'`,
		table, importID)

	_, err := d.ExecuteQuery(ctx, query)
	return err
}

// GetColumnNamesFromTable loads column names from a table into a []string.
func (d *PostgresSQLLoader) getColumnNamesFromTable(ctx context.Context, table string) ([]string, error) {
	query := fmt.Sprintf(`
	SELECT
		column_name
	FROM information_schema.columns where table_name = '%s'`, table)

	rows, err := d.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get table columns: %w", err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return columns, nil
}

// DeleteFromDestTable deletes data from destination table with matching import_id.
func (d *PostgresSQLLoader) deleteFromTable(ctx context.Context, tableName string, import_id string) error {
	query := fmt.Sprintf(`
	DELETE FROM %s WHERE import_id = %s`, tableName, import_id)

	d.Logger.With("table", tableName, "import_id", import_id).InfoContext(ctx, "deleting data from table with import id")
	_, err := d.ExecuteQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete data with import id: %w", err)
	}
	d.Logger.With("table", tableName, "import_id", import_id).InfoContext(ctx, "successfully deleted data from table")

	return nil
}

// AppendStagingData appends data from staging table into destination table.
func (d *PostgresSQLLoader) appendStagingData(ctx context.Context, stagingTableName string, destTableName string, colNames []string) error {
	query := fmt.Sprintf(`
	INSERT INTO %s (%s)
	SELECT
	%s
	FROM %s`, destTableName, strings.Join(colNames, ", "), strings.Join(colNames, ", "), stagingTableName)

	d.Logger.With("staging table", stagingTableName, "dest table", destTableName).InfoContext(ctx, "appending data from staging to destination")
	_, err := d.ExecuteQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to append staging data to destination: %w", err)
	}
	d.Logger.With("staging table", stagingTableName, "dest table", destTableName).InfoContext(ctx, "successfully appended data from staging to destination")
	return nil
}

// DropStagingTable drops table from database.
func (d *PostgresSQLLoader) dropStagingTable(ctx context.Context, stagingTableName string) error {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, stagingTableName)

	d.Logger.With("staging table", stagingTableName).InfoContext(ctx, "dropping staging table")
	_, err := d.ExecuteQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop staging table: %w", err)
	}
	d.Logger.With("staging table", stagingTableName).InfoContext(ctx, "successfully dropped staging table")

	return nil
}

// Run loads a Postgres destination table with the results of a query.
func (d *PostgresSQLLoader) load(ctx context.Context, importId string, query string, mode string, stagingTableName string, destTableName string) error {

	err := d.createStagingTable(ctx, query, stagingTableName)
	if err != nil {
		return fmt.Errorf("failed to create staging table: %w", err)
	}

	err = d.addImportIDColumn(ctx, stagingTableName, importId)
	if err != nil {
		return fmt.Errorf("failed to add import id column: %w", err)
	}

	colNames, err := d.getColumnNamesFromTable(ctx, destTableName)
	if err != nil {
		return fmt.Errorf("failed to get column names from destination table: %w", err)
	}

	if mode == "append" {
		err = d.deleteFromTable(ctx, destTableName, importId)
		if err != nil {
			return fmt.Errorf("failed to delete data in destination table: %w", err)
		}

		err = d.appendStagingData(ctx, stagingTableName, destTableName, colNames)
		if err != nil {
			return fmt.Errorf("failed to append data in destination table: %w", err)
		}

		err = d.dropStagingTable(ctx, stagingTableName)
		if err != nil {
			return fmt.Errorf("failed to clean up staging table: %w", err)
		}
		return nil
	}

	return nil
}

func (d *PostgresSQLLoader) Load(ctx context.Context, params LoadParams) error {
	return d.load(ctx, params.ImportID, params.Query, params.Mode, params.StagingTableName, params.DestTableName)
}

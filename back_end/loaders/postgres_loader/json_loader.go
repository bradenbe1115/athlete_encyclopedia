package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/marcboeker/go-duckdb"
)

var (
	ErrUnexpectedQueryResultsSchema = errors.New("query results schema does not match expected")
)

// DuckDBConnector knows how to connect to a duckdb database.
type DuckDBConnector interface {
	Connect(context.Context) (*sql.DB, error)
}

type DuckDBPostgresConnector struct {
	ConnURI string
}

// Connect creates a duckdb connection with a Postgres database attachment.
func (d *DuckDBPostgresConnector) Connect(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create duckdb connection: %w", err)
	}

	if _, err := db.ExecContext(ctx, `INSTALL postgres; LOAD postgres;`); err != nil {
		return nil, fmt.Errorf("failed to load DuckDB Postgres extension: %v", err)
	}

	query := fmt.Sprintf(`ATTACH 'postgres:%s' AS pg (TYPE POSTGRES)`, d.ConnURI)
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to attach Postgres: %w", err)
	}
	return db, nil
}

type DuckDBJSONPostgresLoader struct {
	DB     *sql.DB
	Logger *slog.Logger
}

// ExecuteQuery executes a query using the duckdb client.
func (d *DuckDBJSONPostgresLoader) ExecuteQuery(ctx context.Context, query string) (sql.Result, error) {
	r, err := d.DB.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}

	return r, nil
}

// Creates a staging table from raw file.
func (d *DuckDBJSONPostgresLoader) createStagingTable(ctx context.Context, table string, rawFilePath string) error {
	fullQuery := fmt.Sprintf(`
	CREATE TEMPORARY TABLE %s AS
	SELECT
		*
	FROM read_json_auto('%s', union_by_name=true)`, table, rawFilePath)

	d.Logger.With("table", table, "rawFilePath", rawFilePath).InfoContext(ctx, "creating temp table")
	_, err := d.ExecuteQuery(ctx, fullQuery)
	if err != nil {
		return fmt.Errorf("failed to create staging tables: %w", err)
	}
	d.Logger.With("table", table, "rawFilePath", rawFilePath).InfoContext(ctx, "created temp table")
	return nil
}

// AddImportIdColumn appends an import id column to the table and populates the value.
func (d *DuckDBJSONPostgresLoader) addImportIDColumn(ctx context.Context, table string, importID string) error {
	query := fmt.Sprintf(`
    ALTER TABLE %s
    ADD COLUMN _import_id VARCHAR DEFAULT '%s'`, table, importID)

	_, err := d.ExecuteQuery(ctx, query)
	return err
}

type DBCol struct {
	Name           string
	DuckType       string
	PostgresType   string
	NormalizedName string
}

// MapDuckDBTypeToPostgres maps duckdb data types to postgres data types and updates DBCol struct.
func mapDuckDBTypeToPostgres(duckType string) string {
	switch strings.ToUpper(duckType) {
	case "VARCHAR", "STRING":
		return "TEXT"
	case "BIGINT", "INTEGER":
		return "BIGINT"
	case "DOUBLE", "FLOAT":
		return "DOUBLE PRECISION"
	case "BOOLEAN":
		return "BOOLEAN"
	case "DATE":
		return "DATE"
	case "TIMESTAMP":
		return "TIMESTAMP"
	default:
		return "TEXT"
	}
}

// Normalize replaces all spaces in a string with an underscore
func normalize(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "_"))
}

// GetDuckDBTableSchema gets schema of duckdb table.
func (d *DuckDBJSONPostgresLoader) getDuckDBTableSchema(ctx context.Context, tableName string) ([]DBCol, error) {
	schemaQuery := fmt.Sprintf(`PRAGMA table_info('%s')`, tableName)
	rows, err := d.DB.QueryContext(ctx, schemaQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get column info: %w", err)
	}
	defer rows.Close()

	var dbCols []DBCol
	for rows.Next() {
		var dbCol DBCol
		var cid int
		var notnull, pk bool
		var dfltValue interface{}
		if err := rows.Scan(&cid, &dbCol.Name, &dbCol.DuckType, &notnull, &dfltValue, &pk); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedQueryResultsSchema, err)
		}
		dbCol.PostgresType = mapDuckDBTypeToPostgres(dbCol.DuckType)
		dbCol.NormalizedName = normalize(dbCol.Name)
		dbCols = append(dbCols, dbCol)
	}

	return dbCols, nil
}

// CreateColDefs creates strings that pair column names with data types.
func createColDefs(dc []DBCol) []string {
	var colDefs []string
	for _, c := range dc {
		colDefs = append(colDefs, fmt.Sprintf("%s %s", c.NormalizedName, c.PostgresType))
	}

	return colDefs
}

// CreateDestTable creates a destination table using provided db columns as a schema definition.
func (d *DuckDBJSONPostgresLoader) createDestTable(ctx context.Context, destTableName string, dc []DBCol) error {
	colDefs := createColDefs(dc)

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS pg.public.%s (%s)`, destTableName, strings.Join(colDefs, ", "))

	d.Logger.With("destTableName", destTableName, "query", query).InfoContext(ctx, "creating destination tables")
	_, err := d.ExecuteQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create destination table: %w", err)
	}
	d.Logger.With("destTableName", destTableName, "query", query).InfoContext(ctx, "created destination tables")

	return nil
}

// LoadFromStagingtoPostgres loads data from staging table into a postgres destination table.
func (d *DuckDBJSONPostgresLoader) loadFromStagingtoPostgres(ctx context.Context, stagingTable string, destTableName string, dc []DBCol) error {
	var colNames []string
	for _, c := range dc {
		colNames = append(colNames, c.NormalizedName)
	}
	query := fmt.Sprintf(`
	INSERT INTO pg.public.%s (%s) SELECT %s FROM %s`,
		destTableName, strings.Join(colNames, ", "), strings.Join(colNames, ", "), stagingTable)

	d.Logger.With("stagingTable", stagingTable, "destTableName", destTableName).InfoContext(ctx, "loading data into postgres")
	_, err := d.ExecuteQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to load data to postgres: %w", err)
	}
	return nil
}

// load loads JSON data into Postgres table.
func (d *DuckDBJSONPostgresLoader) load(ctx context.Context, importId string, rawFilePath string, stagingTableName string, destTableName string) error {

	err := d.createStagingTable(ctx, stagingTableName, rawFilePath)
	if err != nil {
		return fmt.Errorf("failed to create staging table: %w", err)
	}

	err = d.addImportIDColumn(ctx, stagingTableName, importId)
	if err != nil {
		return fmt.Errorf("failed to add import id column to staging table: %w", err)
	}

	dc, err := d.getDuckDBTableSchema(ctx, stagingTableName)
	if err != nil {
		return fmt.Errorf("failed to get db schema: %w", err)
	}

	err = d.createDestTable(ctx, destTableName, dc)
	if err != nil {
		return fmt.Errorf("failed creating destination table: %w", err)
	}

	err = d.loadFromStagingtoPostgres(ctx, stagingTableName, destTableName, dc)
	if err != nil {
		return fmt.Errorf("failed loading data from staging to destination: %w", err)
	}
	return nil
}

func (d *DuckDBJSONPostgresLoader) Load(ctx context.Context, params LoadParams) error {
	return d.load(ctx, params.ImportID, params.RawFilePath, params.StagingTableName, params.DestTableName)
}

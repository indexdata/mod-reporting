package main

import "context"
import "fmt"
import "net/http"
import "encoding/json"
import "github.com/jackc/pgx/v5"
import "github.com/jackc/pgx/v5/pgxpool"


type dbTable struct {
	SchemaName string `db:"schema_name"`
	TableName string `db:"table_name"`
}

type dbColumn struct {
	ColumnName string `db:"column_name"`
	DataType string `db:"data_type"`
	// TableCatalog string `db:"table_catalog"`
	TableSchema string `db:"table_schema"`
	TableName string `db:"table_name"`
	OrdinalPosition string `db:"ordinal_position"`
	// ColumnDefault string `db:"column_default"`
	// IsNullable string `db:"is_nullable"`
	// CharacterMaximumLength string `db:"character_maximum_length"`
	// CharacterOctetLength string `db:"character_octet_length"`
	// NumericPrecision string `db:"numeric_precision"`
	// NumericPrecisionRadix string `db:"numeric_precision_radix"`
	// NumericScale string `db:"numeric_scale"`
	// DatetimePrecision string `db:"datetime_precision"`
	// IntervalType string `db:"interval_type"`
	// IntervalPrecision string `db:"interval_precision"`
	// CharacterSetCatalog string `db:"character_set_catalog"`
	// CharacterSetSchema string `db:"character_set_schema"`
	// CharacterSetName string `db:"character_set_name"`
	// CollationCatalog string `db:"collation_catalog"`
	// CollationSchema string `db:"collation_schema"`
	// CollationName string `db:"collation_name"`
	// DomainCatalog string `db:"domain_catalog"`
	// DomainSchema string `db:"domain_schema"`
	// DomainName string `db:"domain_name"`
	// UdtCatalog string `db:"udt_catalog"`
	// UdtSchema string `db:"udt_schema"`
	// UdtName string `db:"udt_name"`
	// ScopeCatalog string `db:"scope_catalog"`
	// ScopeSchema string `db:"scope_schema"`
	// ScopeName string `db:"scope_name"`
	// MaximumCardinality string `db:"maximum_cardinality"`
	// DtdIdentifier string `db:"dtd_identifier"`
	// IsSelfReferencing string `db:"is_self_referencing"`
	// IsIdentity string `db:"is_identity"`
	// IdentityGeneration string `db:"identity_generation"`
	// IdentityStart string `db:"identity_start"`
	// IdentityIncrement string `db:"identity_increment"`
	// IdentityMaximum string `db:"identity_maximum"`
	// IdentityMinimum string `db:"identity_minimum"`
	// ... and probably more
}


func handleTables(w http.ResponseWriter, req *http.Request, server *ModReportingServer) error {
	tables, err := fetchTables(server.dbConn)
	if err != nil {
		return fmt.Errorf("could not fetch tables from reporting DB: %s", err)
	}

	bytes, err := json.Marshal(tables)
	if err != nil {
		return fmt.Errorf("could not encode JSON for tables: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	return err
}


func fetchTables(dbConn *pgxpool.Pool) ([]dbTable, error) {
	// XXX This is hardwired to MetaDB: we should also support LDP Classic
	query := "SELECT schema_name, table_name FROM metadb.base_table"
	rows, err := dbConn.Query(context.Background(), query)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("could not run query '%s': %s", query, err)
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[dbTable])
}


func handleColumns(w http.ResponseWriter, req *http.Request, server *ModReportingServer) error {
	v := req.URL.Query()
	schema := v.Get("schema")
	table := v.Get("table")
	if schema == "" || table == "" {
		return fmt.Errorf("must specify both schema and table")
	}

	columns, err := fetchColumns(server.dbConn, schema, table)
	if err != nil {
		return fmt.Errorf("could not fetch columns from reporting DB: %s", err)
	}

	bytes, err := json.Marshal(columns)
	if err != nil {
		return fmt.Errorf("could not encode JSON for columns: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	return err
}


func fetchColumns(dbConn *pgxpool.Pool, schema string, table string) ([]dbColumn, error) {
	// This seems to work for both MetaDB and LDP Classic
	cols := "column_name, data_type, ordinal_position, table_schema, table_name"
	query := "SELECT " + cols + " FROM information_schema.columns " +
		"WHERE table_schema = $1 AND table_name = $2 AND column_name != $3";
	rows, err := dbConn.Query(context.Background(), query, schema, table, "data")
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("could not run query '%s': %s", query, err)
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[dbColumn])
}

package main

import "context"
import "io"
import "strings"
import "fmt"
import "net/http"
import "encoding/json"
import "github.com/jackc/pgx/v5"
import "github.com/jackc/pgx/v5/pgxpool"


type dbTable struct {
	SchemaName string `db:"schema_name" json:"schemaName"`
	TableName string `db:"table_name" json:"tableName"`
}

type dbColumn struct {
	ColumnName string `db:"column_name" json:"columnName"`
	DataType string `db:"data_type" json:"data_type"`
	// TableCatalog string `db:"table_catalog"`
	TableSchema string `db:"table_schema" json:"tableSchema"`
	TableName string `db:"table_name" json:"tableName"`
	OrdinalPosition string `db:"ordinal_position" json:"ordinalPosition"`
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
		return fmt.Errorf("could not fetch tables from reporting DB: %w", err)
	}

	bytes, err := json.Marshal(tables)
	if err != nil {
		return fmt.Errorf("could not encode JSON for tables: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	return err
}


func fetchTables(dbConn *pgxpool.Pool) ([]dbTable, error) {
	// XXX This is hardwired to MetaDB: we should also support LDP Classic
	query := "SELECT schema_name, table_name FROM metadb.base_table"
	rows, err := dbConn.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("could not run query '%s': %w", query, err)
	}
	defer rows.Close()

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
		return fmt.Errorf("could not fetch columns from reporting DB: %w", err)
	}

	bytes, err := json.Marshal(columns)
	if err != nil {
		return fmt.Errorf("could not encode JSON for columns: %w", err)
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
	if err != nil {
		return nil, fmt.Errorf("could not run query '%s': %w", query, err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[dbColumn])
}


type queryFilter struct {
	Key string `json:"key"`
	Op string `json:"op"`
	Value string `json:"value"`
}

type queryOrder struct {
	Key string `json:"key"`
	Direction string `json:"direction"`
	Nulls string `json:"nulls"`
}

type queryTable struct {
	Schema string `json:"schema"`
	Table string `json:"tableName"`
	Filters []queryFilter `json:"columnFilters"`
	Columns []string `json:"showColumns"`
	Order []queryOrder `json:"orderBy"`
	Limit int `json:"limit"`
}

type jsonQuery struct {
	Tables []queryTable `json:"tables"`
}


func handleQuery(w http.ResponseWriter, req *http.Request, server *ModReportingServer) error {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("could not read HTTP request body: %w", err)
	}
	var query jsonQuery
	err = json.Unmarshal(bytes, &query)
	if err != nil {
		return fmt.Errorf("could not deserialize JSON from body: %w", err)
	}

	sql, params, err := makeSql(query)
	if err != nil {
		return fmt.Errorf("could not generate SQL from JSON query: %w", err)
	}

	server.Log("sql", sql, fmt.Sprintf("%v", params))
	rows, err := server.dbConn.Query(context.Background(), sql, params...)
	if err != nil {
		return fmt.Errorf("could not execute SQL from JSON query: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		return fmt.Errorf("could not collect query result data: %w", err)
	}

	// Fix up types
	for _, rec := range result {
		for key, val := range rec {
			switch v := val.(type) {
			case [16]uint8:
				// This is how pgx represents fields of type "uuid"
				rec[key] = fmt.Sprintf("%x-%x-%x-%x-%x", v[0:4], v[4:6], v[6:8], v[8:10], v[10:16])
			default:
				// Nothing to do
			}
		}
	}

	// XXX From here on, we should share code with handleTables and handleColumns
	bytes, err = json.Marshal(result)
	if err != nil {
		return fmt.Errorf("could not encode JSON for query result: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	return err
}


func makeSql(query jsonQuery) (string, []any, error) {
	if len(query.Tables) != 1 {
		return "", nil, fmt.Errorf("query must have exactly one table")
	}
	qt := query.Tables[0]

	sql := "SELECT " + makeColumns(qt.Columns) + ` FROM "` + qt.Schema + `"."` + qt.Table + `"`
	if len(qt.Filters) > 0 {
		sql += " WHERE " + makeCond(qt.Filters)
	}
	params := make([]any, len(qt.Filters))
	for i, val := range(qt.Filters) {
		params[i] = val.Value
	}
	if len(qt.Order) > 0 {
		sql += " ORDER BY " + makeOrder(qt.Order)
	}
	if qt.Limit != 0 {
		sql += fmt.Sprintf(" LIMIT %d", qt.Limit)
	}

	return sql, params, nil
}


func makeColumns(cols []string) string {
	if len(cols) == 0 {
		return "*"
	}

	s := ""
	for i, col := range(cols) {
		s += col
		if i < len(cols)-1 {
			s += ", "
		}
	}

	return s
}


func makeCond(filters []queryFilter) string {
	s := ""
	for i, filter := range(filters) {
		s += filter.Key
		if filter.Op == "" {
			s += " = "
		} else {
			s += " " + filter.Op + " "
		}
		s += fmt.Sprintf("$%d", i+1)
		if i < len(filters)-1 {
			s += " AND "
		}
	}

	return s
}


func makeOrder(orders []queryOrder) string {
	s := ""
	for i, order := range(orders) {
		s += order.Key
		s += " " + order.Direction
		// Historically, ui-ldp sends "start" or "end"
		// But we also want to support PostgreSQL's own "FIRST" and "LAST"
		if strings.EqualFold(order.Nulls, "first") ||
			strings.EqualFold(order.Nulls, "start") {
			s += " NULLS FIRST"
		} else {
			s += " NULLS LAST"
		}
		if i < len(orders)-1 {
			s += ", "
		}
	}

	return s
}

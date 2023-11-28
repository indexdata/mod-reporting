package main

import "context"
import "fmt"
import "net/http"
import "encoding/json"
import "github.com/jackc/pgx/v5"


type dbTable struct {
	SchemaName string `db:"schema_name"`
	TableName string `db:"table_name"`
}


func handleTables(w http.ResponseWriter, req *http.Request, server *ModReportingServer) error {
	tables, err := fetchTables(server)
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


func fetchTables(server *ModReportingServer) ([]dbTable, error) {
	// XXX This is hardwired to MetaDB: we should also support LDP Classic
	query := "SELECT schema_name, table_name FROM metadb.base_table"
	rows, err := server.dbConn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[dbTable])
}

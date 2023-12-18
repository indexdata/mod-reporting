package main

import "io"
import "strings"
import "testing"
import "github.com/stretchr/testify/assert"
import "github.com/pashagolub/pgxmock/v3"
import "net/http/httptest"


var reportingTests []testT = []testT{
	{
		name: "bad DB connection for tables",
		useBadSession: true,
		function: handleTables,
		errorstr: "failed to connect",
	},
	{
		name: "bad DB connection for columns",
		path: "/ldp/db/columns?schema=folio_users&table=users",
		useBadSession: true,
		function: handleColumns,
		errorstr: "failed to connect",
	},
	{
		name: "retrieve list of tables",
		path: "/ldp/db/tables",
		establishMock: func(data interface{}) error {
			mock := data.(pgxmock.PgxPoolIface)
			mock.ExpectQuery("SELECT schema_name, table_name FROM metadb.base_table").WillReturnRows(
				pgxmock.NewRows([]string{"schema_name", "table_name"}).
					AddRow("folio_inventory", "records_instances").
					AddRow("folio_inventory", "holdings_record"))
			return nil
		},
		function: handleTables,
		expected: `[{"schemaName":"folio_inventory","tableName":"records_instances"},{"schemaName":"folio_inventory","tableName":"holdings_record"}]`,
	},
	{
		name: "list of columns without table",
		path: "/ldp/db/columns?schema=folio_users",
		function: handleColumns,
		errorstr: "must specify both schema and table",
	},
	{
		name: "list of columns without schema",
		path: "/ldp/db/columns?table=users",
		function: handleColumns,
		errorstr: "must specify both schema and table",
	},
	{
		name: "retrieve list of columns",
		path: "/ldp/db/columns?schema=folio_users&table=users",
		establishMock: func(data interface{}) error {
			mock := data.(pgxmock.PgxPoolIface)
			mock.ExpectQuery(`SELECT`).
				WithArgs("folio_users", "users", "data").
				WillReturnRows(pgxmock.NewRows([]string{"column_name", "data_type", "ordinal_position", "table_schema", "table_name"}).
					AddRow("id", "uuid", "6", "folio_users", "users").
					AddRow("creation_date", "timestamp without time zone", "8", "folio_users", "users"))
			return nil
		},
		function: handleColumns,
		expected: `{"columnName":"id","data_type":"uuid","tableSchema":"folio_users","tableName":"users","ordinalPosition":"6"},{"columnName":"creation_date","data_type":"timestamp without time zone","tableSchema":"folio_users","tableName":"users","ordinalPosition":"8"}]`,
	},
}


func Test_handleTables(t *testing.T) {
	ts := MakeDummyModSettingsServer()
	defer ts.Close()
	baseUrl := ts.URL

	mrs, err := MakeConfiguredServer("../etc/silent.json", ".")
	assert.Nil(t, err)
	session, err := NewModReportingSession(mrs, baseUrl, "dummyTenant")
	assert.Nil(t, err)

	mock, err := pgxmock.NewPool()
	assert.Nil(t, err)
	defer mock.Close()

	for _, test := range reportingTests {
		t.Run(test.name, func(t *testing.T) {
			method := "GET"
			var reader io.Reader
			if test.sendData != "" {
				method = "POST"
				reader = strings.NewReader(test.sendData)
			}
			req := httptest.NewRequest(method, baseUrl + test.path, reader)

			if test.establishMock != nil {
				err = test.establishMock(mock)
				assert.Nil(t, err)
			}

			if test.useBadSession {
				session.dbConn = nil
			} else {
				session.dbConn = mock
			}

			w := httptest.NewRecorder()
			err = test.function(w, req, session)
			resp := w.Result()

			if test.errorstr == "" {
				assert.Nil(t, err)
				assert.Equal(t, resp.StatusCode, 200)
				body, _ := io.ReadAll(resp.Body)
				assert.Regexp(t, test.expected, string(body))
			} else {
				assert.ErrorContains(t, err, test.errorstr)
			}

			assert.Nil(t, mock.ExpectationsWereMet(), "unfulfilled expections")
		})
	}
}

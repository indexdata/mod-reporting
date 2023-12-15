package main

import "io"
import "testing"
import "github.com/stretchr/testify/assert"
import "github.com/pashagolub/pgxmock/v3"
import "net/http/httptest"


func Test_handleTables(t *testing.T) {
	ts := MakeDummyModSettingsServer()
	defer ts.Close()
	baseUrl := ts.URL

	mrs, err := MakeConfiguredServer("../etc/silent.json", ".")
	assert.Nil(t, err)
	session, err := NewModReportingSession(mrs, baseUrl, "dummyTenant")
	assert.Nil(t, err)

	t.Run("unable to obtain DB connection", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/dummy", nil)
		w := httptest.NewRecorder()
		err = handleTables(w, req, session)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to connect")
	})

	mock, err := pgxmock.NewPool()
	assert.Nil(t, err)
	defer mock.Close()
	session.dbConn = mock

	t.Run("retrieve list of tables", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"schema_name", "table_name"}).
			AddRow("folio_inventory", "records_instances").
			AddRow("folio_inventory", "holdings_record")
		mock.ExpectQuery("SELECT schema_name, table_name FROM metadb.base_table").WillReturnRows(rows)

		req := httptest.NewRequest("GET", "http://example.com/dummy", nil)
		w := httptest.NewRecorder()
		err = handleTables(w, req, session)
		resp := w.Result()

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, string(body), `[{"schemaName":"folio_inventory","tableName":"records_instances"},{"schemaName":"folio_inventory","tableName":"holdings_record"}]`)
		assert.Nil(t, mock.ExpectationsWereMet(), "unfulfilled expections")
	})

	t.Run("list of columns without table", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/dummy?schema=folio_users", nil)
		w := httptest.NewRecorder()
		err = handleColumns(w, req, session)

		assert.ErrorContains(t, err, "must specify both schema and table")
	})

	t.Run("list of columns without schema", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/dummy?table=users", nil)
		w := httptest.NewRecorder()
		err = handleColumns(w, req, session)

		assert.ErrorContains(t, err, "must specify both schema and table")
	})

	t.Run("retrieve list of columns", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"column_name", "data_type", "ordinal_position", "table_schema", "table_name"}).
			AddRow("id", "uuid", "6", "folio_users", "users").
			AddRow("creation_date", "timestamp without time zone", "8", "folio_users", "users")
		mock.ExpectQuery(`SELECT`).
			WithArgs("folio_users", "users", "data").
			WillReturnRows(rows)

		req := httptest.NewRequest("GET", "http://example.com/dummy?schema=folio_users&table=users", nil)
		w := httptest.NewRecorder()
		err = handleColumns(w, req, session)
		resp := w.Result()

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, `[{"columnName":"id","data_type":"uuid","tableSchema":"folio_users","tableName":"users","ordinalPosition":"6"},{"columnName":"creation_date","data_type":"timestamp without time zone","tableSchema":"folio_users","tableName":"users","ordinalPosition":"8"}]`, string(body))

		assert.Nil(t, mock.ExpectationsWereMet(), "unfulfilled expections")
	})
}

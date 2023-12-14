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

	mock, err := pgxmock.NewPool()
	assert.Nil(t, err)
	defer mock.Close()
	session.dbConn = mock

	rows := pgxmock.NewRows([]string{"schema_name", "table_name"}).
		AddRow("folio_inventory", "records_instances").
		AddRow("folio_inventory", "holdings_record")
	mock.ExpectQuery("SELECT schema_name, table_name FROM metadb.base_table").WillReturnRows(rows)

	t.Run("retrieve list of tables", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/dummy", nil)
		w := httptest.NewRecorder()
		err = handleTables(w, req, session)
		resp := w.Result()

		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, 200)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, `[{"schemaName":"folio_inventory","tableName":"records_instances"},{"schemaName":"folio_inventory","tableName":"holdings_record"}]`, string(body))
	})
}
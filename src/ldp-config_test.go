package main

import "io"
import "fmt"
import "testing"
import "gotest.tools/assert"
import "net/http"
import "net/http/httptest"


func Test_handleConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
{
  "items": [
    {
      "id": "75c12fcb-ba6c-463f-a5fc-cb0587b7d43b",
      "scope": "ui-ldp.admin",
      "key": "config",
      "value": "{\"defaultShow\":100,\"maxShow\":1000,\"maxExport\":\"100000\",\"disabledTables\":[],\"tqTabs\":[]}"
    }
  ],
  "resultInfo": {
    "totalRecords": 1,
    "diagnostics": []
  }
}
`))
	}))
	defer ts.Close()
	baseUrl := ts.URL

	t.Run("fetch config", func(t *testing.T) {
		session, err := NewModReportingSession(nil, baseUrl, "dummyTenant")
		assert.NilError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", baseUrl + "/ldp/config", nil)
		err = handleConfig(w, req, session)
		assert.NilError(t, err)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, 200)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, string(body), `[{"key":"config","tenant":"dummyTenant","value":"{\"defaultShow\":100,\"maxShow\":1000,\"maxExport\":\"100000\",\"disabledTables\":[],\"tqTabs\":[]}"}]`)
	})

}

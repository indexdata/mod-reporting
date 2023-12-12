package main

import "io"
import "fmt"
import "testing"
import "gotest.tools/assert"
import "net/http"
import "net/http/httptest"


func MakeDummyFolioServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/settings/entries" &&
			req.URL.RawQuery == `query=scope=="ui-ldp.admin"+and+key=="dbinfo"` {
			_, _ = w.Write([]byte(`
{
  "items": [
    {
      "id": "75c12fcb-ba6c-463f-a5fc-cb0587b7d43c",
      "scope": "ui-ldp.admin",
      "key": "dbinfo",
      "value": "{\"defaultShow\":100,\"maxShow\":1000,\"maxExport\":\"100000\",\"disabledTables\":[],\"tqTabs\":[]}"
    }
  ],
  "resultInfo": {
    "totalRecords": 1,
    "diagnostics": []
  }
}
`))
		} else if req.URL.Path == "/settings/entries" {
			_, _ = w.Write([]byte(`
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
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Not found")
		}
	}))
}


func Test_handleConfig(t *testing.T) {
	tests := []struct {
		name string
		path string
		function func(w http.ResponseWriter, req *http.Request, session *ModReportingSession) error
		expected string
	}{
		{"fetch all configs from table", "/ldp/config", handleConfig,`[{"key":"config","tenant":"dummyTenant","value":"{\"defaultShow\":100,\"maxShow\":1000,\"maxExport\":\"100000\",\"disabledTables\":[],\"tqTabs\":[]}"}]`},
		{"fetch single config", "/ldp/config/dbinfo", handleConfigKey, `{"key":"dbinfo","tenant":"dummyTenant","value":"{\"defaultShow\":100,\"maxShow\":1000,\"maxExport\":\"100000\",\"disabledTables\":[],\"tqTabs\":[]}"}`},
	}

	ts := MakeDummyFolioServer()
	defer ts.Close()
	baseUrl := ts.URL

	session, err := NewModReportingSession(nil, baseUrl, "dummyTenant")
	assert.NilError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", baseUrl + test.path, nil)
			err = test.function(w, req, session)
			assert.NilError(t, err)

			resp := w.Result()
			assert.Equal(t, resp.StatusCode, 200)
			body, _ := io.ReadAll(resp.Body)
			assert.Equal(t, string(body), test.expected)
		})
	}
}

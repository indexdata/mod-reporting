package main

import "io"
import "strings"
import "fmt"
import "testing"
import "github.com/stretchr/testify/assert"
import "net/http"
import "net/http/httptest"


type testT struct {
	name string
	path string
	putData string
	function func(w http.ResponseWriter, req *http.Request, session *ModReportingSession) error
	expected string
	errorstr string
	useBadServer bool
}

var tests []testT = []testT{
	{
		"fetch all configs from table",
		"/ldp/config",
		"",
		handleConfig,
		`[{"key":"config","tenant":"dummyTenant","value":"v1"}]`,
		"",
		false,
	},
	{
		"fetch single config",
		"/ldp/config/dbinfo",
		"",
		handleConfigKey,
		`{"key":"dbinfo","tenant":"dummyTenant","value":"v2"}`,
		"",
		false,
	},
	{
		"non-existent config",
		"/ldp/config/not-there",
		"",
		handleConfigKey,
		"",
		"no config item with key",
		false,
	},
	{
		"fetch malformed config",
		"/ldp/config/bad",
		"",
		handleConfigKey,
		"",
	        "could not deserialize",
		false,
	},
	{
		"translate non-string value",
		"/ldp/config/non-string",
		"",
		handleConfigKey,
		`{"key":"non-string","tenant":"dummyTenant","value":"{\\|"v3\\\":42}"}`,
	        "",
		false,
	},
	{
		"failure to reach mod-settings",
		"/ldp/config/non-string",
		"",
		handleConfig,
		"",
	        "could not fetch from mod-settings",
		true,
	},
	{
		"write a config value",
		"/ldp/config/foo",
		`{"key":"foo","tenant":"xxx","value":"{\"user\":\"abc123\"}"}`,
		handleConfigKey,
		"abc123",
		"",
		false,
	},
	// At this point it seems silly to laboriously chase each individual error case
}


func MakeDummyFolioServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/settings/entries" &&
			req.URL.RawQuery == `query=scope=="ui-ldp.admin"` {
			_, _ = w.Write([]byte(`
{
  "items": [
    {
      "id": "75c12fcb-ba6c-463f-a5fc-cb0587b7d43b",
      "scope": "ui-ldp.admin",
      "key": "config",
      "value": "v1"
    }
  ],
  "resultInfo": {
    "totalRecords": 1,
    "diagnostics": []
  }
}
`))
		} else if req.URL.Path == "/settings/entries" &&
			req.URL.RawQuery == `query=scope=="ui-ldp.admin"+and+key=="dbinfo"` {
			_, _ = w.Write([]byte(`
{
  "items": [
    {
      "id": "75c12fcb-ba6c-463f-a5fc-cb0587b7d43c",
      "scope": "ui-ldp.admin",
      "key": "dbinfo",
      "value": "v2"
    }
  ],
  "resultInfo": {
    "totalRecords": 1,
    "diagnostics": []
  }
}
`))
		} else if req.URL.Path == "/settings/entries" &&
			req.URL.RawQuery == `query=scope=="ui-ldp.admin"+and+key=="non-string"` {
			_, _ = w.Write([]byte(`
{
  "items": [
    {
      "id": "75c12fcb-ba6c-463f-a5fc-cb0587b7d43c",
      "scope": "ui-ldp.admin",
      "key": "non-string",
      "value": { "v3": 42 }
    }
  ],
  "resultInfo": {
    "totalRecords": 1,
    "diagnostics": []
  }
}
`))
		} else if req.URL.Path == "/settings/entries" &&
			req.URL.RawQuery == `query=scope=="ui-ldp.admin"+and+key=="bad"` {
			_, _ = w.Write([]byte("some bit of text"))
		} else if req.URL.Path == "/settings/entries" {
			// Searching for some other setting, e.g. "score" before trying to write to it
			_, _ = w.Write([]byte(`
{
  "items": [],
  "resultInfo": {
    "totalRecords": 0,
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
	ts := MakeDummyFolioServer()
	defer ts.Close()
	baseUrl := ts.URL

	session, err := NewModReportingSession(nil, baseUrl, "dummyTenant")
	assert.Nil(t, err)
	badSession, err := NewModReportingSession(nil, "x" + baseUrl, "dummyTenant")
	assert.Nil(t, err)

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			method := "GET"
			var reader io.Reader
			if test.putData != "" {
				method = "PUT"
				reader = strings.NewReader(test.putData)
			}
			req := httptest.NewRequest(method, baseUrl + test.path, reader)
			if i == 0 {
				// Just to exercise a code-path, and get slightly more coverage *sigh*
				req.Header.Add("X-Okapi-Token", "dummy")
			}

			var currentSession = session
			if test.useBadServer {
				currentSession = badSession
			}

			w := httptest.NewRecorder()
			err = test.function(w, req, currentSession)
			resp := w.Result()

			if test.errorstr == "" {
				assert.Nil(t, err)
				assert.Equal(t, resp.StatusCode, 200)
				body, _ := io.ReadAll(resp.Body)
				assert.Regexp(t, test.expected, string(body))
			} else {
				assert.ErrorContains(t, err, test.errorstr)
			}
		})
	}
}

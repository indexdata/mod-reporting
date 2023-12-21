package main

import "io"
import "strings"
import "fmt"
import "testing"
import "github.com/stretchr/testify/assert"
import "net/http"
import "net/http/httptest"


func MakeDummyModSettingsServer() *httptest.Server {
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
			// XXX note that this specific value is also required by the getDbInfo test
			_, _ = w.Write([]byte(`
			  {
			    "items": [
			      {
				"id": "75c12fcb-ba6c-463f-a5fc-cb0587b7d43c",
				"scope": "ui-ldp.admin",
				"key": "dbinfo",
				"value": {
				  "url": "dummyUrl",
				  "user": "fiona",
				  "pass": "pw"
				}
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
		} else if req.URL.Path == "/settings/entries/75c12fcb-ba6c-463f-a5fc-cb0587b7d43c" {
			// Nothing to do
		} else if req.URL.Path == "/reports/noheader.sql" {
			_, _ = w.Write([]byte(`this is a bad report`))
		} else if req.URL.Path == "/reports/bad.sql" {
			_, _ = w.Write([]byte(`--metadb:function users\nthis is bad SQL`))
		} else if req.URL.Path == "/reports/loans.sql" {
			_, _ = w.Write([]byte(`--metadb:function count_loans

DROP FUNCTION IF EXISTS count_loans;

CREATE FUNCTION count_loans(
    start_date date DEFAULT '1000-01-01',
    end_date date DEFAULT '3000-01-01')
RETURNS TABLE(
    item_id uuid,
    loan_count bigint)
AS $$
SELECT item_id,
       count(*) AS loan_count
    FROM folio_circulation.loan__t
    WHERE start_date <= loan_date AND loan_date < end_date
    GROUP BY item_id
$$
LANGUAGE SQL
STABLE
PARALLEL SAFE;
`))
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Not found")
		}
	}))
}


type testT struct {
	name string
	path string
	sendData string
	establishMock func(data interface{}) error
	function func(w http.ResponseWriter, req *http.Request, session *ModReportingSession) error
	expected string
	expectedArgs []string // Used only in reporting_test.go/Test_makeSql
	errorstr string
	useBadSession bool
}


func Test_handleConfig(t *testing.T) {
	tests := []testT{
		{
			name: "fetch all configs from table",
			path: "/ldp/config",
			function: handleConfig,
			expected: `[{"key":"config","tenant":"dummyTenant","value":"v1"}]`,
		},
		{
			name: "fetch single config",
			path: "/ldp/config/dbinfo",
			function: handleConfigKey,
			expected: `{"key":"dbinfo","tenant":"dummyTenant","value":`,
		},
		{
			name: "non-existent config",
			path: "/ldp/config/not-there",
			function: handleConfigKey,
			errorstr: "no config item with key",
		},
		{
			name: "fetch malformed config",
			path: "/ldp/config/bad",
			function: handleConfigKey,
			errorstr: "could not deserialize",
		},
		{
			name: "translate non-string value",
			path: "/ldp/config/non-string",
			function: handleConfigKey,
			expected: `{"key":"non-string","tenant":"dummyTenant","value":"{\\|"v3\\\":42}"}`,
		},
		{
			name: "failure to reach mod-settings",
			path: "/ldp/config/non-string",
			function: handleConfig,
			errorstr: "could not fetch from mod-settings",
			useBadSession: true,
		},
		{
			name: "write a new config value",
			path: "/ldp/config/foo",
			sendData: `{"key":"foo","tenant":"xxx","value":"{\"user\":\"abc123\"}"}`,
			function: handleConfigKey,
			expected: "abc123",
		},
		{
			name: "rewrite an existing config value",
			path: "/ldp/config/dbinfo",
			sendData: `{"key":"dbinfo","tenant":"xxx","value":"{\"user\":\"abc456\"}"}`,
			function: handleConfigKey,
			expected: "abc456",
		},
		// At this point it seems silly to laboriously chase each individual error case
	}

	ts := MakeDummyModSettingsServer()
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
			if test.sendData != "" {
				method = "PUT"
				reader = strings.NewReader(test.sendData)
			}
			req := httptest.NewRequest(method, baseUrl + test.path, reader)
			if i == 0 {
				// Just to exercise a code-path, and get slightly more coverage *sigh*
				req.Header.Add("X-Okapi-Token", "dummy")
			}

			var currentSession = session
			if test.useBadSession {
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

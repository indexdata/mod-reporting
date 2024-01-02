package main

import "fmt"
import "net/http"
import "net/http/httptest"


// Dummy HTTP server used by multiple tests
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
		} else if req.URL.Path == "/authn/login-with-expiry" {
			// Attempted login to create new FOLIO session
			fmt.Fprintln(w, `{"accessTokenExpiration":"2023-12-22T12:35:47Z"}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Not found")
		}
	}))
}


// Various parts of this structure are used by different files' tests
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

package main

import "testing"
import "os"
import "strings"
import "io"
import "fmt"
import "time"
import "net/http"
import "regexp"
import "github.com/pashagolub/pgxmock/v3"
import "github.com/stretchr/testify/assert"


func Test_server(t *testing.T) {
	ts := MakeDummyModSettingsServer()
	defer ts.Close()
	server, err := MakeConfiguredServer("../etc/silent.json", "..")
	assert.Nil(t, err)
	session, err := NewModReportingSession(server, ts.URL, "t1")
	assert.Nil(t, err)
	server.sessions[":" + ts.URL] = session

	go func() {
		err = server.launch()
	}()

	// Allow half a second for the server to start. This is ugly
	time.Sleep(time.Second / 2)
	runTests(t, ts.URL, session)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot launch server: %s\n", err)
		os.Exit(3)
	}
}


func runTests(t *testing.T, baseUrl string, session *ModReportingSession) {
	data := []testT{
		{
			name: "home",
			status: 200,
			expected: "This is .*mod-reporting",
		},
		{
			name: "health check",
			path: "admin/health",
			status: 200,
			expected: "Behold!",
		},
		{
			name: "short bad path",
			path: "foo",
			status: 404,
		},
		{
			name: "long bad path",
			path: "foo/bar/baz",
			status: 404,
		},
		{
			name: "get all config",
			path: "ldp/config",
			status: 200,
			expected: `\[{"key":"config","tenant":"t1","value":"v1"}\]`,
		},
		{
			name: "get single config",
			path: "ldp/config/dbinfo",
			status: 200,
			expected: `{"key":"dbinfo","tenant":"t1","value":"{\\"pass\\":\\"pw\\",\\"url\\":\\"dummyUrl\\",\\"user\\":\\"fiona\\"}"}`,
		},
		{
			name: "create new config",
			sendData: `{"key":"foo","tenant":"xxx","value":"{\"user\":\"abc123\"}"}`,
			path: "ldp/config/foo",
			status: 200,
			expected: "abc123",
		},
		{
			name: "rewrite existing config",
			sendData: `{"key":"dbinfo","tenant":"xxx","value":"{\"user\":\"abc456\"}"}`,
			path: "ldp/config/dbinfo",
			status: 200,
			expected: "abc456" ,
		},
		{
			name: "fetch tables",
			path: "/ldp/db/tables",
			establishMock: func(data interface{}) error {
				return establishMockForTables(data.(pgxmock.PgxPoolIface))
			},
			status: 200,
			expected: `\[{"schemaName":"folio_inventory","tableName":"records_instances"},{"schemaName":"folio_inventory","tableName":"holdings_record"}\]`,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			method := "GET"
			var bodyReader io.Reader
			if d.sendData != "" {
				method = "PUT"
				bodyReader = strings.NewReader(d.sendData)
			}

			url := "http://localhost:12369/" + d.path
			req, err := http.NewRequest(method, url, bodyReader)
			assert.Nil(t, err)
			req.Header.Add("X-Okapi-URL", baseUrl)

			mock, _ := pgxmock.NewPool()
			if d.establishMock != nil {
				err = d.establishMock(mock)
				assert.Nil(t, err)
			}
			session.dbConn = mock

			client := http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Errorf("cannot fetch %s: %v", url, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != d.status {
				t.Errorf("fetch %s had status %s (expected %d)", url, resp.Status, d.status)
				// Do not return; attempt the remaining checks
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("cannot read body %s: %v", url, err)
				return
			}
			matched, err := regexp.Match(d.expected, body)
			if err != nil {
				t.Errorf("cannot match body of %s against regexp /%s/: %v", url, d.expected, err)
				return
			}
			if !matched {
				t.Errorf("body of %s does not match regexp /%s/: body = %s", url, d.expected, body)
			}
			assert.Nil(t, mock.ExpectationsWereMet(), "unfulfilled expections")
		})
	}
}

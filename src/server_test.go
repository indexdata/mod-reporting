package main

import "testing"
import "os"
import "strings"
import "io"
import "fmt"
import "time"
import "net/http"
import "regexp"
import "github.com/stretchr/testify/assert"


func Test_server(t *testing.T) {
	server, err := MakeConfiguredServer("../etc/silent.json", "..")
	assert.Nil(t, err)
	go func() {
		err = server.launch()
	}()

	// Allow half a second for the server to start. This is ugly
	time.Sleep(time.Second / 2)
	runTests(t, http.Client{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot launch server: %s\n", err)
		os.Exit(3)
	}
}


func runTests(t *testing.T, client http.Client) {
	data := []struct {
		name   string
		path   string
		status int
		re     string
	}{
		{"home", "", 200, "This is .*mod-reporting"},
		{"health check", "admin/health", 200, "Behold!"},
		{"short bad path", "foo", 404, ""},
		{"long bad path", "foo/bar/baz", 404, ""},
		{"get all config", "ldp/config", 200, `\[{"key":"config","tenant":"","value":"v1"}\]`},
		{"get single config", "ldp/config/dbinfo", 200, `{"key":"dbinfo","tenant":"","value":"{\\"pass\\":\\"pw\\",\\"url\\":\\"dummyUrl\\",\\"user\\":\\"fiona\\"}"}`},
		// XXX more cases to come here
	}

	ts := MakeDummyModSettingsServer()
	defer ts.Close()
	baseUrl := ts.URL

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var bodyReader io.Reader
			sendBody := "" // for now
			if sendBody != "" {
				bodyReader = strings.NewReader(sendBody)
			}

			url := "http://localhost:12369/" + d.path
			req, err := http.NewRequest("GET", url, bodyReader)
			assert.Nil(t, err)
			req.Header.Add("X-Okapi-URL", baseUrl)

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
			matched, err := regexp.Match(d.re, body)
			if err != nil {
				t.Errorf("cannot match body of %s against regexp /%s/: %v", url, d.re, err)
				return
			}
			if !matched {
				t.Errorf("body of %s does not match regexp /%s/: body = %s", url, d.re, body)
			}
		})
	}
}

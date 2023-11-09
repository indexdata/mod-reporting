package main

import "testing"
import "fmt"
import "os"
import "time"
import "net/http"
import "io"
import "regexp"
import "github.com/MikeTaylor/catlogger"

func TestModReporting(t *testing.T) {
	file := "../etc/config.json"

	// For now we duplicate code from main.go
	var cfg *config
	cfg, err := readConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read config file '%s': %v\n", file, err)
		os.Exit(2)
	}

	cl := cfg.Logging
	logger := catlogger.MakeLogger(cl.Categories, cl.Prefix, cl.Timestamp)
	logger.Log("config", fmt.Sprintf("%+v", cfg))

	server := MakeModReportingServer(cfg, logger, "..")
	go func() {
		err = server.launch(cfg.Listen.Host + ":12369")
	}()

	// Allow half a second for the server to start. This is ugly
	time.Sleep(time.Second / 2)
	client := http.Client{}
	runTests(t, client)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create HTTP server:", err)
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
		{"home", "", 200, "This is"},
		{"health check", "admin/health", 200, "Behold."},
		{"short bad path", "foo", 404, ""},
		{"long bad path", "foo/bar/baz", 404, ""},
		// XXX more cases to come here
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			url := "http://localhost:12369/" + d.path
			resp, err := client.Get(url)
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

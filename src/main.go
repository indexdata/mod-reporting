package main

import "os"
import "fmt"
import "github.com/MikeTaylor/catlogger"


func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<configFile.json>")
		os.Exit(1)
	}

	var file = os.Args[1]
	var cfg *config
	cfg, err := readConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read config file '%s': %v\n", file, err)
		os.Exit(2)
	}

	cl := cfg.Logging
	logger := catlogger.MakeLogger(cl.Categories, cl.Prefix, cl.Timestamp)
	logger.Log("config", fmt.Sprintf("%+v", cfg))

	server := MakeModReportingServer(cfg, logger, ".")
	err = server.launch(cfg.Listen.Host + ":" + fmt.Sprint(cfg.Listen.Port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create HTTP server:", err)
		os.Exit(3)
	}
}

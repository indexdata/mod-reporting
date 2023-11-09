package main

import "os"
import "fmt"
import "github.com/MikeTaylor/catlogger"

func MakeConfiguredServer(configFile string, httpRoot string) (*config, *ModReportingServer) {
	var cfg *config
	cfg, err := readConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read config file '%s': %v\n", configFile, err)
		os.Exit(2)
	}

	cl := cfg.Logging
	logger := catlogger.MakeLogger(cl.Categories, cl.Prefix, cl.Timestamp)
	logger.Log("config", fmt.Sprintf("%+v", cfg))

	server := MakeModReportingServer(cfg, logger, httpRoot)
	return cfg, server
}

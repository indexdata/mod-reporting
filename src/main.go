package main

import "os"
import "fmt"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<configFile.json>")
		os.Exit(1)
	}

	cfg, server := MakeConfiguredServer(os.Args[1], ".")
	err := server.launch(cfg.Listen.Host + ":" + fmt.Sprint(cfg.Listen.Port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create HTTP server:", err)
		os.Exit(3)
	}
}

package main

import "os"
import "fmt"
import "github.com/indexdata/foliogo"


func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<configFile.json>")
		os.Exit(1)
	}

	session, err := foliogo.NewDefaultSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: FOLIO session creation failed: %s\n", os.Args[0], err)
		os.Exit(2)
	}

	cfg, server := MakeConfiguredServer(os.Args[1], ".", session)

	dbUrl, dbUser, _ /*dbPass*/, err := getDbInfo(session)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot extract data from 'dbinfo': %s\n", os.Args[0], err)
		os.Exit(2)
	}
	server.Log("db", "url=" + dbUrl + ", user=" + dbUser)

	err = server.launch(cfg.Listen.Host + ":" + fmt.Sprint(cfg.Listen.Port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot create HTTP server: %s\n", os.Args[0], err)
		os.Exit(3)
	}
}

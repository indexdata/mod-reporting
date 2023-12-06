package main

import "os"
import "fmt"
import "github.com/indexdata/foliogo"


func exitIfError(err error, exitStatus int, message string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s: %s\n", os.Args[0], message, err)
		os.Exit(exitStatus)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<configFile.json>")
		os.Exit(1)
	}

	cfg, server := MakeConfiguredServer(os.Args[1], ".", foliogo.Session{})
	// dbUrl, dbUser, dbPass, err := getDbInfo(session)
	// exitIfError(err, 3, "cannot extract data from 'dbinfo'")
	// server.Log("db", "url=" + dbUrl + ", user=" + dbUser)
	//
	// err = server.connectDb(dbUrl, dbUser, dbPass)
	// exitIfError(err, 4, "cannot connect to DB")
	// server.Log("db", "connected to DB", dbUrl)

	err := server.launch(cfg.Listen.Host + ":" + fmt.Sprint(cfg.Listen.Port))
	exitIfError(err, 5, "cannot launch HTTP server")
}

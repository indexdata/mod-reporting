package main

import "os"
import "fmt"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<configFile.json>")
		os.Exit(1)
	}

	server, err := MakeConfiguredServer(os.Args[1], ".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot create server: %s\n", os.Args[0], err)
		os.Exit(2)
	}

	// dbUrl, dbUser, dbPass, err := getDbInfo(session)
	// exitIfError(err, 3, "cannot extract data from 'dbinfo'")
	// server.Log("db", "url=" + dbUrl + ", user=" + dbUser)
	//
	// err = server.connectDb(dbUrl, dbUser, dbPass)
	// exitIfError(err, 4, "cannot connect to DB")
	// server.Log("db", "connected to DB", dbUrl)

	err = server.launch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot launch server: %s\n", os.Args[0], err)
		os.Exit(3)
	}
}

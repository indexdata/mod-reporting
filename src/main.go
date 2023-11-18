package main

import "os"
import "fmt"
import "errors"
import "github.com/indexdata/foliogo"


// This is sensationally clumsy compared with the corresponding JavaScript
func getDbInfo(json foliogo.Hash) (string, string, string, error) {
	resultInfo := json["resultInfo"]
	totalRecords := resultInfo.(map[string]interface{})["totalRecords"]
	count := int(totalRecords.(float64))
	if count < 1 {
		return "", "", "", errors.New("no 'dbinfo' setting in FOLIO database")
	}
	items := json["items"]
	itemArray := items.([]interface{})
	item0 := itemArray[0]
	record := item0.(map[string]interface{})
	value := record["value"]
	hash := value.(map[string]interface{})
	return hash["url"].(string), hash["user"].(string), hash["pass"].(string), nil
}


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

	json, err := session.Fetch(`settings/entries?query=scope=="ui-ldp.admin"+and+key=="dbinfo"`, foliogo.RequestParams{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot fetch 'dbinfo' from config: %s\n", os.Args[0], err)
		os.Exit(2)
	}

	dbUrl, dbUser, _ /*dbPass*/, err := getDbInfo(json)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: extract data from 'dbinfo': %s\n", os.Args[0], err)
		os.Exit(2)
	}
	cfg, server := MakeConfiguredServer(os.Args[1], ".", session)
	server.Log("db", "url=" + dbUrl + ", user=" + dbUser)

	err = server.launch(cfg.Listen.Host + ":" + fmt.Sprint(cfg.Listen.Port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot create HTTP server: %s\n", os.Args[0], err)
		os.Exit(3)
	}
}

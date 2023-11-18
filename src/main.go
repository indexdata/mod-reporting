package main

import "os"
import "fmt"
import "errors"
import "github.com/indexdata/foliogo"


func getDbInfo(json foliogo.Hash) (string, string, string, error) {
	resultInfo := json["resultInfo"]
	totalRecords := resultInfo.(map[string]interface{})["totalRecords"]
	count := int(totalRecords.(float64))
	fmt.Printf("found %d 'dbinfo' settings\n", count);
	if count < 1 {
		return "", "", "", errors.New("no 'dbinfo' setting in FOLIO database")
	}
	items := json["items"]
	fmt.Printf("got items %v\n", items);
	itemArray := items.([]interface{})
	fmt.Printf("got itemArray %v\n", itemArray);
	item0 := itemArray[0]
	fmt.Printf("got item0 %v\n", item0);
	record := item0.(map[string]interface{})
	fmt.Printf("got record %v\n", record);
	value := record["value"]
	fmt.Printf("got value %v\n", value);
	hash := value.(map[string]interface{})
	fmt.Printf("got hash %v\n", hash);
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

	dbUrl, dbUser, dbPass, err := getDbInfo(json)
	fmt.Printf("DB url=%s, user=%s, pass=%s", dbUrl, dbUser, dbPass)

	cfg, server := MakeConfiguredServer(os.Args[1], ".", session)
	err = server.launch(cfg.Listen.Host + ":" + fmt.Sprint(cfg.Listen.Port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot create HTTP server: %s\n", os.Args[0], err)
		os.Exit(3)
	}
}

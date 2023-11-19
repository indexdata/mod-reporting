package main

import "os"
import "fmt"
import "errors"
import "encoding/json"
import "github.com/indexdata/foliogo"


type value struct {
	Url string `json:"url"`
	Pass string `json:"pass"`
	User string `json:"user"`
}

type item struct {
	Value value `json:"value"`
}

type resultInfo struct {
	TotalRecords int `json:"totalRecords"`
}

type response struct {
	Items []item `json:"items"`
	ResultInfo resultInfo `json:"resultInfo"`
}

func getDbInfo(bytes []byte) (string, string, string, error) {
	var r response
	err := json.Unmarshal(bytes, &r)	
	if err != nil {
		return "", "", "", errors.New("decode 'dbinfo' JSON failed: " + err.Error())
	}
	if r.ResultInfo.TotalRecords < 1 {
		return "", "", "", errors.New("no 'dbinfo' setting in FOLIO database")
	}
	value := r.Items[0].Value
	return value.Url, value.User, value.Pass, nil
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

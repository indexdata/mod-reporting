package main

import "os"
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

func getDbInfo(session foliogo.Session) (string, string, string, error) {
	// If defined, environment variables override the setting from the database
	dburl := os.Getenv("REPORTING_DB_URL")
	dbuser := os.Getenv("REPORTING_DB_USER")
	dbpass := os.Getenv("REPORTING_DB_PASS")
	if dburl != "" && dbuser != "" && dbpass != "" {
		return dburl, dbuser, dbpass, nil
	}

	bytes, err := session.Fetch0(`settings/entries?query=scope=="ui-ldp.admin"+and+key=="dbinfo"`)
	if err != nil {
		return "", "", "", errors.New("cannot fetch 'dbinfo' from config: " + err.Error())
	}

	var r response
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return "", "", "", errors.New("decode 'dbinfo' JSON failed: " + err.Error())
	}
	if r.ResultInfo.TotalRecords < 1 {
		return "", "", "", errors.New("no 'dbinfo' setting in FOLIO database")
	}
	value := r.Items[0].Value
	return value.Url, value.User, value.Pass, nil
}

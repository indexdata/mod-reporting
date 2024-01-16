package main

import "os"
import "errors"
import "encoding/json"
import "github.com/indexdata/foliogo"


type settingsValue struct {
	Url string `json:"url"`
	Pass string `json:"pass"`
	User string `json:"user"`
}

type settingsItem struct {
	Value settingsValue `json:"value"`
}

type settingsResultInfo struct {
	TotalRecords int `json:"totalRecords"`
}

type settingsResponse struct {
	Items []settingsItem `json:"items"`
	ResultInfo settingsResultInfo `json:"resultInfo"`
}

func getDbInfo(session foliogo.Session, token string) (string, string, string, error) {
	// If defined, environment variables override the setting from the database
	dburl := os.Getenv("REPORTING_DB_URL")
	dbuser := os.Getenv("REPORTING_DB_USER")
	dbpass := os.Getenv("REPORTING_DB_PASS")
	if dburl != "" && dbuser != "" && dbpass != "" {
		return dburl, dbuser, dbpass, nil
	}

	params := foliogo.RequestParams{ Token: token }
	bytes, err := session.Fetch(`settings/entries?query=scope=="ui-ldp.admin"+and+key=="dbinfo"`, params)
	if err != nil {
		return "", "", "", errors.New("cannot fetch 'dbinfo' from config: " + err.Error())
	}

	var r settingsResponse
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

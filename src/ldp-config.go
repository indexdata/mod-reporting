// handle the /ldp/config and /ldp/config/{key} endpoints
package main

import "fmt"
import "strings"
import "net/http"
import "encoding/json"
import "github.com/indexdata/foliogo"


// Types for what we read from mod-settings
// "General" here means not knowing the structure of the value

type settingsItemGeneral struct {
	Key string `json:"key"`
	Value interface{}
	// We don't care about id (irrelevant) or scope (always "ui-ldp.admin")
}

type settingsResultInfo struct {
	TotalRecords int `json:"totalRecords"`
	// We don't care about diagnostics
}

type settingsResponseGeneral struct {
	Items []settingsItemGeneral `json:"items"`
	ResultInfo settingsResultInfo `json:"resultInfo"`
}


// Types for what we return as /ldp/config

type configItem struct {
	Key string `json:"key"`
	Tenant string `json:"tenant"`
	Value interface{} `json:"value"`
}


// The /ldp/config endpoint only supports GET, with no URL parameters
func handleConfig(w http.ResponseWriter, req *http.Request, server *ModReportingServer) {
	bytes, err := server.folioSession.Fetch(`settings/entries?query=scope=="ui-ldp.admin"`, foliogo.RequestParams{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not fetch from mod-settings: %s", err)
		return
	}

	var r settingsResponseGeneral
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not deserialize JSON from mod-settings: %s", err)
		return
	}

	// XXX in a system with many settings, we might get back less
	// than the full set, in which case we'd need to using paging
	// to accumulate them all. In practice, that's never going to
	// happen. But we could look at resultInfo.totalRecords to
	// determine whether this has happened.

	tenant := server.folioSession.GetTenant()

	config := make([]configItem, len(r.Items))
	for i, item := range(r.Items) {
		config[i] = configItem{
			Key: item.Key,
			Value: item.Value,
			Tenant: tenant,
		}
	}

	bytes, err = json.Marshal(config)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not serialize JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(bytes)
}


// The /ldp/config/{key} endpoint only GET and PUT
func handleConfigKey(w http.ResponseWriter, req *http.Request, server *ModReportingServer) {
	key := strings.Replace(req.URL.Path, "/ldp/config/", "", 1)
	path := `settings/entries?query=scope=="ui-ldp.admin"+and+key=="` + key + `"`
	bytes, err := server.folioSession.Fetch(path, foliogo.RequestParams{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not fetch from mod-settings: %s", err)
		return
	}

	var r settingsResponseGeneral
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not deserialize JSON from mod-settings: %s", err)
		return
	}

	if r.ResultInfo.TotalRecords < 1 {
		// No config entry of that name
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no config item with key '%s'", key)
		return
	}

	item := r.Items[0]
	tenant := server.folioSession.GetTenant()
	config := configItem{
		Key: item.Key,
		Value: item.Value,
		Tenant: tenant,
	}

	bytes, err = json.Marshal(config)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not serialize JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(bytes)
}

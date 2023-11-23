// handle the /ldp/config and /ldp/config/{key} endpoints
package main

import "io"
import "fmt"
import "strings"
import "net/http"
import "encoding/json"
import "github.com/google/uuid"
import "github.com/indexdata/foliogo"


// Types for what we read from mod-settings
// "General" here means not knowing the structure of the value

type settingsItemGeneral struct {
	Id string `json:"id"`
	Scope string `json:"scope"`
	Key string `json:"key"`
	Value interface{}
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
	Value string `json:"value"`
}


// The /ldp/config endpoint only supports GET, with no URL parameters
func handleConfig(w http.ResponseWriter, req *http.Request, server *ModReportingServer) {
	bytes, err := server.folioSession.Fetch0(`settings/entries?query=scope=="ui-ldp.admin"`)
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
		valueString, err := json.Marshal(item.Value)
		if (err != nil) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "could not serialize JSON for value: %s", err)
			return
		}
		config[i] = configItem{
			Key: item.Key,
			Value: string(valueString),
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


func handleConfigKey(w http.ResponseWriter, req *http.Request, server *ModReportingServer) {
	err := underlyingHandleConfigKey(w, req, server)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
	}
}


// The /ldp/config/{key} endpoint only supports GET and PUT
func underlyingHandleConfigKey(w http.ResponseWriter, req *http.Request, server *ModReportingServer) error {
	key := strings.Replace(req.URL.Path, "/ldp/config/", "", 1)
	var bytes []byte
	var err error
	var verb string

	if (req.Method == "PUT") {
		bytes, err = writeConfigKey(w, req, server, key)
		verb = "write to"
	} else {
		// Assume GET
		bytes, err = server.folioSession.Fetch0(`settings/entries?query=scope=="ui-ldp.admin"+and+key=="` + key + `"`)
		verb = "read from"
	}
	if err != nil {
		return fmt.Errorf("could not %s mod-settings: %s", verb, err)
	}

	var r settingsResponseGeneral
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return fmt.Errorf("could not deserialize JSON from mod-settings: %s", err)
	}

	if r.ResultInfo.TotalRecords < 1 {
		return fmt.Errorf("no config item with key '%s'", key)
	}

	item := r.Items[0]
	tenant := server.folioSession.GetTenant()
	valueString, err := json.Marshal(item.Value)
	if (err != nil) {
		return fmt.Errorf("could not serialize JSON for value: %s", err)
	}

	config := configItem{
		Key: item.Key,
		Value: string(valueString),
		Tenant: tenant,
	}

	bytes, err = json.Marshal(config)
	if err != nil {
		return fmt.Errorf("could not serialize JSON: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	return nil
}


func writeConfigKey(w http.ResponseWriter, req *http.Request, server *ModReportingServer, key string) ([]byte, error) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read HTTP request body: %s", err)
	}

	var item configItem
	err = json.Unmarshal(bytes, &item)
	if err != nil {
		return nil, fmt.Errorf("could not deserialize JSON from body: %s", err)
	}
	// fmt.Println("item.Value =", item.Value)

	// XXX For now, assume we're creating a new entry
	// But we need to check if there is already an entry with this key

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("could not generate v4 UUID: %s", err)
	}

	bytes, err = json.Marshal(item.Value)
	if err != nil {
		return nil, fmt.Errorf("could not serialize JSON for value: %s", err)
	}

	var simpleSettingsItem map[string]interface{} = map[string]interface{}{
		"id": id.String(),
		"scope": "ui-ldp.admin",
		"key": key,
		"value": item.Value,
	}
	fmt.Printf("simpleSettingsItem = %+v\n", simpleSettingsItem)
	return server.folioSession.Fetch("settings/entries", foliogo.RequestParams{
		Method: "POST",
		Json: simpleSettingsItem,
	})
}

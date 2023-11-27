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


type handlerFn func(w http.ResponseWriter, req *http.Request, server *ModReportingServer) error;


func handleConfig(w http.ResponseWriter, req *http.Request, server *ModReportingServer) {
	runWithErrorHandling(w, req, server, underlyingHandleConfig)
}


func handleConfigKey(w http.ResponseWriter, req *http.Request, server *ModReportingServer) {
	runWithErrorHandling(w, req, server, underlyingHandleConfigKey)
}


func runWithErrorHandling(w http.ResponseWriter, req *http.Request, server *ModReportingServer, f handlerFn) {
	err := f(w, req, server)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
	}
}


func settingsItemToConfigItem(item settingsItemGeneral, tenant string) (configItem, error) {
	value, ok := item.Value.(string)
	if !ok {
		// mod-settings can contain values of any type: needs serializing
		bytes, err := json.Marshal(item.Value)
		if err != nil {
			return configItem{}, fmt.Errorf("could not serialize value from mod-settings: %s", err)
		}
		value = string(bytes)
	}
	ci := configItem{
		Key: item.Key,
		Value: value,
		Tenant: tenant,
	}
	return ci, nil
}


// The /ldp/config endpoint only supports GET, with no URL parameters
func underlyingHandleConfig(w http.ResponseWriter, req *http.Request, server *ModReportingServer) error {
	bytes, err := server.folioSession.Fetch0(`settings/entries?query=scope=="ui-ldp.admin"`)
	if err != nil {
		return fmt.Errorf("could not fetch from mod-settings: %s", err)
	}

	var r settingsResponseGeneral
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return fmt.Errorf("could not deserialize JSON from mod-settings: %s", err)
	}

	// XXX in a system with many settings, we might get back less
	// than the full set, in which case we'd need to using paging
	// to accumulate them all. In practice, that's never going to
	// happen. But we could look at resultInfo.totalRecords to
	// determine whether this has happened.

	tenant := server.folioSession.GetTenant()
	config := make([]configItem, len(r.Items))
	for i, item := range(r.Items) {
		config[i], err = settingsItemToConfigItem(item, tenant)
		if err != nil {
			return err
		}
	}

	bytes, err = json.Marshal(config)
	if err != nil {
		return fmt.Errorf("could not serialize JSON: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(bytes)
	return nil
}


// The /ldp/config/{key} endpoint only supports GET and PUT
func underlyingHandleConfigKey(w http.ResponseWriter, req *http.Request, server *ModReportingServer) error {
	key := strings.Replace(req.URL.Path, "/ldp/config/", "", 1)
	var bytes []byte
	var err error

	if (req.Method == "PUT") {
		return writeConfigKey(w, req, server, key)
	}
	
	// Assume GET
	bytes, err = server.folioSession.Fetch0(`settings/entries?query=scope=="ui-ldp.admin"+and+key=="` + key + `"`)
	if err != nil {
		return fmt.Errorf("could not read from mod-settings: %s", err)
	}

	var r settingsResponseGeneral
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return fmt.Errorf("could not deserialize JSON %+v from mod-settings: %s", bytes, err)
	}

	if r.ResultInfo.TotalRecords < 1 {
		return fmt.Errorf("no config item with key '%s'", key)
	}

	item := r.Items[0]
	tenant := server.folioSession.GetTenant()
	config, err := settingsItemToConfigItem(item, tenant)
	if err != nil {
		return err
	}

	bytes, err = json.Marshal(config)
	if err != nil {
		return fmt.Errorf("could not serialize JSON: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	return err
}


func writeConfigKey(w http.ResponseWriter, req *http.Request, server *ModReportingServer, key string) (error) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("could not read HTTP request body: %s", err)
	}

	var item configItem
	err = json.Unmarshal(bytes, &item)
	if err != nil {
		return fmt.Errorf("could not deserialize JSON from body: %s", err)
	}
	// fmt.Println("item.Value =", item.Value)

	// Irritatingly, the WSAPI for mod-settings is different if
	// we're creating a new key from if we're replacing an
	// existing one, so we need first to search for an existing
	// record
	bytes, err = server.folioSession.Fetch0(`settings/entries?query=scope=="ui-ldp.admin"+and+key=="` + key + `"`)
	if err != nil {
		return fmt.Errorf("could not read from mod-settings: %s", err)
	}

	var r settingsResponseGeneral
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return fmt.Errorf("could not deserialize JSON %+v from mod-settings: %s", bytes, err)
	}

	var id, method, path string
	if r.ResultInfo.TotalRecords > 0 {
		// We need to PUT to the existing record
		id = r.Items[0].Id
		method = "PUT"
		path = "settings/entries/" + id
	} else {
		dumbId, err2 := uuid.NewRandom()
		if err2 != nil {
			return fmt.Errorf("could not generate v4 UUID: %s", err2)
		}
		id = dumbId.String()
		method = "POST"
		path = "settings/entries"
	}

	var simpleSettingsItem map[string]interface{} = map[string]interface{}{
		"id": id,
		"scope": "ui-ldp.admin",
		"key": key,
		"value": item.Value,
	}
	fmt.Printf("simpleSettingsItem = %+v\n", simpleSettingsItem)
	_, err = server.folioSession.Fetch(path, foliogo.RequestParams{
		Method: method,
		Json: simpleSettingsItem,
	})
	if err != nil {
		return fmt.Errorf("could not write to mod-settings: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	bytes, err = json.Marshal(simpleSettingsItem)
	if err != nil {
		return fmt.Errorf("could not serialize JSON for response: %s", err)
	}
	_, err = w.Write(bytes)
	return err

}

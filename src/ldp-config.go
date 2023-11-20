package main

import "net/http"

func handleConfig(w http.ResponseWriter, req *http.Request, server *ModReportingServer) {
	w.WriteHeader(http.StatusNotImplemented)
}

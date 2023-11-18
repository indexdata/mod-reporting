package main

import "fmt"
import "net/http"
import "time"
import "github.com/MikeTaylor/catlogger"
import "github.com/indexdata/foliogo"

type ModReportingServer struct {
	config *config
	logger *catlogger.Logger
	root string
	folioSession foliogo.Session
	server http.Server
}

func MakeModReportingServer(cfg *config, logger *catlogger.Logger, root string, folioSession foliogo.Session) *ModReportingServer {
	tr := &http.Transport{}
	tr.RegisterProtocol("file", http.NewFileTransport(http.Dir(root)))

	mux := http.NewServeMux()
	var server = ModReportingServer {
		config: cfg,
		logger: logger,
		root: root,
		folioSession: folioSession,
		server: http.Server{
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			Handler: mux,
		},
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { handler(w, r, &server) })
	fs := http.FileServer(http.Dir(root + "/htdocs"))
	mux.Handle("/htdocs/", http.StripPrefix("/htdocs/", fs))
	fs2 := http.FileServer(http.Dir(root + "/htdocs"))
	mux.Handle("/favicon.ico", fs2)

	return &server
}


func (server *ModReportingServer) launch(hostspec string) error {
	server.server.Addr = hostspec
	server.logger.Log("listen", "listening on", hostspec)
	err := server.server.ListenAndServe()
	server.logger.Log("listen", "finished listening on", hostspec)
	return err
}


func handler(w http.ResponseWriter, req *http.Request, server *ModReportingServer) {
	path := req.URL.Path
	server.logger.Log("path", path)

	if (path == "/") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `
This is <a href="https://github.com/indexdata/mod-reporting">mod-reporting</a>. Try:
<ul>
  <li><a href="/admin/health">Health check</a></li>
  <li><a href="/htdocs/">Static area</a></li>
</ul>`);
		return;
	} else if (path == "/admin/health") {
		fmt.Fprintln(w, "Behold! I live!!");
		return;
	} else {
		// Unrecognized
		w.WriteHeader(http.StatusNotFound)
	}
}

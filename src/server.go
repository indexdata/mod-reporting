package main

import "fmt"
import "net/http"
import "time"
import "strings"
import "github.com/MikeTaylor/catlogger"


type handlerFn func(w http.ResponseWriter, req *http.Request, server *ModReportingServer) error;
type sessionFn func(w http.ResponseWriter, req *http.Request, session *ModReportingSession) error;


type ModReportingServer struct {
	config *config
	logger *catlogger.Logger
	root string
	server http.Server
	sessions map[string]*ModReportingSession
}


func MakeModReportingServer(cfg *config, logger *catlogger.Logger, root string) *ModReportingServer {
	tr := &http.Transport{}
	tr.RegisterProtocol("file", http.NewFileTransport(http.Dir(root)))

	mux := http.NewServeMux()
	var server = ModReportingServer {
		config: cfg,
		logger: logger,
		root: root,
		server: http.Server{
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			Handler: mux,
		},
		sessions: map[string]*ModReportingSession{},
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { handler(w, r, &server) })
	fs := http.FileServer(http.Dir(root + "/htdocs"))
	mux.Handle("/htdocs/", http.StripPrefix("/htdocs/", fs))
	fs2 := http.FileServer(http.Dir(root + "/htdocs"))
	mux.Handle("/favicon.ico", fs2)

	return &server
}


func (server *ModReportingServer)Log(cat string, args ...string) {
	server.logger.Log(cat, args...)
}


func (server *ModReportingServer) launch() error {
	cfg := server.config
	hostspec := cfg.Listen.Host + ":" + fmt.Sprint(cfg.Listen.Port)
	server.server.Addr = hostspec
	server.Log("listen", "listening on", hostspec)
	err := server.server.ListenAndServe()
	server.Log("listen", "finished listening on", hostspec)
	return err
}


// We maintain a map of tenant:url to session
func (server *ModReportingServer) findSession(url string, tenant string) (*ModReportingSession, error) {
	key := tenant + ":" + url
	session := server.sessions[key]
	if session != nil {
		return session, nil
	}

	session, err := NewModReportingSession(server, url, tenant)
	if err != nil {
		return nil, fmt.Errorf("could not create session for key '%s': %w", key, err)
	}

	server.sessions[key] = session
	return session, nil
}


func handler(w http.ResponseWriter, req *http.Request, server *ModReportingServer) {
	path := req.URL.Path
	server.Log("path", path)

	if path == "/" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `
This is <a href="https://github.com/indexdata/mod-reporting">mod-reporting</a>. Try:
<ul>
  <li><a href="/admin/health">Health check</a></li>
  <li><a href="/htdocs/">Static area</a></li>
  <li><a href="/ldp/config">Legacy configuration WSAPI</a></li>
  <li><a href="/ldp/db/tables">List tables from reporting database</a></li>
  <li><a href="/ldp/db/columns?schema=folio_users&table=users">List columns for "users" table</a></li>
</ul>`)
		return
	} else if path == "/admin/health" {
		fmt.Fprintln(w, "Behold! I live!!")
		return
	}

	host := req.Header.Get("X-Okapi-Url")
	tenant := req.Header.Get("X-Okapi-Tenant")
	session, err := server.findSession(host, tenant)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not make session: %s\n", err)
		server.Log("error", fmt.Sprintf("%s: %s", req.RequestURI, err.Error()))
		return
	}

	if path == "/ldp/config" {
		sessionWithErrorHandling(w, req, session, handleConfig)
	} else if strings.HasPrefix(path, "/ldp/config/") {
		sessionWithErrorHandling(w, req, session, handleConfigKey)
	} else if path == "/ldp/db/tables" {
		sessionWithErrorHandling(w, req, session, handleTables)
	} else if path == "/ldp/db/columns" {
		sessionWithErrorHandling(w, req, session, handleColumns)
	} else if path == "/ldp/db/query" && req.Method == "POST" {
		sessionWithErrorHandling(w, req, session, handleQuery)
	} else if path == "/ldp/db/reports" && req.Method == "POST" {
		sessionWithErrorHandling(w, req, session, handleReport)
	} else {
		// Unrecognized
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Not found")
	}
}


func runWithErrorHandling(w http.ResponseWriter, req *http.Request, server *ModReportingServer, f handlerFn) {
	err := f(w, req, server)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
		server.Log("error", fmt.Sprintf("%s: %s", req.RequestURI, err.Error()))
	}
}


func sessionWithErrorHandling(w http.ResponseWriter, req *http.Request, session *ModReportingSession, f sessionFn) {
	err := f(w, req, session)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
		session.Log("error", fmt.Sprintf("%s: %s", req.RequestURI, err.Error()))
	}
}

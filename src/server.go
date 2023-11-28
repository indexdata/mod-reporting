package main

import "context"
import "fmt"
import "net/http"
import "time"
import "strings"
import "github.com/MikeTaylor/catlogger"
import "github.com/indexdata/foliogo"
import "github.com/jackc/pgx/v5"


type handlerFn func(w http.ResponseWriter, req *http.Request, server *ModReportingServer) error;


type ModReportingServer struct {
	config *config
	logger *catlogger.Logger
	root string
	folioSession foliogo.Session
	server http.Server
	dbConn *pgx.Conn
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


func (server *ModReportingServer)Log(cat string, args ...string) {
	server.logger.Log(cat, args...)
}


func (server *ModReportingServer) connectDb(url string, user string, pass string) error {
	// For historical reasons, database connection configuration is often JDBCish
	url = strings.Replace(url, "jdbc:postgresql://", "", 1)
	url = strings.Replace(url, "postgres://", "", 1)
	// We may need `?sslmode=require` on the end of the URL.
	conn, err := pgx.Connect(context.Background(), "postgres://" + user + ":" + pass + "@" + url)
	if err != nil {
		return err
	}
	server.dbConn = conn
	return nil
}


func (server *ModReportingServer) launch(hostspec string) error {
	server.server.Addr = hostspec
	server.Log("listen", "listening on", hostspec)
	err := server.server.ListenAndServe()
	server.Log("listen", "finished listening on", hostspec)
	return err
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
  <li><a href="/db/tables">List tables from reporting database</a></li>
</ul>`)
	} else if path == "/admin/health" {
		fmt.Fprintln(w, "Behold! I live!!")
	} else if path == "/ldp/config" {
		runWithErrorHandling(w, req, server, handleConfig)
	} else if strings.HasPrefix(path, "/ldp/config/") {
		runWithErrorHandling(w, req, server, handleConfigKey)
	} else if strings.HasPrefix(path, "/db/tables") {
		runWithErrorHandling(w, req, server, handleTables)
	} else {
		// Unrecognized
		fmt.Fprintln(w, "Not found")
		w.WriteHeader(http.StatusNotFound)
	}
}


func runWithErrorHandling(w http.ResponseWriter, req *http.Request, server *ModReportingServer, f handlerFn) {
	err := f(w, req, server)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
	}
}

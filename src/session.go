package main

import "fmt"
import "github.com/indexdata/foliogo"


type ModReportingSession struct {
	server *ModReportingServer // back-reference
	url string
	tenant string
	folioSession foliogo.Session
	// dbConn *pgxpool.Pool
}


/*
 * There are two valid cases:
 * 1. The url and tenant parameters are defined: we have received a request from Okapi, and the new FOLIO session should be pointed to the specified URL+tenant
 * 2. The url and tenant parameters are not defined: we have a request directly from a browser, curl command or similar, and need ot make a default FOLIO session (which will use environment variables such as OKAPI_URL)
 *
 * There is also one common INvalid case:
 * 3. The url parameter is not defined but the tenant parameter is. This arises when running a curl command copied from a browser session, as the url parameter is added only by Okapi. We explicitly catch this are reject it.
 */

func NewModReportingSession(server *ModReportingServer, url string, tenant string) (*ModReportingSession, error) {
	if url == "" && tenant != "" {
		return nil, fmt.Errorf("no URL provided with tenant: responding to a request with no X-Okapi-Url header?")
	}

	session := ModReportingSession{
		server: server,
		url: url,
		tenant: tenant,
	}

	if url != "" {
		// A request that has arrived via Okapi (or been faked to look that way)
		service := foliogo.NewService(url)
		folioSession, err := service.ResumeSession(tenant)
		session.folioSession = folioSession
		if err != nil {
			return nil, fmt.Errorf("could not resume existing FOLIO session: %w", err)
		}

		return &session, nil
	}

	// Probably a request from command-line, not via Okapi
	// In this case, we use a FOLIO service specified in the environment
	folioSession, err := foliogo.NewDefaultSession()
	if err != nil {
		return nil, fmt.Errorf("could not create new FOLIO session: %w", err)
	}

	session.folioSession = folioSession
	return &session, nil
}


func (session *ModReportingSession)Log(cat string, args ...string) {
	session.server.Log(cat, args...)
}



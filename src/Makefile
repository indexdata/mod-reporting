TARGET=../target/mod-reporting

$(TARGET): main.go configured-server.go config-file.go getdbinfo.go http-error.go server.go session.go ldp-config.go reporting.go
	go build -o $@

run: $(TARGET)
	cd ..; env LOGCAT=listen,path,db,error,sql,op,curl,status,response src/$(TARGET) etc/config.json

run-local: $(TARGET)
	cd ..; env LOGCAT=listen,path,db,error,sql,op,curl,status,response OKAPI_URL=https://folio-snapshot-okapi.dev.folio.org OKAPI_TENANT=diku OKAPI_USER=diku_admin src/$(TARGET) etc/config.json

lint:
	-go vet ./...
	-go vet -vettool=/Users/mike/go/bin/shadow ./...
	-egrep -n '[ 	]+$$' *.go
	-egrep -n 'if +\(' *.go
	-staticcheck ./...
	-errcheck -exclude .errcheck-exclude ./...
	-ineffassign ./...
	-deadcode

test:
	go test -v -coverprofile=c.out ./...
	go tool cover -func=c.out | sed 's/^github.com\/indexdata\/mod-reporting\/src\///'

clean:
	rm -f $(TARGET) c.out


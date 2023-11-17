TARGET=../target/mod-reporting

$(TARGET): main.go config.go http-error.go server.go
	go build -o $@

run: $(TARGET)
	cd ..; src/$(TARGET) etc/config.json

lint:
	-go vet ./...
	-go vet -vettool=/Users/mike/go/bin/shadow ./...
	-staticcheck ./...
	-errcheck -exclude .errcheck-exclude ./...
	-ineffassign ./...
	-deadcode

test:
	go test -coverprofile=c.out ./...
	go tool cover -func=c.out | sed 's/^github.com\/indexdata\/mod-reporting\/src\///'

clean:
	rm -f $(TARGET)


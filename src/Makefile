TARGET=../target/mod-reporting

$(TARGET): main.go config.go http-error.go server.go
	go build -o $@

run: $(TARGET)
	cd ..; src/$(TARGET) etc/config.json

lint:
	go vet ./...

superlint:
	-go vet ./...
	-go vet -vettool=/Users/mike/go/bin/shadow ./...
	-staticcheck ./...
	-errcheck -exclude .errcheck-exclude ./...
	-ineffassign ./...
	-deadcode

test:
	go test ./...

clean:
	rm -f $(TARGET)


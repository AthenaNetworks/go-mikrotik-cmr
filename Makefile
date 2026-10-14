.PHONY: build test vet fmt run example clean

build:
	go build -o bin/cmr-server ./cmd/cmr-server

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

run: build
	./bin/cmr-server -code dev-code -http 127.0.0.1:8080 -v

example:
	go run ./examples/extend

clean:
	rm -rf bin

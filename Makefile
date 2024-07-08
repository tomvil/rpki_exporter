EXPORTER_VERSION=1.0.0
PACKAGES_DIR=compiled_packages

all: test build clean

test:
	go fmt ./
	go fix ./
	go vet -v ./
	staticcheck ./ || true
	golines -w ./
	golangci-lint run
	go mod tidy

build:
	go build -o rpki_exporter -v

clean:
	rm -f rpki_exporter

run:
	go run . --config.file config.example.yaml

compile:
	GOARCH=amd64 GOOS=darwin CGO_ENABLED=0 go build -o ${PACKAGES_DIR}/rpki_exporter-${EXPORTER_VERSION}-darwin
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o ${PACKAGES_DIR}/rpki_exporter-${EXPORTER_VERSION}-linux
	GOARCH=amd64 GOOS=windows CGO_ENABLED=0 go build -o ${PACKAGES_DIR}/rpki_exporter-${EXPORTER_VERSION}-windows

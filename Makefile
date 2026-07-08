BIN_NAME := ogit
DIST_ZIP := dist/ogit-linux-amd64.zip

.PHONY: lint test build build-zip install-server

lint:
	gofmt -l .
	go vet ./...

test:
	go test ./...

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/$(BIN_NAME)-linux-amd64 .

build-zip:
	powershell.exe -NoProfile -ExecutionPolicy Bypass -File ./scripts/build-zip.ps1

install-server:
	sh ./scripts/install-server.sh

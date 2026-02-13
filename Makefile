.PHONY: build
build:
	@sh ./scripts/build.sh

lcu-build: cmd/token
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags=jsoniter -ldflags "-s -w " -o bin/lcu-info.exe cmd/token/main.go

embed-front:
	@sh ./scripts/build.sh --skip-go-build

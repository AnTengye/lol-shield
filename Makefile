.PHONY: build
build: cmd/shield
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags=jsoniter -ldflags "-s -w " -o bin/lol-shield.exe cmd/shield/main.go

lcu-build: cmd/token
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags=jsoniter -ldflags "-s -w " -o bin/lcu-info.exe cmd/token/main.go
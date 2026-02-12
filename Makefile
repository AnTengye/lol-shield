.PHONY: build
build: cmd/shield-v2
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags=jsoniter -ldflags "-s -w " -o bin/lol-shield-v2.exe cmd/shield-v2/main.go

lcu-build: cmd/token
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags=jsoniter -ldflags "-s -w " -o bin/lcu-info.exe cmd/token/main.go

embed-front:
	@powershell -ExecutionPolicy Bypass -File .\scripts\sync_frontend_dist.ps1

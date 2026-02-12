.PHONY: build build-full build-backend frontend-build frontend-sync

build: build-full

frontend-build:
	@cd frontend && npm install && npm run build

frontend-sync:
	@mkdir -p internal/v2/api/web/dist
	@rm -rf internal/v2/api/web/dist/*
	@cp -r frontend/dist/* internal/v2/api/web/dist/

build-full: frontend-build frontend-sync
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags="jsoniter with_frontend" -ldflags "-s -w " -o bin/lol-shield-v2.exe ./cmd/shield-v2

build-backend:
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags="jsoniter no_frontend" -ldflags "-s -w " -o bin/lol-shield-v2-backend.exe ./cmd/shield-v2

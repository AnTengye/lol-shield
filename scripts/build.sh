#!/usr/bin/env sh
set -eu

SKIP_FRONTEND_BUILD=0
SKIP_GO_BUILD=0

for arg in "$@"; do
  case "$arg" in
    --skip-frontend-build) SKIP_FRONTEND_BUILD=1 ;;
    --skip-go-build) SKIP_GO_BUILD=1 ;;
    *)
      echo "unknown arg: $arg" >&2
      exit 1
      ;;
  esac
done

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
FRONTEND_DIR="$ROOT_DIR/frontend"
FRONTEND_DIST="$FRONTEND_DIR/dist"
EMBED_DIST="$ROOT_DIR/internal/client/web/dist"
BIN_DIR="$ROOT_DIR/bin"

build_frontend() {
  if command -v pnpm >/dev/null 2>&1; then
    pnpm --dir "$FRONTEND_DIR" install --frozen-lockfile
    pnpm --dir "$FRONTEND_DIR" build
    return
  fi

  if command -v npm >/dev/null 2>&1; then
    (
      cd "$FRONTEND_DIR"
      npm ci
      npm run build
    )
    return
  fi

  echo "pnpm and npm are both unavailable" >&2
  exit 1
}

if [ "$SKIP_FRONTEND_BUILD" -eq 0 ]; then
  build_frontend
fi

if [ ! -d "$FRONTEND_DIST" ]; then
  echo "frontend dist not found: $FRONTEND_DIST" >&2
  exit 1
fi

rm -rf "$EMBED_DIST"
mkdir -p "$EMBED_DIST"
cp -R "$FRONTEND_DIST"/. "$EMBED_DIST"/

if [ "$SKIP_GO_BUILD" -eq 1 ]; then
  echo "frontend dist synced to $EMBED_DIST"
  exit 0
fi

mkdir -p "$BIN_DIR"
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build \
  -tags jsoniter \
  -ldflags "-s -w" \
  -o "$BIN_DIR/lol-shield.exe" \
  ./cmd/shield/main.go

echo "build done: $BIN_DIR/lol-shield.exe"

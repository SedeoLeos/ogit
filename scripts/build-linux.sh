#!/usr/bin/env sh
set -eu

mkdir -p dist

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags="-s -w" \
  -o dist/ogit-linux-amd64 \
  .

echo "Built dist/ogit-linux-amd64"

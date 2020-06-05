#!/bin/bash
trap onerr ERR
onerr(){ while caller $((n++)); do :; done; }

# # Beware! Overwrites host env.

mkdir -p ./.tmp
export GOPATH="$(pwd)/.tmp"

go mod download
go mod vendor

# Fix to build ginkgo on js.
sed -i 's/build windows/build windows js/g' ./vendor/github.com/onsi/ginkgo/internal/remote/output_interceptor_win.go

export GOOS=js
export GOARCH=wasm
go test -mod=vendor -count=1 -v -exec="$(go env GOROOT)/misc/wasm/go_js_wasm_exec" ./...

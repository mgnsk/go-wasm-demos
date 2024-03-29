#!/bin/bash

set -eo pipefail

root=$1

wasmExecJSPath="$(go env GOROOT)/misc/wasm/wasm_exec.js"

echo "# Generating public dirs..."

for app in $root/cmdwasm/*/; do
	app=$(basename "$app")

	echo "$app"

	mkdir -p "$root/public/$app"
	cp "$root/third_party/js/stats.min.js" "$root/public/$app/stats.min.js"
	cat "$wasmExecJSPath" "$root/template/index.js" >"$root/public/$app/index.js"
	cp "$root/template/index.html" "$root/public/$app/index.html"
done

//go:build js && wasm
// +build js,wasm

package {{ .Package }}

import "encoding/base64"

// IndexJS contains minimal code to spin up go main binary in a browser environment.
// It is base64-encoded should the script contain backticks.
var IndexJS = mustDecodeString(`{{.IndexJS}}`)

func mustDecodeString(s string) string {
	indexJS, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return string(indexJS)
}

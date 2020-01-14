// +build js,wasm

package {{ .Package }}

// IndexJS contains minimal code to spin up go main binary in a browser environment.
// It is base64-encoded should the script contain backticks.
const IndexJS = `
{{ .IndexJS }}
`
//go:build js && wasm
// +build js,wasm

package webgl

import (
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
)

type Buffer struct {
	buffer   js.Value
	bufType  GLType
	drawType GLType
}

func (b *Buffer) JSValue() js.Value {
	return b.buffer
}

// CreateBuffer from js typed array
// Default bufferType should be gl.Types.ArrayBuffer
// Default drawType should be gl.Types.StaticDraw
func CreateBuffer(gl *GL, arr array.TypedArray, bufType, drawType GLType) (*Buffer, error) {
	// TODO check errors
	buffer := gl.Ctx().Call("createBuffer", bufType.JSValue())
	gl.Ctx().Call("bindBuffer", bufType.JSValue(), buffer)
	gl.Ctx().Call("bufferData", bufType.JSValue(), arr.JSValue(), drawType.JSValue())
	gl.Ctx().Call("bindBuffer", bufType.JSValue(), nil)

	return &Buffer{
		buffer:   buffer,
		bufType:  bufType,
		drawType: drawType,
	}, nil
}

type BufferInfo struct {
	NumElements   int
	IndicesBuffer *Buffer
	Attribs       Attribs
}

// TODO move this into objects.go

//func CreateBufferInfoFromData(gl *GL,

func CreateBufferInfo(gl *GL, data ObjectData) (*BufferInfo, error) {
	indicesArray := array.NewTypedArrayFromSlice(data.Indices)

	indicesBuffer, err := CreateBuffer(
		gl,
		indicesArray,
		gl.Types.ElementArrayBuffer,
		gl.Types.StaticDraw,
	)
	if err != nil {
		return nil, err
	}

	attribs, err := CreateAttribs(gl, data)
	if err != nil {
		return nil, err
	}

	return &BufferInfo{
		NumElements:   len(data.Indices),
		IndicesBuffer: indicesBuffer,
		Attribs:       attribs,
	}, nil
}

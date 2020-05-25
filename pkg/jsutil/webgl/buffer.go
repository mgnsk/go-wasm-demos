// +build js,wasm

package webgl

import (
	"syscall/js"

	"github.com/mgnsk/jsutil/array"
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
	buffer := gl.Ctx().Call("createBuffer", bufType)
	gl.Ctx().Call("bindBuffer", bufType, buffer)
	gl.Ctx().Call("bufferData", bufType, arr, drawType)
	gl.Ctx().Call("bindBuffer", bufType, nil)

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
	indicesArray, err := array.CreateTypedArrayFromSlice(data.Indices)
	if err != nil {
		return nil, err
	}

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

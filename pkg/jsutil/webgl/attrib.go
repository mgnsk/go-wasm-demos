// +build js,wasm

package webgl

import (
	"github.com/mgnsk/jsutil/array"
)

// Attrib is shader attribute
type Attrib struct {
	Name string
	// Only needed when creating setters
	//Location      js.Value
	Buffer        *Buffer
	NumComponents int
	Type          GLType
	Normalize     bool
	Offset        int
	Stride        int
}

// Set attrib value from buffer
/*func (p Program) CreateAttribSetters() func(name string, buffer js.Value) {

}*/

// CreateAttrib from array
func CreateAttrib(gl *GL, name string, arr array.TypedArray, numComponents int, typ GLType) (*Attrib, error) {
	buffer, err := CreateBuffer(
		gl,
		arr,
		gl.Types.ArrayBuffer,
		gl.Types.StaticDraw,
	)
	if err != nil {
		return nil, err
	}

	return &Attrib{
		Name:          name,
		NumComponents: numComponents,
		Type:          typ,
		Buffer:        buffer,
	}, nil
}

// Attribs map
type Attribs map[string]*Attrib

func CreateAttribs(gl *GL, data ObjectData) (map[string]*Attrib, error) {
	positionsArray, err := array.CreateTypedArrayFromSlice(data.Positions)
	if err != nil {
		return nil, err
	}

	normalsArray, err := array.CreateTypedArrayFromSlice(data.Normals)
	if err != nil {
		return nil, err
	}

	texcoordsArray, err := array.CreateTypedArrayFromSlice(data.TexCoords)
	if err != nil {
		return nil, err
	}

	positionAttrib, err := CreateAttrib(
		gl,
		"a_position",
		positionsArray,
		3, // numComponents
		gl.Types.Float,
	)
	if err != nil {
		return nil, err
	}

	normalAttrib, err := CreateAttrib(
		gl,
		"a_normal",
		normalsArray,
		3,
		gl.Types.Float,
	)
	if err != nil {
		return nil, err
	}

	texCoordAttrib, err := CreateAttrib(
		gl,
		"a_texcoord",
		texcoordsArray,
		2,
		gl.Types.Float,
	)
	if err != nil {
		return nil, err
	}

	return map[string]*Attrib{
		"a_position": positionAttrib,
		"a_normal":   normalAttrib,
		"a_texcoord": texCoordAttrib,
	}, nil
}

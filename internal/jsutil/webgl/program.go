// +build js,wasm

package webgl

import (
	"fmt"
	"strings"
	"syscall/js"
)

// Program is a shader program.
type Program struct {
	gl             *GL
	program        js.Value
	UniformSetters map[string]func(gl *GL, val interface{})
	AttribMap      map[string]js.Value
	AttribSetters  map[string]func(*Attrib)
}

// JSValue returns the js value.
func (p *Program) JSValue() js.Value {
	return p.program
}

// func (p *Program) SetAttribute(gl *GL, attrib string, val interface{}) {
// 	p.AttribSetters[attrib](val)
// }

func (p *Program) SetAttributes(gl *GL, bufferInfo *BufferInfo) {
	for _, attrib := range bufferInfo.Attribs {
		setter := p.AttribSetters[attrib.Name]
		setter(attrib)
	}
}

func (p *Program) SetBuffersAndAttributes(gl *GL, bufferInfo *BufferInfo) {
	p.SetAttributes(gl, bufferInfo)
	// if bufferInfo.IndicesBuffer
	gl.Ctx().Call("bindBuffer", gl.Types.ElementArrayBuffer, bufferInfo.IndicesBuffer)

}

// CreateProgram creates a new program from shader sources.
func CreateProgram(gl *GL, vertShaderSrc, fragShaderSrc string, attribs []string) (*Program, error) {
	vertShader, err := CreateShader(gl, vertShaderSrc, gl.Types.VertexShader)
	if err != nil {
		return nil, err
	}

	fragShader, err := CreateShader(gl, fragShaderSrc, gl.Types.FragmentShader)
	if err != nil {
		return nil, err
	}

	program := gl.Ctx().Call("createProgram")
	// TODO check gl errors
	gl.Ctx().Call("attachShader", program, vertShader)
	gl.Ctx().Call("attachShader", program, fragShader)

	// Bind shader attributes.
	for idx, attrib := range attribs {
		gl.Ctx().Call("bindAttribLocation", program, idx, attrib)
	}

	gl.Ctx().Call("linkProgram", program)

	// Check the link status
	linked := gl.Ctx().Call("getProgramParameter", program, gl.Types.LinkStatus)
	if !linked.Truthy() {
		lastError := gl.Ctx().Call("getProgramInfoLog", program).String()
		gl.Ctx().Call("deleteProgram", program)
		return nil, fmt.Errorf("error linking program: %s", lastError)
	}

	attribMap, setters := createAttribsAndSetters(gl, program)
	// TODO error handling for setters
	return &Program{
		program:        program,
		UniformSetters: createUniformSetters(gl, program),
		AttribMap:      attribMap,
		AttribSetters:  setters,
	}, nil
}

// TODO clean up the code to make it not look like a javascript wrapper
// or use g3n

func createUniformSetters(gl *GL, program js.Value) map[string]func(g *GL, val interface{}) {
	var textureUnit int

	createUniformSetter := func(uniformInfo js.Value) func(gl *GL, val interface{}) {

		name := uniformInfo.Get("name").String()

		location := gl.Ctx().Call("getUniformLocation", program, name)
		if !location.InstanceOf(js.Global().Get("WebGLUniformLocation")) {
			panic("expected location")
		}

		size := uniformInfo.Get("size").Int()

		isArray := size > 1 && strings.HasPrefix(name, "[0]")

		switch GLType(uniformInfo.Get("type").Int()) {
		case gl.Types.Float:
			if isArray {
				return func(g *GL, val interface{}) {
					gl.Ctx().Call("uniform1v", location, val)
				}
			} else {
				return func(gl *GL, val interface{}) {
					gl.Ctx().Call("uniform1f", location, val)
				}
			}
		case gl.Types.FloatVec2:
			return func(g *GL, val interface{}) {
				gl.Ctx().Call("uniform2fv", location, val)
			}
		case gl.Types.FloatVec3:
			return func(g *GL, val interface{}) {
				gl.Ctx().Call("uniform3fv", location, val)
			}
		case gl.Types.FloatVec4:
			return func(g *GL, val interface{}) {
				gl.Ctx().Call("uniform4fv", location, val)
			}
		case gl.Types.Int:
			if isArray {
				return func(g *GL, val interface{}) {
					gl.Ctx().Call("uniform1iv", location, val)
				}
			} else {
				return func(g *GL, val interface{}) {
					gl.Ctx().Call("uniform1i", location, val)
				}
			}
		case gl.Types.Bool:
			return func(g *GL, val interface{}) {
				gl.Ctx().Call("uniform1iv", location, val)
			}
		case gl.Types.IntVec2:
			fallthrough
		case gl.Types.BoolVec2:
			return func(g *GL, val interface{}) {
				gl.Ctx().Call("uniform2iv", location, val)
			}
		case gl.Types.IntVec3:
			fallthrough
		case gl.Types.BoolVec3:
			return func(g *GL, val interface{}) {
				gl.Ctx().Call("uniform3iv", location, val)
			}
		case gl.Types.IntVec4:
			fallthrough
		case gl.Types.BoolVec4:
			return func(g *GL, val interface{}) {
				gl.Ctx().Call("uniform4iv", location, val)
			}
		case gl.Types.FloatMat2:
			return func(g *GL, val interface{}) {
				gl.Ctx().Call("uniformMatrix2fv", location, false, val)
			}
		case gl.Types.FloatMat3:
			return func(g *GL, val interface{}) {
				gl.Ctx().Call("uniformMatrix3fv", location, false, val)
			}
		case gl.Types.FloatMat4:
			return func(g *GL, val interface{}) {
				gl.Ctx().Call("uniformMatrix4fv", location, false, val)
			}
		case gl.Types.Sampler2D:
			fallthrough
		case gl.Types.SamplerCube:
			bindPoint := gl.getBindPointForSamplerType(GLType(uniformInfo.Get("type").Int()))

			var units []int

			if isArray {
				for i := 0; i < uniformInfo.Get("size").Int(); i++ {
					textureUnit++
					units = append(units, textureUnit)
				}

				gl.Ctx().Call("uniform1iv", location, units)

				return func(g *GL, val interface{}) {
					if v, ok := val.([]interface{}); ok {
						for idx, texture := range v {
							gl.Ctx().Call("activeTexture", int(gl.Types.Texture0)+units[idx])
							gl.Ctx().Call("bindTexture", bindPoint, texture)
						}
					} else {
						panic("only []interface{} supported SamplerCube")
					}
				}
			}

			textureUnit++

			return func(g *GL, val interface{}) {
				gl.Ctx().Call("uniform1i", location, textureUnit)
				gl.Ctx().Call("activeTexture", int(gl.Types.Texture0)+textureUnit)
				gl.Ctx().Call("bindTexture", bindPoint, val)
			}
		default:
			panic("Invalid uniform type")
		}
	}

	uniformSetters := make(map[string]func(gl *GL, val interface{}))

	numUniforms := gl.Ctx().Call("getProgramParameter", program, gl.Types.ActiveUniforms).Int()
	for i := 0; i < numUniforms; i++ {
		uniformInfo := gl.Ctx().Call("getActiveUniform", program, i)
		if !uniformInfo.Truthy() {
			panic("Error getting uniform info")
		}

		setter := createUniformSetter(uniformInfo)
		name := strings.TrimPrefix(uniformInfo.Get("name").String(), "[0]")
		uniformSetters[name] = setter
	}

	return uniformSetters
}

func createAttribsAndSetters(gl *GL, program js.Value) (map[string]js.Value, map[string]func(a *Attrib)) {
	createAttribSetter := func(location js.Value) func(a *Attrib) {
		return func(a *Attrib) {
			gl.Ctx().Call("bindBuffer", gl.Types.ArrayBuffer, a.Buffer)

			// turn on getting data out of a buffer for this attribute
			gl.Ctx().Call("enableVertexAttribArray", location)
			// Point an attribute to the currently bound VBO
			gl.Ctx().Call(
				"vertexAttribPointer",
				location,
				a.NumComponents, // or a.Size
				a.Type,          // default gl.FLOAT
				a.Normalize,
				a.Stride,
				a.Offset,
			)
		}
	}

	attribs := make(map[string]js.Value)
	attribSetters := make(map[string]func(a *Attrib))

	numAttribs := gl.Ctx().Call("getProgramParameter", program, gl.Types.ActiveAttributes).Int()
	for i := 0; i < numAttribs; i++ {
		attribInfo := gl.Ctx().Call("getActiveAttrib", program, i)
		if !attribInfo.Truthy() {
			panic("Error getting attributes")
		}

		attribName := attribInfo.Get("name").String()

		attribs[attribName] = attribInfo

		location := gl.Ctx().Call("getAttribLocation", program, attribName)

		attribSetters[attribName] = createAttribSetter(location)
	}

	return attribs, attribSetters
}

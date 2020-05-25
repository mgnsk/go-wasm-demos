// +build js,wasm

package webgl

import (
	"fmt"
	"syscall/js"
)

// Shader represents a WebGL shader.
type Shader struct {
	shader js.Value
}

// JSValue the js value.
func (s Shader) JSValue() js.Value {
	return s.shader
}

// CreateShader creates and compiles a WebGLShader.
func CreateShader(gl *GL, src string, typ GLType) (s Shader, err error) {
	shader := gl.Ctx().Call("createShader", typ)
	gl.Ctx().Call("shaderSource", shader, src)
	gl.Ctx().Call("compileShader", shader)

	// Check the compile status
	compiled := gl.Ctx().Call("getShaderParameter", shader, gl.Types.CompileStatus)
	if !compiled.Truthy() {
		lastError := gl.Ctx().Call("getShaderInfoLog", shader).String()
		gl.Ctx().Call("deleteShader", shader)
		err = fmt.Errorf("Error compiling shader: %s", lastError)
		return
	}

	s = Shader{
		shader: shader,
	}
	return
}

type Uniform struct {
	name     string
	location js.Value
	size     int
	glType   GLType
}

func NewUniform(name string, location js.Value, size int, glType GLType) Uniform {
	return Uniform{
		name:     name,
		location: location,
		size:     size,
		glType:   glType,
	}
}

func (u Uniform) Location() js.Value {
	return u.location
}

//func (u Uniform) Set(gl *GL, val interface{})

type ShaderProgram struct {
	program  js.Value
	Uniforms map[string]Uniform
}

func (p *ShaderProgram) JSValue() js.Value {
	return p.program
}

func CreateShaderProgram(gl *GL, vertShader Shader, fragShader Shader) (p ShaderProgram, err error) {
	program := gl.Ctx().Call("createProgram")
	// TODO check gl errors
	gl.Ctx().Call("attachShader", program, vertShader)
	gl.Ctx().Call("attachShader", program, fragShader)

	// Bind shader attributes.
	// for idx, attrib := range attribs {
	// 	gl.Ctx().Call("bindAttribLocation", program, idx, attrib)
	// }

	gl.Ctx().Call("linkProgram", program)

	// Check the link status
	linked := gl.Ctx().Call("getProgramParameter", program, gl.Types.LinkStatus)
	if !linked.Truthy() {
		lastError := gl.Ctx().Call("getProgramInfoLog", program).String()
		gl.Ctx().Call("deleteProgram", program)
		err = fmt.Errorf("error linking program: %s", lastError)
		return
	}

	uniforms := make(map[string]Uniform)

	// Fetch all uniforms from vertex shader.
	numUniforms := gl.Ctx().Call("getProgramParameter", program, gl.Types.ActiveUniforms).Int()
	for i := 0; i < numUniforms; i++ {
		uniformInfo := gl.Ctx().Call("getActiveUniform", program, i)
		if !uniformInfo.Truthy() {
			err = fmt.Errorf("error getting uniform info at index %d", i)
			return
		}

		name := uniformInfo.Get("name").String()
		location := gl.Ctx().Call("getUniformLocation", program, name)
		if !location.InstanceOf(js.Global().Get("WebGLUniformLocation")) {
			err = fmt.Errorf("expected WebGLUniformLocation")
			return
		}

		uniforms[name] = Uniform{
			name:     name,
			location: location,
			size:     uniformInfo.Get("size").Int(),
			glType:   GLType(uniformInfo.Get("type").Int()),
		}
		//	isArray := size > 1 && strings.HasPrefix(name, "[0]")

		//spew.Dump(name)
		//spew.Dump(size)

		//	setter := createUniformSetter(uniformInfo)
		//	name := strings.TrimPrefix(uniformInfo.Get("name").String(), "[0]")
		//uniformSetters[name] = setter
	}

	return ShaderProgram{
		program:  program,
		Uniforms: uniforms,
	}, nil
}

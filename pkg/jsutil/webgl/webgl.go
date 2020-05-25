// +build js,wasm

package webgl

import (
	"errors"
	"fmt"
	"syscall/js"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/mgnsk/jsutil/array"
)

type GLType int

func (t GLType) JSValue() js.Value {
	return js.ValueOf(int(t))
}

// GLTypes provides WebGL bindings.
type GLTypes struct {
	StaticDraw         GLType
	ArrayBuffer        GLType
	ElementArrayBuffer GLType
	VertexShader       GLType
	FragmentShader     GLType
	DepthTest          GLType
	ColorBufferBit     GLType
	DepthBufferBit     GLType
	Triangles          GLType
	UnsignedShort      GLType
	UnsignedByte       GLType
	LEqual             GLType
	LineLoop           GLType
	CompileStatus      GLType
	LinkStatus         GLType
	Float              GLType
	FloatVec2          GLType
	FloatVec3          GLType
	FloatVec4          GLType
	Int                GLType
	IntVec2            GLType
	IntVec3            GLType
	IntVec4            GLType
	Bool               GLType
	BoolVec2           GLType
	BoolVec3           GLType
	BoolVec4           GLType
	FloatMat2          GLType
	FloatMat3          GLType
	FloatMat4          GLType
	Sampler2D          GLType
	SamplerCube        GLType
	Texture2D          GLType
	TextureCubeMap     GLType
	Texture0           GLType
	ActiveUniforms     GLType
	ActiveAttributes   GLType
	Rgba               GLType
	TextureMinFilter   GLType
	TextureMagFilter   GLType
	Nearest            GLType
	Lequal             GLType
}

// GL wrapper for WebGL
type GL struct {
	ctx    js.Value
	Types  GLTypes
	Width  GLType
	Height GLType
}

// GLContext specifies interface for calling the underlying GL implementation.
type GLContext interface {
	Call(m string, args ...interface{}) js.Value
}

// Ctx of webgl
func (gl *GL) Ctx() js.Value {
	return gl.ctx
}

// NewGL tries to get a new context
func NewGL(canvas js.Value) (*GL, error) {
	gl := &GL{}

	ctx := canvas.Call("getContext", "webgl2")

	if ctx == js.Undefined() {
		panic("WebGL version 2 (OpenGL ES 3.0) support missing")
	}

	// once again
	if ctx == js.Null() {
		return nil, errors.New("WebGL unavailable")
	}

	gl.ctx = ctx

	gl.Types.StaticDraw = GLType(ctx.Get("STATIC_DRAW").Int())
	gl.Types.ArrayBuffer = GLType(ctx.Get("ARRAY_BUFFER").Int())
	gl.Types.ElementArrayBuffer = GLType(ctx.Get("ELEMENT_ARRAY_BUFFER").Int())
	gl.Types.VertexShader = GLType(ctx.Get("VERTEX_SHADER").Int())
	gl.Types.FragmentShader = GLType(ctx.Get("FRAGMENT_SHADER").Int())
	gl.Types.DepthTest = GLType(ctx.Get("DEPTH_TEST").Int())
	gl.Types.ColorBufferBit = GLType(ctx.Get("COLOR_BUFFER_BIT").Int())
	gl.Types.Triangles = GLType(ctx.Get("TRIANGLES").Int())
	gl.Types.UnsignedShort = GLType(ctx.Get("UNSIGNED_SHORT").Int())
	gl.Types.UnsignedByte = GLType(ctx.Get("UNSIGNED_BYTE").Int())
	gl.Types.LEqual = GLType(ctx.Get("LEQUAL").Int())
	gl.Types.DepthBufferBit = GLType(ctx.Get("DEPTH_BUFFER_BIT").Int())
	gl.Types.LineLoop = GLType(ctx.Get("LINE_LOOP").Int())
	gl.Types.CompileStatus = GLType(ctx.Get("COMPILE_STATUS").Int())
	gl.Types.LinkStatus = GLType(ctx.Get("LINK_STATUS").Int())
	gl.Types.Float = GLType(ctx.Get("FLOAT").Int())
	gl.Types.FloatVec2 = GLType(ctx.Get("FLOAT_VEC2").Int())
	gl.Types.FloatVec3 = GLType(ctx.Get("FLOAT_VEC3").Int())
	gl.Types.FloatVec4 = GLType(ctx.Get("FLOAT_VEC4").Int())
	gl.Types.Int = GLType(ctx.Get("INT").Int())
	gl.Types.IntVec2 = GLType(ctx.Get("INT_VEC2").Int())
	gl.Types.IntVec3 = GLType(ctx.Get("INT_VEC3").Int())
	gl.Types.IntVec4 = GLType(ctx.Get("INT_VEC4").Int())
	gl.Types.Bool = GLType(ctx.Get("BOOL").Int())
	gl.Types.BoolVec2 = GLType(ctx.Get("BOOL_VEC2").Int())
	gl.Types.BoolVec3 = GLType(ctx.Get("BOOL_VEC3").Int())
	gl.Types.BoolVec4 = GLType(ctx.Get("BOOL_VEC4").Int())
	gl.Types.FloatMat2 = GLType(ctx.Get("FLOAT_MAT2").Int())
	gl.Types.FloatMat3 = GLType(ctx.Get("FLOAT_MAT3").Int())
	gl.Types.FloatMat4 = GLType(ctx.Get("FLOAT_MAT4").Int())
	gl.Types.Sampler2D = GLType(ctx.Get("SAMPLER_2D").Int())
	gl.Types.SamplerCube = GLType(ctx.Get("SAMPLER_CUBE").Int())
	gl.Types.Texture2D = GLType(ctx.Get("TEXTURE_2D").Int())
	gl.Types.TextureCubeMap = GLType(ctx.Get("TEXTURE_CUBE_MAP").Int())
	gl.Types.Texture0 = GLType(ctx.Get("TEXTURE0").Int())
	gl.Types.ActiveUniforms = GLType(ctx.Get("ACTIVE_UNIFORMS").Int())
	gl.Types.ActiveAttributes = GLType(ctx.Get("ACTIVE_ATTRIBUTES").Int())
	gl.Types.Rgba = GLType(ctx.Get("RGBA").Int())
	gl.Types.TextureMinFilter = GLType(ctx.Get("TEXTURE_MIN_FILTER").Int())
	gl.Types.TextureMagFilter = GLType(ctx.Get("TEXTURE_MAG_FILTER").Int())
	gl.Types.Nearest = GLType(ctx.Get("NEAREST").Int())
	gl.Types.Lequal = GLType(ctx.Get("LEQUAL").Int())

	return gl, nil
}

// TODO

// type WebGLRenderingContext struct {

// }

// type WebGLProgram struct {

// }

// func CreateProgram(gl *GL)

// func (ctx WebGLRenderingContext) UseProgram()

// UniformMatrix4fv function
func (gl *GL) UniformMatrix4fv(uniform js.Value, transform mgl32.Mat4) {
	var buf *[16]float64
	buf = (*[16]float64)(unsafe.Pointer(&transform))
	arr, err := array.CreateTypedArrayFromSlice(buf[:])
	if err != nil {
		panic(fmt.Errorf("error creating TypedArray: %s", err))
	}

	gl.Ctx().Call("uniformMatrix4fv", uniform, false, arr.JSValue())
}

// Set up shaders

type UniformInfo struct {
	Name string
	Size GLType
}

// type Uniform struct {
// 	Name     string
// 	Value    js.Value
// 	Type     GLType
// 	Size     int
// 	Location js.Value
// }

func (gl *GL) getBindPointForSamplerType(typ GLType) GLType {
	if typ == gl.Types.Sampler2D {
		return gl.Types.Texture2D
	} else if typ == gl.Types.SamplerCube {
		return gl.Types.TextureCubeMap
	}

	panic("Invalid type")
}

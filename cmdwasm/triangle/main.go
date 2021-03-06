// +build js,wasm

package main

import (
	"fmt"
	"math/rand"
	"syscall/js"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/mgnsk/go-wasm-demos/gen/shader"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/webgl"
)

var (
	width  int
	height int
)

func init() {
	rand.Seed(time.Now().UnixNano()) // initialize global pseudo random generator
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("button clicked")
		// cb.Release() // release the function if the button will not be clicked again
		return nil
	})
	js.Global().Get("document").Call("getElementById", "gocanvas").Call("addEventListener", "click", cb)

	// Init Canvas stuff
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "gocanvas")
	width = doc.Get("body").Get("clientWidth").Int()
	height = doc.Get("body").Get("clientHeight").Int()
	canvas.Set("width", width)
	canvas.Set("height", height)

	gl, err := webgl.NewGL(canvas)
	if err != nil {
		js.Global().Call("alert", err.Error())
		panic(err)
	}

	//	s := gl.getParameter(gl.Ctx().GSHADING_LANGUAGE_VERSION)

	// WebGL GLSL ES 3.00 (OpenGL ES GLSL ES 3.0 Chromium)
	s := gl.Ctx().Call("getParameter", gl.Ctx().Get("SHADING_LANGUAGE_VERSION"))

	jsutil.ConsoleLog(s)

	//// VERTEX BUFFER ////
	verticesNative := []float32{
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,
	}
	vertexArr := array.Must(array.CreateTypedArrayFromSlice(verticesNative))
	vertexBuffer, err := webgl.CreateBuffer(gl, vertexArr, gl.Types.ArrayBuffer, gl.Types.StaticDraw)
	check(err)

	jsutil.ConsoleLog(vertexBuffer)

	//// INDEX BUFFER ////
	indicesNative := []uint32{
		2, 1, 0,
	}
	indexArr := array.Must(array.CreateTypedArrayFromSlice(indicesNative))
	indexBuffer, err := webgl.CreateBuffer(gl, indexArr, gl.Types.ElementArrayBuffer, gl.Types.StaticDraw)
	check(err)

	jsutil.ConsoleLog(indexBuffer)

	// 	//// Shaders ////

	vertShader := shader.Shaders["triangle/triangle.vert"]
	fragShader := shader.Shaders["triangle/triangle.frag"]

	attribs := []string{"coordinates"}

	spew.Dump(vertShader)

	triangleProgram, err := webgl.CreateProgram(gl, vertShader, fragShader, attribs)
	check(err)

	spew.Dump(triangleProgram)
	jsutil.ConsoleLog(triangleProgram)

	// Clear the canvas
	gl.Ctx().Call("clearColor", 0.5, 0.5, 0.5, 0.9)
	gl.Ctx().Call("clear", gl.Types.ColorBufferBit)

	// Enable the depth test
	gl.Ctx().Call("enable", gl.Types.DepthTest)

	// Set the view port
	gl.Ctx().Call("viewport", 0, 0, width, height)

	gl.Ctx().Call("useProgram", triangleProgram)

	//// Associating shaders to buffer objects ////

	// Bind vertex buffer object
	gl.Ctx().Call("bindBuffer", gl.Types.ArrayBuffer, vertexBuffer)

	// Bind index buffer object
	gl.Ctx().Call("bindBuffer", gl.Types.ElementArrayBuffer, indexBuffer)

	// Get the attribute location
	coord := gl.Ctx().Call("getAttribLocation", triangleProgram, "coordinates")

	// Point an attribute to the currently bound VBO
	gl.Ctx().Call("vertexAttribPointer", coord, 3, gl.Types.Float, false, 0, 0)

	// Enable the attribute
	gl.Ctx().Call("enableVertexAttribArray", coord)

	// positionAttrib, err := CreateAttrib(
	// 	gl,
	// 	"a_position",
	// 	positionsArray,
	// 	3, // numComponents
	// 	gl.Types.Float,
	// )
	// if err != nil {
	// 	return nil, err
	// }

	fpsStats := js.Global().Get("Stats").New()
	fpsStats.Call("showPanel", 0)
	js.Global().Get("document").Get("body").Call("appendChild", fpsStats.Get("dom"))

	var tmark float32
	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fpsStats.Call("begin")

		now := float32(args[0].Float())
		dt := now - tmark
		tmark = now

		_ = dt

		width := canvas.Get("clientWidth").Int()
		height := canvas.Get("clientHeight").Int()

		if canvas.Get("width").Int() != width || canvas.Get("height").Int() != height {
			canvas.Set("width", width)
			canvas.Set("height", height)
		}

		//// Drawing the triangle ////

		// Clear the canvas
		gl.Ctx().Call("clearColor", 0.5, 0.5, 0.5, 0.9)
		gl.Ctx().Call("clear", gl.Types.ColorBufferBit)

		// Enable the depth test
		gl.Ctx().Call("enable", gl.Types.DepthTest)

		// Set the view port
		gl.Ctx().Call("viewport", 0, 0, width, height)

		// Draw the triangle
		gl.Ctx().Call("drawElements", gl.Types.Triangles, len(indicesNative), gl.Types.UnsignedShort, 0)

		fpsStats.Call("end")
		// Call next frame
		js.Global().Call("requestAnimationFrame", renderFrame)

		return nil
	})

	defer renderFrame.Release()

	js.Global().Call("requestAnimationFrame", renderFrame)

	select {}
}

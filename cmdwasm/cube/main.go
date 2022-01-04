//go:build js && wasm
// +build js,wasm

package main

import (
	_ "embed"
	"fmt"
	"math/rand"
	"syscall/js"
	"time"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/mgnsk/go-wasm-demos/internal/gfx"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/webgl"
)

var (
	gl js.Value
	//go:embed shader/cube.vert
	vertShader string
	//go:embed shader/cube.frag
	fragShader string
)

// https://www.tutorialspoint.com/webgl/webgl_cube_rotation.htm //
var verticesNative = []float32{
	-1, -1, -1, 1, -1, -1, 1, 1, -1, -1, 1, -1,
	-1, -1, 1, 1, -1, 1, 1, 1, 1, -1, 1, 1,
	-1, -1, -1, -1, 1, -1, -1, 1, 1, -1, -1, 1,
	1, -1, -1, 1, 1, -1, 1, 1, 1, 1, -1, 1,
	-1, -1, -1, -1, -1, 1, 1, -1, 1, 1, -1, -1,
	-1, 1, -1, -1, 1, 1, 1, 1, 1, 1, 1, -1,
}
var colorsNative = []float32{
	5, 3, 7, 5, 3, 7, 5, 3, 7, 5, 3, 7,
	1, 1, 3, 1, 1, 3, 1, 1, 3, 1, 1, 3,
	0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1,
	1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0,
	1, 1, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0,
	0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0,
}
var indicesNative = []uint16{
	0, 1, 2, 0, 2, 3, 4, 5, 6, 4, 6, 7,
	8, 9, 10, 8, 10, 11, 12, 13, 14, 12, 14, 15,
	16, 17, 18, 16, 18, 19, 20, 21, 22, 20, 22, 23,
}

var (
	width  int
	height int
)

func init() {
	rand.Seed(time.Now().UnixNano()) // initialize global pseudo random generator
}

// TODO haven't really thought of panic handling yet.
func check(err error) {
	if err != nil {
		panic(err)
	}
}

func float32SliceFromMat4(m mgl32.Mat4) []float32 {
	var p *[16]float32
	p = (*[16]float32)(unsafe.Pointer(&m))
	return p[:]
}

func main() {
	// Sanity check.
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

	jsutil.ConsoleLog(canvas)

	gl, err := webgl.NewGL(canvas)
	if err != nil {
		jsutil.AlertPanic(err)
	}

	// WebGL GLSL ES 3.00 (OpenGL ES GLSL ES 3.0 Chromium)
	// or WebGL GLSL ES 3.00 on firefox.
	s := gl.Ctx().Call("getParameter", gl.Ctx().Get("SHADING_LANGUAGE_VERSION"))

	jsutil.ConsoleLog(s)

	// Convert buffers to JS TypedArrays
	colors := array.NewTypedArrayFromSlice(colorsNative)
	vertices := array.NewTypedArrayFromSlice(verticesNative)
	indices := array.NewTypedArrayFromSlice(indicesNative)

	// Create vertex buffer
	vertexBuffer, err := webgl.CreateBuffer(gl, vertices, gl.Types.ArrayBuffer, gl.Types.StaticDraw)
	check(err)

	// Create color buffer
	colorBuffer, err := webgl.CreateBuffer(gl, colors, gl.Types.ArrayBuffer, gl.Types.StaticDraw)
	check(err)

	// Create index buffer
	indexBuffer, err := webgl.CreateBuffer(gl, indices, gl.Types.ElementArrayBuffer, gl.Types.StaticDraw)
	check(err)

	// * Shaders *

	// Create a vertex shader object
	vertShader, err := webgl.CreateShader(gl, vertShader, gl.Types.VertexShader)
	check(err)

	// Create fragment shader object
	fragShader, err := webgl.CreateShader(gl, fragShader, gl.Types.FragmentShader)
	check(err)

	shaderProgram, err := webgl.CreateShaderProgram(gl, vertShader, fragShader)
	check(err)

	//jsutil.ConsoleLog(p.JSValue())
	_ = spew.Dump

	jsutil.ConsoleLog(shaderProgram.Uniforms["Pmatrix"].Location())
	//	spew.Dump(p.Uniforms)

	// Associate attributes to vertex shader

	gl.Ctx().Call("bindBuffer", gl.Types.ArrayBuffer.JSValue(), vertexBuffer.JSValue())
	position := gl.Ctx().Call("getAttribLocation", shaderProgram.JSValue(), "position")
	gl.Ctx().Call("vertexAttribPointer", position, 3, gl.Types.Float.JSValue(), false, 0, 0)
	gl.Ctx().Call("enableVertexAttribArray", position)

	gl.Ctx().Call("bindBuffer", gl.Types.ArrayBuffer.JSValue(), colorBuffer.JSValue())
	color := gl.Ctx().Call("getAttribLocation", shaderProgram.JSValue(), "color")
	gl.Ctx().Call("vertexAttribPointer", color, 3, gl.Types.Float.JSValue(), false, 0, 0)
	gl.Ctx().Call("enableVertexAttribArray", color)

	gl.Ctx().Call("useProgram", shaderProgram.JSValue())

	// Set WebGL properties
	gl.Ctx().Call("clearColor", 0.5, 0.5, 0.5, 0.9) // Color the screen is cleared to
	gl.Ctx().Call("clearDepth", 1.0)                // Z value that is set to the Depth buffer every frame
	gl.Ctx().Call("viewport", 0, 0, width, height)  // Viewport size
	gl.Ctx().Call("depthFunc", gl.Types.Lequal.JSValue())

	// Bind to element array for draw function
	gl.Ctx().Call("bindBuffer", gl.Types.ElementArrayBuffer.JSValue(), indexBuffer.JSValue())

	fpsStats := js.Global().Get("Stats").New()
	fpsStats.Call("showPanel", 0)
	js.Global().Get("document").Get("body").Call("appendChild", fpsStats.Get("dom"))

	camera := gfx.NewCamera(
		mgl32.Vec3{3.0, 3.0, 30},
		mgl32.Vec3{0.0, 0.0, 0.0},
		mgl32.Vec3{0.0, 1.0, 0.0},
		mgl32.DegToRad(45.0),
		1,
		float32(width)/float32(height),
	)

	var keydown js.Func
	keydown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		switch args[0].Get("code").String() {
		case "ArrowUp":
			camera.Rotate(gfx.RotateDown)
		case "ArrowDown":
			camera.Rotate(gfx.RotateUp)
		case "ArrowLeft":
			camera.Rotate(gfx.RotateLeft)
		case "ArrowRight":
			camera.Rotate(gfx.RotateRight)
		case "KeyW":
			camera.Move(gfx.MoveForward)
		case "KeyS":
			camera.Move(gfx.MoveBack)
		case "KeyA":
			camera.Move(gfx.MoveLeft)
		case "KeyD":
			camera.Move(gfx.MoveRight)
		case "KeyQ":
			camera.Roll(gfx.RollLeft)
		case "KeyE":
			camera.Roll(gfx.RollRight)

		case "KeyP":
			panic("KeyP pressed")
		}
		// cb.Release() // release the function if the button will not be clicked again
		return nil
	})
	js.Global().Get("document").Call("addEventListener", "keydown", keydown)

	movMatrix := mgl32.Ident4()
	var rotation float32
	var tmark float32
	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fpsStats.Call("begin")

		now := float32(args[0].Float())
		dt := now - tmark
		tmark = now

		rotation = rotation + float32(dt)/1000

		width := canvas.Get("clientWidth").Int()
		height := canvas.Get("clientHeight").Int()

		if canvas.Get("width").Int() != width || canvas.Get("height").Int() != height {
			canvas.Set("width", width)
			canvas.Set("height", height)
		}

		// * Create Matrixes *

		// Generate and apply projection and view matrices
		projMatrix, viewMatrix := camera.Projection(), camera.View()

		//	spew.Dump(projMatrix, viewMatrix)

		typedProjMatrixBuffer := array.NewTypedArrayFromSlice(float32SliceFromMat4(projMatrix))
		typedViewMatrixBuffer := array.NewTypedArrayFromSlice(float32SliceFromMat4(viewMatrix))

		gl.Ctx().Call("uniformMatrix4fv", shaderProgram.Uniforms["Pmatrix"].Location(), false, typedProjMatrixBuffer.JSValue())
		gl.Ctx().Call("uniformMatrix4fv", shaderProgram.Uniforms["Vmatrix"].Location(), false, typedViewMatrixBuffer.JSValue())

		// // Do new model matrix calculations
		movMatrix = mgl32.HomogRotate3DX(0.5 * rotation)
		movMatrix = movMatrix.Mul4(mgl32.HomogRotate3DY(0.3 * rotation))
		movMatrix = movMatrix.Mul4(mgl32.HomogRotate3DZ(0.2 * rotation))

		// Convert model matrix to a JS TypedArray
		typedModelMatrixBuffer := array.NewTypedArrayFromSlice(float32SliceFromMat4(movMatrix))

		// Apply the model matrix
		gl.Ctx().Call("uniformMatrix4fv", shaderProgram.Uniforms["Mmatrix"].Location(), false, typedModelMatrixBuffer.JSValue())

		// Clear the screen
		gl.Ctx().Call("enable", gl.Types.DepthTest.JSValue())
		gl.Ctx().Call("clear", gl.Types.ColorBufferBit.JSValue())
		gl.Ctx().Call("clear", gl.Types.DepthBufferBit.JSValue())

		// Draw the cube
		gl.Ctx().Call("drawElements", gl.Types.Triangles.JSValue(), len(indicesNative), gl.Types.UnsignedShort.JSValue(), 0)

		fpsStats.Call("end")
		// Call next frame
		js.Global().Call("requestAnimationFrame", renderFrame)

		return nil
	})

	defer renderFrame.Release()

	js.Global().Call("requestAnimationFrame", renderFrame)

	select {}
}

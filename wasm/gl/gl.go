package gl

import (
	"fmt"
	"syscall/js"
)

// Constants for holding GL constants (ex: gl.COLOR_BUFFER_BIT)
type Constants struct {
	staticDraw         js.Value
	arrayBuffer        js.Value
	elementArrayBuffer js.Value
	vertexShader       js.Value
	fragmentShader     js.Value
	float              js.Value
	depthTest          js.Value
	colorBufferBit     js.Value
	triangles          js.Value
	unsignedShort      js.Value
}

// Context a handle to canvas webgl
type Context struct {
	DocumentEl js.Value
	CanvasEl   js.Value
	Ctx        js.Value
	Constants  Constants
}

// New initialize a new gl.Context
func New(canvasID string) (*Context, error) {
	gl := new(Context)

	// get document elements
	gl.DocumentEl = js.Global().Get("document")
	if gl.DocumentEl == js.Undefined() {
		return nil, fmt.Errorf("failed to load document element")
	}

	// get canvas element
	gl.CanvasEl = gl.DocumentEl.Call("getElementById", canvasID)
	if gl.CanvasEl == js.Undefined() {
		return gl, fmt.Errorf("invalid canvas id: %s", canvasID)
	}

	// get webgl context
	gl.Ctx = gl.CanvasEl.Call("getContext", "webgl")
	if gl.Ctx == js.Undefined() {
		return gl, fmt.Errorf("failed to load webgl context - may be unsupported by browser")
	}

	// initialize constants
	gl.Constants.staticDraw = gl.Ctx.Get("STATIC_DRAW")
	gl.Constants.arrayBuffer = gl.Ctx.Get("ARRAY_BUFFER")
	gl.Constants.elementArrayBuffer = gl.Ctx.Get("ELEMENT_ARRAY_BUFFER")
	gl.Constants.vertexShader = gl.Ctx.Get("VERTEX_SHADER")
	gl.Constants.fragmentShader = gl.Ctx.Get("FRAGMENT_SHADER")
	gl.Constants.float = gl.Ctx.Get("FLOAT")
	gl.Constants.depthTest = gl.Ctx.Get("DEPTH_TEST")
	gl.Constants.colorBufferBit = gl.Ctx.Get("COLOR_BUFFER_BIT")
	gl.Constants.triangles = gl.Ctx.Get("TRIANGLES")
	gl.Constants.unsignedShort = gl.Ctx.Get("UNSIGNED_SHORT")

	return gl, nil
}

// Test testing 1.. 2.. 3..
func (gl *Context) Test() {
	fmt.Println("gl.Conext.test() - testing... 1.. 2.. 3..")
}

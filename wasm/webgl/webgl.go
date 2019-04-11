package webgl

import (
	"fmt"
	"syscall/js"
)

type mesh struct {
	vertexBufferID   js.Value
	elementsBufferID js.Value
	vertexArrayID    js.Value
}

type shaderProgram struct {
	vertShaderID js.Value
	fragShaderID js.Value
	programID    js.Value
}

// Context a handle to canvas webgl
type Context struct {
	DocumentEl js.Value
	CanvasEl   js.Value
	ctx        js.Value

	programsByName map[string]shaderProgram
	meshesByName   map[string]mesh

	constants struct {
		vertexShader       js.Value
		fragmentShader     js.Value
		arrayBuffer        js.Value
		elementArrayBuffer js.Value
		staticDraw         js.Value
		colorBufferBit     js.Value
		depthBufferBit     js.Value
		depthTest          js.Value
		lEqual             js.Value
		float              js.Value
	}
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
	gl.ctx = gl.CanvasEl.Call("getContext", "webgl")
	if gl.ctx == js.Undefined() {
		return gl, fmt.Errorf("failed to load webgl context - may be unsupported by browser")
	}

	gl.programsByName = make(map[string]shaderProgram)
	gl.meshesByName = make(map[string]mesh)

	// initialize constants
	gl.constants.vertexShader = gl.ctx.Get("VERTEX_SHADER")
	gl.constants.fragmentShader = gl.ctx.Get("FRAGMENT_SHADER")
	gl.constants.arrayBuffer = gl.ctx.Get("ARRAY_BUFFER")
	gl.constants.elementArrayBuffer = gl.ctx.Get("ELEMENT_ARRAY_BUFFER")
	gl.constants.staticDraw = gl.ctx.Get("STATIC_DRAW")
	gl.constants.colorBufferBit = gl.ctx.Get("COLOR_BUFFER_BIT")
	gl.constants.depthBufferBit = gl.ctx.Get("DEPTH_BUFFER_BIT")
	gl.constants.depthTest = gl.ctx.Get("DEPTH_TEST")
	gl.constants.lEqual = gl.ctx.Get("LEQUAL")
	gl.constants.float = gl.ctx.Get("FLOAT")

	// do some initialization for stuff we know we'll need for the block game
	gl.ctx.Call("enable", gl.constants.depthTest)
	gl.ctx.Call("depthFunc", gl.constants.lEqual)

	return gl, nil
}

// NewShaderProgram links, compiles & registers a shader program using the given vertex & fragment shader
func (gl *Context) NewShaderProgram(name string, vertCode string, fragCode string) error {
	if _, programExists := gl.programsByName[name]; programExists {
		return fmt.Errorf("a program already exists with name: %s", name)
	}

	vertShaderID := gl.ctx.Call("createShader", gl.constants.vertexShader)
	gl.ctx.Call("shaderSource", vertShaderID, vertCode)
	gl.ctx.Call("compileShader", vertShaderID)

	fragShaderID := gl.ctx.Call("createShader", gl.constants.fragmentShader)
	gl.ctx.Call("shaderSource", fragShaderID, fragCode)
	gl.ctx.Call("compileShader", fragShaderID)

	programID := gl.ctx.Call("createProgram")
	gl.ctx.Call("attachShader", programID, vertShaderID)
	gl.ctx.Call("attachShader", programID, fragShaderID)
	gl.ctx.Call("linkProgram", programID)

	if gl.ctx.Call("getAttributeLocation", "position").Int() != 0 {
		return fmt.Errorf("all vertex shaders MUST have 'position' as it's first attribute")
	}

	program := shaderProgram{vertShaderID, fragShaderID, programID}
	gl.programsByName[name] = program

	return nil
}

// NewMesh creates a new mesh with the given name (meshes are simply combinations of verticies & elments)
func (gl *Context) NewMesh(name string, verticies []float32, elements []uint32) error {
	_, meshExists := gl.meshesByName[name]
	if meshExists {
		return fmt.Errorf("a mesh already exists with name '%s'", name)
	}

	// create vbo
	verticiesTyped := js.TypedArrayOf(verticies)
	vertBufferID := gl.ctx.Call("createBuffer", gl.constants.arrayBuffer)
	gl.ctx.Call("bindBuffer", gl.constants.arrayBuffer, verticiesTyped, gl.constants.staticDraw)

	// create ebo
	elementsTyped := js.TypedArrayOf(elements)
	elementBufferID := gl.ctx.Call("createBuffer", gl.constants.elementArrayBuffer)
	gl.ctx.Call("bindBuffer", gl.constants.elementArrayBuffer, elementsTyped, gl.constants.staticDraw)

	// create vao
	vertexArrayID := gl.ctx.Call("createVertexArray")
	gl.ctx.Call("bindVertexArray", vertexArrayID)
	gl.ctx.Call("vertexAttribPointer", 0, 3, gl.constants.float, false, 0, 0)
	gl.ctx.Call("enableVertexAttribArray", 0)

	gl.meshesByName[name] = mesh{vertBufferID, elementBufferID, vertexArrayID}

	// unbind everything
	gl.ctx.Call("bindBuffer", gl.constants.arrayBuffer, nil)
	gl.ctx.Call("bindVertexArray", 0)
	gl.ctx.Call("bindBuffer", gl.constants.elementArrayBuffer, nil)

	return nil
}

// UseProgram uses a shader program
func (gl *Context) UseProgram(name string) error {
	programID, programExists := gl.programsByName[name]
	if !programExists {
		return fmt.Errorf("no program found with name '%s'", name)
	}

	gl.ctx.Call("useProgram", programID)
	return nil
}

func (gl *Context) BindUniformMat4(name string, mat []float32) {

}

// ClearScreen clears the canvas to white
func (gl *Context) ClearScreen() {
	gl.ctx.Call("clearColor", 1.0, 1.0, 1.0, 1.0)
	gl.ctx.Call("clear", gl.constants.colorBufferBit)
	gl.ctx.Call("clear", gl.constants.depthBufferBit)
}

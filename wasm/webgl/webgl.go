package webgl

import (
	"fmt"
	// "github.com/go-gl/mathgl/mgl32"
	"syscall/js"
)

// Context a handle to canvas webgl
type Context struct {
	DocumentEl js.Value
	CanvasEl   js.Value
	ctx        js.Value

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

// ClearScreen clears the canvas to white
func (gl *Context) ClearScreen() {
	gl.ctx.Call("clearColor", 1.0, 1.0, 1.0, 1.0)
	gl.ctx.Call("clear", gl.constants.colorBufferBit)
	gl.ctx.Call("clear", gl.constants.depthBufferBit)
}

// Render renders the given mesh with the shader
func (gl *Context) Render(shader *ShaderProgram, mesh *Mesh) {

	return
}

// ShaderProgram a struct for managing a shader program
type ShaderProgram struct {
	gl           *Context
	vertShaderID js.Value
	fragShaderID js.Value
	programID    js.Value

	uniformsMat4f map[string]js.TypedArray
	uniformsVec4f map[string]js.TypedArray
}

// NewShaderProgram links, compiles & registers a shader program using the given vertex & fragment shader
func (gl *Context) NewShaderProgram(vertCode string, fragCode string) (*ShaderProgram, error) {
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

	if gl.ctx.Call("getAttribLocation", programID, "position").Int() != 0 {
		return nil, fmt.Errorf("all vertex shaders MUST have 'position' as it's first attribute")
	}

	program := new(ShaderProgram)
	program.gl = gl
	program.vertShaderID = vertShaderID
	program.fragShaderID = fragShaderID
	program.programID = programID
	program.uniformsMat4f = make(map[string]js.TypedArray)
	program.uniformsVec4f = make(map[string]js.TypedArray)

	return program, nil
}

// BindUniformMat4f binds a mat4f uniform to the shader program
func (program *ShaderProgram) BindUniformMat4f(name string, val js.TypedArray) error {
	if program.gl.ctx.Call("getUniformLocation", program.programID, name).Int() < 0 {
		return fmt.Errorf("No uniform exists with name '%s' for the given program", name)
	}

	program.uniformsMat4f[name] = val
	return nil
}

// BindUniformVec4f binds a vec4f uniform to the shader program
func (program *ShaderProgram) BindUniformVec4f(name string, val js.TypedArray) error {
	if program.gl.ctx.Call("getUniformLocation", program.programID, name).Int() < 0 {
		return fmt.Errorf("No uniform exists with name '%s' for the given program", name)
	}

	program.uniformsVec4f[name] = val
	return nil
}

// Mesh a struct for managing a mesh of vbo's, ebo's, & vao's
type Mesh struct {
	gl               *Context
	vertexBufferID   js.Value
	elementsBufferID js.Value
	vertexArrayID    js.Value
}

// NewMesh creates a new mesh with the given name (meshes are simply combinations of verticies & elments)
func (gl *Context) NewMesh(name string, verticies []float32, elements []uint32) *Mesh {
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

	// unbind everything
	gl.ctx.Call("bindBuffer", gl.constants.arrayBuffer, nil)
	gl.ctx.Call("bindVertexArray", 0)
	gl.ctx.Call("bindBuffer", gl.constants.elementArrayBuffer, nil)

	mesh := new(Mesh)
	mesh.vertexBufferID = vertBufferID
	mesh.elementsBufferID = elementBufferID
	mesh.vertexArrayID = vertexArrayID

	return mesh
}

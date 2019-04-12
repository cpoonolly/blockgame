package webgl

import (
	"fmt"
	"github.com/cpoonolly/blockgame/core"
	"syscall/js"
)

// Context a handle to canvas webgl
type Context struct {
	DocumentEl js.Value
	CanvasEl   js.Value
	ctx        js.Value
	width      int
	height     int

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
		linkStatus         js.Value
		compileStatus      js.Value
		float              js.Value
		unsignedShort      js.Value
		triangles          js.Value
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
	gl.constants.linkStatus = gl.ctx.Get("LINK_STATUS")
	gl.constants.compileStatus = gl.ctx.Get("COMPILE_STATUS")
	gl.constants.float = gl.ctx.Get("FLOAT")
	gl.constants.unsignedShort = gl.ctx.Get("UNSIGNED_SHORT")
	gl.constants.triangles = gl.ctx.Get("TRIANGLES")

	// do some initialization for stuff we know we'll need for the block game
	body := gl.DocumentEl.Get("body")
	gl.width = body.Get("clientWidth").Int()
	gl.height = body.Get("clientHeight").Int()

	gl.ctx.Call("viewport", 0, 0, gl.width, gl.height)
	gl.CanvasEl.Set("width", gl.width)
	gl.CanvasEl.Set("height", gl.height)

	return gl, nil
}

// GetViewportWidth gets the viewport width
func (gl *Context) GetViewportWidth() int {
	return gl.width
}

// GetViewportHeight gets the viewport height
func (gl *Context) GetViewportHeight() int {
	return gl.height
}

// ClearScreen clears the canvas to white
func (gl *Context) ClearScreen() error {
	gl.ctx.Call("clearColor", 0.0, 0.0, 0.0, 0.9)
	gl.ctx.Call("clearDepth", 1.0)
	gl.ctx.Call("enable", gl.constants.depthTest)
	gl.ctx.Call("depthFunc", gl.constants.lEqual)
	gl.ctx.Call("clear", gl.constants.colorBufferBit)
	gl.ctx.Call("clear", gl.constants.depthBufferBit)

	return nil
}

// Render renders the given mesh with the shader
func (gl *Context) Render(coreMesh core.Mesh, coreProgram core.ShaderProgram) error {
	mesh, isWebGlMesh := coreMesh.(*Mesh)
	if !isWebGlMesh {
		return fmt.Errorf("invalid mesh passed to this gl context. must be a webgl.Mesh")
	}

	program, isWebGlProgram := coreProgram.(*ShaderProgram)
	if !isWebGlProgram {
		return fmt.Errorf("invalid shader passed to this gl context. must be a webgl.ShaderProgram")
	}

	gl.ctx.Call("useProgram", program.programID)

	// bind all mat4f uniforms
	for uniformName, uniformVal := range program.uniformsMat4f {
		uniformLoc := gl.ctx.Call("getUniformLocation", program.programID, uniformName)
		gl.ctx.Call("uniformMatrix4fv", uniformLoc, false, uniformVal)
	}

	// bind all vec4f uniforms
	for uniformName, uniformVal := range program.uniformsVec4f {
		uniformLoc := gl.ctx.Call("getUniformLocation", program.programID, uniformName)
		gl.ctx.Call("uniform4fv", uniformLoc, uniformVal)
	}

	// bind elements
	gl.ctx.Call("bindBuffer", gl.constants.elementArrayBuffer, mesh.elementsBufferID)

	// bind position attribute
	gl.ctx.Call("bindBuffer", gl.constants.arrayBuffer, mesh.vertexBufferID)
	gl.ctx.Call("vertexAttribPointer", 0, 3, gl.constants.float, false, 0, 0)
	gl.ctx.Call("enableVertexAttribArray", 0)

	// bind normal attribute
	gl.ctx.Call("bindBuffer", gl.constants.arrayBuffer, mesh.normalBufferID)
	gl.ctx.Call("vertexAttribPointer", 1, 3, gl.constants.float, false, 0, 0)
	gl.ctx.Call("enableVertexAttribArray", 1)

	gl.ctx.Call("drawElements", gl.constants.triangles, mesh.size, gl.constants.unsignedShort, 0)

	return nil
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
func (gl *Context) NewShaderProgram(
	vertCode string,
	fragCode string,
	uniformsMat4f map[string][]float32,
	uniformsVec4f map[string][]float32,
) (core.ShaderProgram, error) {

	// TODO should defer gl.ctx.Call("deleteShader", vertShaderID) on failure
	// TODO should defer gl.ctx.Call("deleteShader", fragShaderID) on failure

	vertShaderID := gl.ctx.Call("createShader", gl.constants.vertexShader)
	gl.ctx.Call("shaderSource", vertShaderID, vertCode)
	gl.ctx.Call("compileShader", vertShaderID)

	if compileStatusOk := gl.ctx.Call("getShaderParameter", vertShaderID, gl.constants.compileStatus).Truthy(); !compileStatusOk {
		return nil, fmt.Errorf("failed to compile vert shader: %s", gl.ctx.Call("getShaderInfoLog", vertShaderID).String())
	}

	fragShaderID := gl.ctx.Call("createShader", gl.constants.fragmentShader)
	gl.ctx.Call("shaderSource", fragShaderID, fragCode)
	gl.ctx.Call("compileShader", fragShaderID)

	if compileStatusOk := gl.ctx.Call("getShaderParameter", fragShaderID, gl.constants.compileStatus).Truthy(); !compileStatusOk {
		return nil, fmt.Errorf("failed to compile frag shader: %s", gl.ctx.Call("getShaderInfoLog", fragShaderID).String())
	}

	programID := gl.ctx.Call("createProgram")
	gl.ctx.Call("attachShader", programID, vertShaderID)
	gl.ctx.Call("attachShader", programID, fragShaderID)
	gl.ctx.Call("linkProgram", programID)

	if linkStatusOk := gl.ctx.Call("getProgramParameter", programID, gl.constants.linkStatus).Truthy(); !linkStatusOk {
		js.Global().Call("showProgramLinkError", programID)
		return nil, fmt.Errorf("failed to generate shader progam: %s", gl.ctx.Call("getProgramInfoLog", programID).String())
	}

	if gl.ctx.Call("getAttribLocation", programID, "position").Int() != 0 {
		return nil, fmt.Errorf("all vertex shaders MUST have 'position' as it's first attribute")
	}

	if gl.ctx.Call("getAttribLocation", programID, "normal").Int() != 1 {
		return nil, fmt.Errorf("all vertex shaders MUST have 'normal' as it's second attribute")
	}

	program := new(ShaderProgram)
	program.gl = gl
	program.vertShaderID = vertShaderID
	program.fragShaderID = fragShaderID
	program.programID = programID

	program.uniformsMat4f = make(map[string]js.TypedArray)
	program.uniformsVec4f = make(map[string]js.TypedArray)

	for uniformName, uniformVal := range uniformsMat4f {
		if !gl.ctx.Call("getUniformLocation", programID, uniformName).Truthy() {
			return nil, fmt.Errorf("invalid uniform '%s' passed to shader", uniformName)
		}

		program.uniformsMat4f[uniformName] = js.TypedArrayOf(uniformVal)
	}

	for uniformName, uniformVal := range uniformsVec4f {
		if !gl.ctx.Call("getUniformLocation", programID, uniformName).Truthy() {
			return nil, fmt.Errorf("invalid uniform '%s' passed to shader", uniformName)
		}

		program.uniformsVec4f[uniformName] = js.TypedArrayOf(uniformVal)
	}

	return program, nil
}

// Mesh a struct for managing a mesh of vbo's, ebo's, & vao's
type Mesh struct {
	gl               *Context
	vertexBufferID   js.Value
	normalBufferID   js.Value
	elementsBufferID js.Value
	verticies        js.TypedArray
	normals          js.TypedArray
	elements         js.TypedArray
	size             int
}

// NewMesh creates a new mesh (meshes are simply combinations of verticies & elments)
func (gl *Context) NewMesh(verticies []float32, normals []float32, elements []uint16) (core.Mesh, error) {
	verticiesTyped := js.TypedArrayOf(verticies)
	vertBufferID := gl.ctx.Call("createBuffer", gl.constants.arrayBuffer)
	gl.ctx.Call("bindBuffer", gl.constants.arrayBuffer, vertBufferID)
	gl.ctx.Call("bufferData", gl.constants.arrayBuffer, verticiesTyped, gl.constants.staticDraw)

	normalsTyped := js.TypedArrayOf(verticies)
	normBufferID := gl.ctx.Call("createBuffer", gl.constants.arrayBuffer)
	gl.ctx.Call("bindBuffer", gl.constants.arrayBuffer, normBufferID)
	gl.ctx.Call("bufferData", gl.constants.arrayBuffer, normalsTyped, gl.constants.staticDraw)

	elementsTyped := js.TypedArrayOf(elements)
	elementBufferID := gl.ctx.Call("createBuffer", gl.constants.elementArrayBuffer)
	gl.ctx.Call("bindBuffer", gl.constants.elementArrayBuffer, elementBufferID)
	gl.ctx.Call("bufferData", gl.constants.elementArrayBuffer, elementsTyped, gl.constants.staticDraw)

	// unbind everything
	gl.ctx.Call("bindBuffer", gl.constants.arrayBuffer, nil)
	gl.ctx.Call("bindBuffer", gl.constants.elementArrayBuffer, nil)

	mesh := new(Mesh)
	mesh.vertexBufferID = vertBufferID
	mesh.normalBufferID = normBufferID
	mesh.elementsBufferID = elementBufferID
	mesh.verticies = verticiesTyped
	mesh.normals = normalsTyped
	mesh.elements = elementsTyped
	mesh.size = len(elements)

	return mesh, nil
}

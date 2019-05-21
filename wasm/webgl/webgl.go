package webgl

import (
	"fmt"
	"syscall/js"

	"github.com/cpoonolly/blockgame/core"
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
		lines              js.Value
		cullFace           js.Value
	}
}

// New initialize a new webgl.Context
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
	gl.constants.lines = gl.ctx.Get("GL_LINES")

	// calculate Viewport
	gl.UpdateViewport()

	return gl, nil
}

// UpdateViewport recalculates the viewport given the canvas' width/height
func (gl *Context) UpdateViewport() {
	gl.width = gl.CanvasEl.Get("clientWidth").Int()
	gl.height = gl.CanvasEl.Get("clientHeight").Int()

	gl.ctx.Call("viewport", 0, 0, gl.width, gl.height)
	gl.CanvasEl.Set("width", gl.width)
	gl.CanvasEl.Set("height", gl.height)
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
func (gl *Context) ClearScreen(colorR, colorG, colorB float32) error {
	gl.ctx.Call("clearColor", colorR, colorG, colorB, 0.9)
	gl.ctx.Call("clearDepth", 1.0)
	gl.ctx.Call("enable", gl.constants.depthTest)
	gl.ctx.Call("depthFunc", gl.constants.lEqual)
	gl.ctx.Call("clear", gl.constants.colorBufferBit)
	gl.ctx.Call("clear", gl.constants.depthBufferBit)

	return nil
}

// Enable simple interface to gl enable
func (gl *Context) Enable(constName string) {
	glConst := gl.ctx.Get(constName)
	gl.ctx.Call("enable", glConst)
}

// Disable simple interface to gl disable
func (gl *Context) Disable(constName string) {
	glConst := gl.ctx.Get(constName)
	gl.ctx.Call("enable", glConst)
}

// RenderTriangles renders the triangles of the given mesh with the shader
func (gl *Context) RenderTriangles(coreMesh core.Mesh, coreProgram core.ShaderProgram) error {
	return gl.render(coreMesh, coreProgram, gl.constants.triangles)
}

// RenderLines renders the lines of the given mesh with the shader
func (gl *Context) RenderLines(coreMesh core.Mesh, coreProgram core.ShaderProgram) error {
	return gl.render(coreMesh, coreProgram, gl.constants.lines)
}

// Render renders the given mesh with the shader
func (gl *Context) render(coreMesh core.Mesh, coreProgram core.ShaderProgram, renderConst js.Value) error {
	mesh, isWebGlMesh := coreMesh.(*Mesh)
	if !isWebGlMesh {
		return fmt.Errorf("invalid mesh passed to this gl context. must be a webgl.Mesh")
	}

	program, isWebGlProgram := coreProgram.(*ShaderProgram)
	if !isWebGlProgram {
		return fmt.Errorf("invalid shader passed to this gl context. must be a webgl.ShaderProgram")
	}

	gl.Enable("GL_CULL_FACE")
	gl.ctx.Call("useProgram", program.programID)

	// bind all uniforms
	for uniformName, uniformVal := range program.uniforms {
		uniformLoc := gl.ctx.Call("getUniformLocation", program.programID, uniformName)

		switch uniformSize := uniformVal.Length(); uniformSize {
		case 16:
			gl.ctx.Call("uniformMatrix4fv", uniformLoc, false, uniformVal)
		case 4:
			gl.ctx.Call("uniform4fv", uniformLoc, uniformVal)
		case 3:
			gl.ctx.Call("uniform3fv", uniformLoc, uniformVal)
		default:
			return fmt.Errorf("Unsupported uniform: %s", uniformName)
		}
	}

	// bind elements
	gl.ctx.Call("bindBuffer", gl.constants.elementArrayBuffer, mesh.elementsBufferID)

	// bind position attribute
	posAttrLoc := gl.ctx.Call("getAttribLocation", program.programID, "aPosition").Int()
	gl.ctx.Call("bindBuffer", gl.constants.arrayBuffer, mesh.vertexBufferID)
	gl.ctx.Call("vertexAttribPointer", posAttrLoc, 3, gl.constants.float, false, 0, 0)
	gl.ctx.Call("enableVertexAttribArray", 0)

	// bind normal attribute
	normAttrLoc := gl.ctx.Call("getAttribLocation", program.programID, "aNormal").Int()
	gl.ctx.Call("bindBuffer", gl.constants.arrayBuffer, mesh.normalBufferID)
	gl.ctx.Call("vertexAttribPointer", normAttrLoc, 3, gl.constants.float, false, 0, 0)
	gl.ctx.Call("enableVertexAttribArray", 1)

	gl.ctx.Call("drawElements", renderConst, mesh.size, gl.constants.unsignedShort, 0)

	return nil
}

// ShaderProgram a struct for managing a shader program
type ShaderProgram struct {
	gl           *Context
	vertShaderID js.Value
	fragShaderID js.Value
	programID    js.Value

	uniforms map[string]js.TypedArray
}

// NewShaderProgram links, compiles & registers a shader program using the given vertex & fragment shader
func (gl *Context) NewShaderProgram(
	vertCode string,
	fragCode string,
	uniforms map[string][]float32,
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

	if gl.ctx.Call("getAttribLocation", programID, "aPosition").Int() < 0 {
		return nil, fmt.Errorf("all vertex shaders MUST have 'aPosition' as an attribute")
	}

	if gl.ctx.Call("getAttribLocation", programID, "aNormal").Int() < 0 {
		return nil, fmt.Errorf("all vertex shaders MUST have 'aNormal' as an attribute")
	}

	program := new(ShaderProgram)
	program.gl = gl
	program.vertShaderID = vertShaderID
	program.fragShaderID = fragShaderID
	program.programID = programID

	program.uniforms = make(map[string]js.TypedArray)

	for uniformName, uniformVal := range uniforms {
		if !gl.ctx.Call("getUniformLocation", programID, uniformName).Truthy() {
			return nil, fmt.Errorf("invalid uniform '%s' passed to shader", uniformName)
		}

		program.uniforms[uniformName] = js.TypedArrayOf(uniformVal)
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

	normalsTyped := js.TypedArrayOf(normals)
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

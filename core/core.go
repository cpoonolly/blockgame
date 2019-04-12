package core

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
)

// ShaderProgram generic interface for shader returned by GlContext below
type ShaderProgram interface{}

// Mesh generic interface for mesh returned by GlContext below
type Mesh interface{}

// GlContext represents a generic gl context (not necessarily WebGL) that can be used by the game
type GlContext interface {
	GetViewportWidth() int
	GetViewportHeight() int
	ClearScreen() error
	NewShaderProgram(string, string, map[string][]float32, map[string][]float32) (ShaderProgram, error)
	NewMesh([]float32, []float32, []uint16) (Mesh, error)
	Render(Mesh, ShaderProgram) error
}

// Game represents a game
type Game struct {
	gl          GlContext
	blockShader ShaderProgram
	blockMesh   Mesh

	projMatrix      mgl32.Mat4
	modelViewMatrix mgl32.Mat4
	color           mgl32.Vec4
}

// NewGame creates a new Game instance
func NewGame(glCtx GlContext) (*Game, error) {
	game := new(Game)
	game.gl = glCtx

	fmt.Println("Created a new Game!")
	var err error

	viewportWidth := float32(game.gl.GetViewportWidth())
	viewportHeight := float32(game.gl.GetViewportHeight())
	aspectRatio := viewportWidth / viewportHeight

	game.projMatrix = mgl32.Perspective(mgl32.DegToRad(45.0), aspectRatio, 1, 50.0)
	game.modelViewMatrix = mgl32.Ident4()
	game.color = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}

	uniformsMat4f := map[string][]float32{
		"pMatrix":  game.projMatrix[:],
		"mvMatrix": game.modelViewMatrix[:],
	}

	uniformsVec4f := map[string][]float32{
		"color": game.color[:],
	}

	game.blockShader, err = game.gl.NewShaderProgram(blockVertShaderCode, blockFragShaderCode, uniformsMat4f, uniformsVec4f)
	if err != nil {
		return nil, err
	}

	game.blockMesh, err = game.gl.NewMesh(blockVerticies[:], blockNormals[:], blockIndicies[:])
	if err != nil {
		return nil, err
	}

	return game, nil
}

// Update updates the game models
func (game *Game) Update(cameraX float32, cameraY float32, cameraZ float32) {
	cameraPos := mgl32.Vec3{0.0 + cameraX, 0.0 + cameraY, -6.0 + cameraZ}
	cameraLookAt := mgl32.Vec3{0.0, 0.0, 0.0}
	cameraUp := mgl32.Vec3{0.0, 1.0, 0.0}

	viewMatrix := mgl32.LookAtV(cameraPos, cameraLookAt, cameraUp)
	modelMatrix := mgl32.Ident4()
	game.modelViewMatrix = viewMatrix.Mul4(modelMatrix)
}

// Render renders the frame
func (game *Game) Render() {
	if err := game.gl.ClearScreen(); err != nil {
		panic(err)
	}

	if err := game.gl.Render(game.blockMesh, game.blockShader); err != nil {
		panic(err)
	}
}

var blockVerticies = [...]float32{
	// Front face
	-1.0, -1.0, 1.0,
	1.0, -1.0, 1.0,
	1.0, 1.0, 1.0,
	-1.0, 1.0, 1.0,

	// Back face
	-1.0, -1.0, -1.0,
	-1.0, 1.0, -1.0,
	1.0, 1.0, -1.0,
	1.0, -1.0, -1.0,

	// Top face
	-1.0, 1.0, -1.0,
	-1.0, 1.0, 1.0,
	1.0, 1.0, 1.0,
	1.0, 1.0, -1.0,

	// Bottom face
	-1.0, -1.0, -1.0,
	1.0, -1.0, -1.0,
	1.0, -1.0, 1.0,
	-1.0, -1.0, 1.0,

	// Right face
	1.0, -1.0, -1.0,
	1.0, 1.0, -1.0,
	1.0, 1.0, 1.0,
	1.0, -1.0, 1.0,

	// Left face
	-1.0, -1.0, -1.0,
	-1.0, -1.0, 1.0,
	-1.0, 1.0, 1.0,
	-1.0, 1.0, -1.0,
}

var blockNormals = [...]float32{
	// Front
	0.0, 0.0, 1.0,
	0.0, 0.0, 1.0,
	0.0, 0.0, 1.0,
	0.0, 0.0, 1.0,

	// Back
	0.0, 0.0, -1.0,
	0.0, 0.0, -1.0,
	0.0, 0.0, -1.0,
	0.0, 0.0, -1.0,

	// Top
	0.0, 1.0, 0.0,
	0.0, 1.0, 0.0,
	0.0, 1.0, 0.0,
	0.0, 1.0, 0.0,

	// Bottom
	0.0, -1.0, 0.0,
	0.0, -1.0, 0.0,
	0.0, -1.0, 0.0,
	0.0, -1.0, 0.0,

	// Right
	1.0, 0.0, 0.0,
	1.0, 0.0, 0.0,
	1.0, 0.0, 0.0,
	1.0, 0.0, 0.0,

	// Left
	-1.0, 0.0, 0.0,
	-1.0, 0.0, 0.0,
	-1.0, 0.0, 0.0,
	-1.0, 0.0, 0.0,
}

var blockIndicies = [...]uint16{
	// front
	0, 1, 2,
	0, 2, 3,
	// back
	4, 5, 6,
	4, 6, 7,
	// top
	8, 9, 10,
	8, 10, 11,
	// bottom
	12, 13, 14,
	12, 14, 15,
	// right
	16, 17, 18,
	16, 18, 19,
	// left
	20, 21, 22,
	20, 22, 23,
}

var blockVertShaderCode = `
	attribute vec3 position;
	attribute vec3 normal;

	uniform mat4 pMatrix;
	uniform mat4 mvMatrix;
	uniform vec4 color;

	varying highp vec3 vLighting;
	varying highp vec3 vColor;

	void main(void) {
		gl_Position = pMatrix * mvMatrix * vec4(position, 1.);

		highp vec3 ambientLight = vec3(0.3, 0.3, 0.3);
		highp vec3 directionalLightColor = vec3(.5, .5, .5);
		highp vec3 directionalVector = normalize(vec3(0.85, 0.8, 0.75));
		highp float directional = max(dot(normal, directionalVector), 0.0);
		vLighting = ambientLight + (directionalLightColor * directional);

		vColor = color.rgb;
	}	
`

var blockFragShaderCode = `
	varying highp vec3 vLighting;
	varying highp vec3 vColor;

	void main(void) {
		gl_FragColor = vec4(vColor * vLighting, 1.);
	}
`

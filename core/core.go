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
	Enable(string)
	Disable(string)
	ClearScreen(float32, float32, float32) error
	NewShaderProgram(string, string, map[string][]float32, map[string][]float32) (ShaderProgram, error)
	NewMesh([]float32, []float32, []uint16) (Mesh, error)
	RenderTriangles(Mesh, ShaderProgram) error
	RenderLines(Mesh, ShaderProgram) error
}

type block struct {
	pos   mgl32.Vec3
	scale mgl32.Vec3
	color mgl32.Vec4
}

type camera struct {
	pos    mgl32.Vec3
	lookAt mgl32.Vec3
	up     mgl32.Vec3
}

// Game represents a game
type Game struct {
	gl          GlContext
	blockShader ShaderProgram
	blockMesh   Mesh

	projMatrix      mgl32.Mat4
	modelViewMatrix mgl32.Mat4
	normalMatrix    mgl32.Mat4
	color           mgl32.Vec4

	blocks []block
	camera camera

	Log string
}

// NewGame creates a new Game instance
func NewGame(glCtx GlContext) (*Game, error) {
	game := new(Game)
	game.gl = glCtx

	var err error

	viewportWidth := float32(game.gl.GetViewportWidth())
	viewportHeight := float32(game.gl.GetViewportHeight())
	aspectRatio := viewportWidth / viewportHeight

	game.projMatrix = mgl32.Perspective(mgl32.DegToRad(45.0), aspectRatio, 1, 50.0)
	game.modelViewMatrix = mgl32.Ident4()
	game.normalMatrix = mgl32.Ident4()
	game.color = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}

	game.blockShader, err = game.gl.NewShaderProgram(
		blockVertShaderCode,
		blockFragShaderCode,
		map[string][]float32{
			"pMatrix":    game.projMatrix[:],
			"mvMatrix":   game.modelViewMatrix[:],
			"normMatrix": game.normalMatrix[:],
		},
		map[string][]float32{"color": game.color[:]},
	)

	if err != nil {
		return nil, err
	}

	game.blockMesh, err = game.gl.NewMesh(blockVerticies[:], blockNormals[:], blockIndicies[:])
	if err != nil {
		return nil, err
	}

	game.blocks = make([]block, 3)

	// block 1
	game.blocks[0].pos = mgl32.Vec3{3.0, 0.0, 0.0}
	game.blocks[0].scale = mgl32.Vec3{1.0, 1.0, 1.0}
	game.blocks[0].color = mgl32.Vec4{0.1, 1.0, 0.1, 1.0}

	// block 2
	game.blocks[1].pos = mgl32.Vec3{-3.0, 0.0, 0.0}
	game.blocks[1].scale = mgl32.Vec3{1.0, 1.0, 1.0}
	game.blocks[1].color = mgl32.Vec4{1.0, 0.1, 0.1, 1.0}

	// block 3
	game.blocks[2].pos = mgl32.Vec3{0.0, 0.0, -3.0}
	game.blocks[2].scale = mgl32.Vec3{1.0, 1.0, 1.0}
	game.blocks[2].color = mgl32.Vec4{0.1, 0.1, 1.0, 1.0}

	game.camera.pos = mgl32.Vec3{0.0, 0.0, -6.0}
	game.camera.lookAt = mgl32.Vec3{0.0, 0.0, 0.0}
	game.camera.up = mgl32.Vec3{0.0, 1.0, 0.0}

	return game, nil
}

// Update updates the game models
func (game *Game) Update(dt, dx, dy, dz float32) {
	game.camera.pos = game.camera.pos.Add(mgl32.Vec3{dx, dy, dz})

	game.Log = fmt.Sprintf(
		"FPS: %f\tCamera: (x: %f, y: %f, z: %f)",
		1000.0/dt,
		game.camera.pos.X(),
		game.camera.pos.Y(),
		game.camera.pos.Z(),
	)
}

// Render renders the frame
func (game *Game) Render() {
	if err := game.gl.ClearScreen(0.0, 0.0, 0.0); err != nil {
		panic(err)
	}

	camera := game.camera
	viewMatrix := mgl32.LookAtV(camera.pos, camera.lookAt, camera.up)

	for _, block := range game.blocks {
		if err := game.renderBlock(block, viewMatrix); err != nil {
			panic(err)
		}
	}
}

func (game *Game) renderBlock(block block, viewMatrix mgl32.Mat4) error {
	scaleMatrix := mgl32.Scale3D(block.scale.X(), block.scale.Y(), block.scale.Z())
	translateMatrix := mgl32.Translate3D(block.pos.X(), block.pos.Y(), block.pos.Z())

	modelMatrix := mgl32.Ident4().Mul4(scaleMatrix).Mul4(translateMatrix)

	game.modelViewMatrix = viewMatrix.Mul4(modelMatrix)
	game.normalMatrix = game.modelViewMatrix.Inv().Transpose()
	game.color = block.color

	if err := game.gl.RenderTriangles(game.blockMesh, game.blockShader); err != nil {
		return err
	}

	return nil
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

// TODO - Look into cell shading from this guy: https://github.com/aakshayy/toonshader-webgl
var blockVertShaderCode = `
	attribute vec3 position;
	attribute vec3 normal;

	uniform mat4 pMatrix;
	uniform mat4 mvMatrix;
	uniform mat4 normMatrix;
	uniform vec4 color;

	varying highp vec3 vColor;

	void main(void) {
		vec4 vertPos = mvMatrix * vec4(position, 1.);
		gl_Position = pMatrix * vertPos;

		vec3 ambient = 0.4 * color.rgb;

		vec3 lightPos = vec3(1.0, 1.0, 1.0);
		vec3 transformedLight = normalize(lightPos - vertPos.xyz);
		vec3 transformedNorm = normalize(vec3(normMatrix * vec4(normal, 0.0)));
		float lambert = max(dot(transformedNorm, transformedLight), 0.0);
		vec3 diffuse = lambert * 0.7 * color.rgb;

		vColor = ambient + diffuse;
	}	
`

var blockFragShaderCode = `
	varying highp vec3 vColor;

	void main(void) {
		gl_FragColor = vec4(vColor, 1.);
	}
`

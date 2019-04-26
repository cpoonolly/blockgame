package core

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
)

// GameInput an input for the game
type GameInput int

const (
	// GameInputPlayerMoveForward input to move the player forward
	GameInputPlayerMoveForward GameInput = iota + 1
	// GameInputPlayerMoveBack input to move the player back
	GameInputPlayerMoveBack
	// GameInputPlayerMoveLeft input to move the player left
	GameInputPlayerMoveLeft
	// GameInputPlayerMoveRight input to move the player right
	GameInputPlayerMoveRight
	// GameInputCameraZoomIn input to zoom camera in
	GameInputCameraZoomIn
	// GameInputCameraZoomOut input to zoom camera out
	GameInputCameraZoomOut
	// GameInputCameraRotateLeft input to rotate camera left relative to it's lookAt point
	GameInputCameraRotateLeft
	// GameInputCameraRotateRight input to rotate camera right relative to it's lookAt point
	GameInputCameraRotateRight
	// GameInputEditModeToggle toggles edit mode
	GameInputEditModeToggle
	// GameInputEditModeMoveUp input to move the player up
	GameInputEditModeMoveUp
	// GameInputEditModeMoveDown input to move the player down
	GameInputEditModeMoveDown
)

type gameUpdatable interface {
	update(game *Game, dt float32, inputs map[GameInput]bool)
}

type gameRenderable interface {
	render(game *Game, viewMatrix mgl32.Mat4) error
}

// maximum velocity for a moving object
const maxVelocity float32 = 10
const dampening float32 = 1

const gravityAcceleration float32 = 1

// Game represents a game
type Game struct {
	gl          GlContext
	blockShader ShaderProgram
	blockMesh   Mesh

	projMatrix      mgl32.Mat4
	modelViewMatrix mgl32.Mat4
	normalMatrix    mgl32.Mat4
	color           mgl32.Vec4

	player      *player
	enemies     map[uint32]*enemy
	worldBlocks map[uint32]*worldBlock
	camera      camera

	IsEditModeEnabled bool
	lastBlockID       uint32

	Log string
}

// NewGame creates a new Game instance
func NewGame(glCtx GlContext) (*Game, error) {
	game := new(Game)
	game.gl = glCtx

	var err error

	game.OnViewPortChange()
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

	game.player = new(player)
	game.player.scale = mgl32.Vec3{0.5, 0.5, 0.5}
	game.player.color = mgl32.Vec4{.3, .5, 1, 1.0}

	// generate world blocks
	game.worldBlocks = make(map[uint32]*worldBlock)

	// generate enemies
	game.enemies = make(map[uint32]*enemy)

	// create a camera
	arcballCamera := new(arcballCamera)
	arcballCamera.up = mgl32.Vec3{0.0, 1.0, 0.0}
	game.camera = arcballCamera

	game.IsEditModeEnabled = true

	return game, nil
}

// Update updates the game models
func (game *Game) Update(dt float32, inputs map[GameInput]bool) {
	if inputs[GameInputEditModeToggle] {
		game.IsEditModeEnabled = !game.IsEditModeEnabled
	}

	if game.IsEditModeEnabled {
		player := game.player
		camera := game.camera.(*arcballCamera)
		eyePos := camera.getEyePos()
		game.Log = fmt.Sprintf(
			"FPS: %.2f\tCamera: (x:%.2f, y:%.2f, z:%.2f)\tPlayer: (x:%.2f, y:%.2f, z:%.2f)",
			1000.0/dt,
			eyePos.X(),
			eyePos.Y(),
			eyePos.Z(),
			player.pos.X(),
			player.pos.Y(),
			player.pos.Z(),
		)
	} else {
		game.Log = ""
	}

	game.player.update(game, dt, inputs)
	game.camera.update(game, dt, inputs)

	for _, enemy := range game.enemies {
		enemy.update(game, dt, inputs)
	}

	if !game.IsEditModeEnabled && game.player.pos.Y() < -10.0 {
		game.GameOver()
	}
}

// Render renders the frame
func (game *Game) Render() {
	color := mgl32.Vec3{0.0, 0.0, 0.0}
	if game.IsEditModeEnabled {
		color = mgl32.Vec3{0.9, 0.9, 0.9}
	}

	if err := game.gl.ClearScreen(color.X(), color.Y(), color.Z()); err != nil {
		panic(err)
	}

	viewMatrix := game.camera.getViewMatrix()

	// Render player
	if err := game.player.render(game, viewMatrix); err != nil {
		panic(err)
	}

	// Render enemies
	for _, enemy := range game.enemies {
		if err := enemy.render(game, viewMatrix); err != nil {
			panic(err)
		}
	}

	// Render world
	for _, block := range game.worldBlocks {
		if err := block.render(game, viewMatrix); err != nil {
			panic(err)
		}
	}
}

// CreateWorldBlock editor function to create a new world block.
func (game *Game) CreateWorldBlock(pos, dimensions, color [3]float32) uint32 {
	newBlock := new(worldBlock)
	newBlock.id = game.lastBlockID + 1 // NOTE: not thread safe...
	game.lastBlockID = newBlock.id
	game.worldBlocks[newBlock.id] = newBlock
	game.UpdateWorldBlock(newBlock.id, pos, dimensions, color)

	return newBlock.id
}

// UpdateWorldBlock editor function to update a existing world block
func (game *Game) UpdateWorldBlock(id uint32, position, dimensions, color [3]float32) {
	block := game.worldBlocks[id]
	block.color = (mgl32.Vec3{color[0], color[1], color[2]}).Vec4(1.0)
	block.scale = (mgl32.Vec3{dimensions[0], dimensions[1], dimensions[2]}).Mul(0.5)
	block.pos = (mgl32.Vec3{position[0], position[1], position[2]}).Add(block.scale)
}

// DeleteWorldBlock editor function to delete a block
func (game *Game) DeleteWorldBlock(id uint32) {
	delete(game.worldBlocks, id)
}

// GetWorldBlockPosition get's the x,y,z coordinates (in that order) of right, bottom, back vertex of the block
func (game *Game) GetWorldBlockPosition(id uint32) [3]float32 {
	block := game.worldBlocks[id]
	return block.pos.Add(block.scale.Mul(-1))
}

// GetWorldBlockDimensions get's the width, height, & length (in that order) of the block
func (game *Game) GetWorldBlockDimensions(id uint32) [3]float32 {
	return game.worldBlocks[id].scale.Mul(2)
}

// GetWorldBlockColor get's the rgb color (in that order) of the block
func (game *Game) GetWorldBlockColor(id uint32) [3]float32 {
	return game.worldBlocks[id].color.Vec3()
}

// CreateEnemy editor function to create a new world block.
func (game *Game) CreateEnemy(pos, dimensions, color [3]float32) uint32 {
	newBlock := new(enemy)
	newBlock.id = game.lastBlockID + 1 // NOTE: not thread safe...
	game.lastBlockID = newBlock.id
	game.enemies[newBlock.id] = newBlock
	game.UpdateEnemy(newBlock.id, pos, dimensions, color)

	return newBlock.id
}

// UpdateEnemy editor function to update a existing world block
func (game *Game) UpdateEnemy(id uint32, position, dimensions, color [3]float32) {
	block := game.enemies[id]
	block.color = (mgl32.Vec3{color[0], color[1], color[2]}).Vec4(1.0)
	block.scale = (mgl32.Vec3{dimensions[0], dimensions[1], dimensions[2]}).Mul(0.5)
	block.pos = (mgl32.Vec3{position[0], position[1], position[2]}).Add(block.scale)
	block.start = block.pos
}

// DeleteEnemy editor function to delete a block
func (game *Game) DeleteEnemy(id uint32) {
	delete(game.enemies, id)
}

// GetEnemyPosition get's the x,y,z coordinates (in that order) of right, bottom, back vertex of the block
func (game *Game) GetEnemyPosition(id uint32) [3]float32 {
	block := game.enemies[id]
	return block.pos.Add(block.scale.Mul(-1))
}

// GetEnemyDimensions get's the width, height, & length (in that order) of the block
func (game *Game) GetEnemyDimensions(id uint32) [3]float32 {
	return game.enemies[id].scale.Mul(2)
}

// GetEnemyColor get's the rgb color (in that order) of the block
func (game *Game) GetEnemyColor(id uint32) [3]float32 {
	return game.enemies[id].color.Vec3()
}

// OnViewPortChange recalculates the projection matrix after a viewport adjustment
func (game *Game) OnViewPortChange() {
	game.gl.UpdateViewport()

	viewportWidth := float32(game.gl.GetViewportWidth())
	viewportHeight := float32(game.gl.GetViewportHeight())
	aspectRatio := viewportWidth / viewportHeight

	game.projMatrix = mgl32.Perspective(mgl32.DegToRad(45.0), aspectRatio, 1, 50.0)
}

// GameOver the game is over
func (game *Game) GameOver() {
	// GAME OVER
	game.player.pos = mgl32.Vec3{0, 0, 0}

	for _, enemy := range game.enemies {
		enemy.pos = enemy.start
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

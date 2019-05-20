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
	// GameInputPlayerJump input to have the player jump
	GameInputPlayerJump
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
	// GameInputEditModeCreateWorldBlock input to create a world block in edit mode
	GameInputEditModeCreateWorldBlock
	// GameInputEditModeCreateEnemy input to create a enemy in edit mode
	GameInputEditModeCreateEnemy
	// GameInputEditModeDelete input to delete all colliding blocks in edit mode
	GameInputEditModeDelete
)

type gameUpdatable interface {
	update(game *Game, dt float32, inputs map[GameInput]bool)
}

type gameRenderable interface {
	render(game *Game, viewMatrix mgl32.Mat4) error
}

// maximum velocity for a moving object
const maxVelocity float32 = 10
const terminalVelocity float32 = 20
const dampening float32 = 1

const gravityAcceleration float32 = 1

// Game represents a game
type Game struct {
	gl            GlContext
	phongShader   ShaderProgram
	gouraudShader ShaderProgram
	blockMesh     Mesh

	projMatrix      mgl32.Mat4
	modelViewMatrix mgl32.Mat4
	normalMatrix    mgl32.Mat4

	color    mgl32.Vec4
	material mgl32.Vec4 // vector of [Ka (ambient constant), Kd (diffuse constant), Ks (specular constant), shininess (shininess constant)] for the material
	lightPos mgl32.Vec3

	player      *player
	enemies     []*enemy
	worldBlocks []*worldBlock
	camera      camera
	editor      *gameEditor

	IsEditModeEnabled bool
	IsGameOver        bool

	Log string
}

// NewGame creates a new Game instance
func NewGame(glCtx GlContext) (*Game, error) {
	game := new(Game)
	game.gl = glCtx

	var err error

	game.OnViewPortChange()

	game.player = new(player)
	game.player.scale = mgl32.Vec3{0.5, 0.5, 0.5}

	// generate world blocks
	game.worldBlocks = make([]*worldBlock, 0, 100)

	// generate enemies
	game.enemies = make([]*enemy, 0, 20)

	// create a camera
	arcballCamera := new(arcballCamera)
	arcballCamera.up = mgl32.Vec3{0.0, 1.0, 0.0}
	arcballCamera.yaw = -45.0
	arcballCamera.zoom = 1.0
	game.camera = arcballCamera

	// setup shaders/matrices/meshes
	game.modelViewMatrix = mgl32.Ident4()
	game.normalMatrix = mgl32.Ident4()
	game.color = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}

	uniforms := map[string][]float32{
		"uMatP":     game.projMatrix[:],
		"uMatMV":    game.modelViewMatrix[:],
		"uMatNorm":  game.normalMatrix[:],
		"uColor":    game.color[:],
		"uMaterial": game.material[:],
		"uEyePos":   arcballCamera.eyePos[:],
		"uLightPos": game.lightPos[:],
	}

	game.phongShader, err = game.gl.NewShaderProgram(phongVertShaderCode, phongFragShaderCode, uniforms)
	if err != nil {
		return nil, err
	}

	game.gouraudShader, err = game.gl.NewShaderProgram(gouraudVertShaderCode, gouraudFragShaderCode, uniforms)
	if err != nil {
		return nil, err
	}

	game.blockMesh, err = game.gl.NewMesh(blockVerticies[:], blockNormals[:], blockIndicies[:])
	if err != nil {
		return nil, err
	}

	// setup edit mode
	game.IsEditModeEnabled = false
	game.editor = new(gameEditor)

	return game, nil
}

// Update updates the game models
func (game *Game) Update(dt float32, inputs map[GameInput]bool) {
	if game.IsGameOver {
		return
	}

	if inputs[GameInputEditModeToggle] {
		game.IsEditModeEnabled = !game.IsEditModeEnabled
	}

	if game.IsEditModeEnabled {
		player := game.player
		camera := game.camera.(*arcballCamera)
		eyePos := camera.eyePos
		game.Log = fmt.Sprintf(
			"FPS: %.2f<br/>Camera: (x:%.2f, y:%.2f, z:%.2f)<br/>Player: (x:%.2f, y:%.2f, z:%.2f)<br/>#WorldBlocks: %d<br/>#Enemies: %d",
			1000.0/dt,
			eyePos.X(),
			eyePos.Y(),
			eyePos.Z(),
			player.pos.X(),
			player.pos.Y(),
			player.pos.Z(),
			len(game.worldBlocks),
			len(game.enemies),
		)
	} else {
		game.Log = ""
	}

	game.player.update(game, dt, inputs)
	game.camera.update(game, dt, inputs)

	for _, enemy := range game.enemies {
		enemy.update(game, dt, inputs)
	}

	for _, worldBlock := range game.worldBlocks {
		worldBlock.update(game, dt, inputs)
	}

	if !game.IsEditModeEnabled {
		if game.player.pos.Y() < -10.0 {
			game.IsGameOver = true
			return
		}

		for _, enemy := range game.enemies {
			if checkForStaticOnStaticCollision(game.player, enemy) {
				game.IsGameOver = true
				return
			}
		}
	}

	game.editor.update(game, dt, inputs)
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

	// Render editor related items
	if err := game.editor.render(game, viewMatrix); err != nil {
		panic(err)
	}
}

// OnViewPortChange recalculates the projection matrix after a viewport adjustment
func (game *Game) OnViewPortChange() {
	game.gl.UpdateViewport()

	viewportWidth := float32(game.gl.GetViewportWidth())
	viewportHeight := float32(game.gl.GetViewportHeight())
	aspectRatio := viewportWidth / viewportHeight

	game.projMatrix = mgl32.Perspective(mgl32.DegToRad(45.0), aspectRatio, 1, 50.0)
}

func (game *Game) MovePlayerToPos(pos [3]float32) {
	game.player.pos = mgl32.Vec3(pos).Add(game.player.scale)
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

var phongVertShaderCode = `
	precision highp float;

	attribute vec3 aPosition;
	attribute vec3 aNormal;

	uniform mat4 uMatP;
	uniform mat4 uMatMV;
	uniform mat4 uMatNorm;

	varying vec3 vPos;
	varying vec3 vNorm;

	void main(void) {
		vec4 pos = uMatMV * vec4(aPosition, 1.);
		gl_Position = uMatP * pos;
		vPos = pos.xyz;
		vNorm = vec3(uMatNorm * vec4(aNormal, 0.0));
	}	
`

var phongFragShaderCode = `
	precision highp float;

	uniform vec4 uColor;
	uniform vec4 uMaterial;
	uniform vec3 uEyePos;
	uniform vec3 uLightPos;

	varying vec3 vPos;
	varying vec3 vNorm;

	void main(void) {
		float ka = uMaterial.x;
		float kd = uMaterial.y;
		float ks = uMaterial.z;
		float shininess = uMaterial.w;

		vec3 ambient = ka * uColor.rgb;
		
		float lightDist = length(uLightPos - vPos) * 0.8;
		vec3 L = normalize(uLightPos - vPos);
		vec3 N = normalize(vNorm);
		float lambert = max(dot(N, L), 0.0);
		vec3 diffuse = lambert * kd * uColor.rgb / lightDist;

		vec3 specular = vec3(0.0, 0.0, 0.0);
		if (lambert > 0.0) {
			vec3 R = reflect(-L, N);
			vec3 V = normalize(uEyePos);
			specular = pow(max(dot(R, V), 0.0), shininess) * ks * uColor.rgb;
		}

		gl_FragColor = vec4(ambient + diffuse + specular, 1.);
	}
`

var gouraudVertShaderCode = `
	precision highp float;

	attribute vec3 aPosition;
	attribute vec3 aNormal;

	uniform mat4 uMatP;
	uniform mat4 uMatMV;
	uniform mat4 uMatNorm;

	uniform vec4 uColor;
	uniform vec4 uMaterial;
	uniform vec3 uEyePos;
	uniform vec3 uLightPos;

	varying vec3 vColor;

	void main(void) {
		vec4 pos = uMatMV * vec4(aPosition, 1.);
		vec4 norm = uMatNorm * vec4(aNormal, 0.0);
		gl_Position = uMatP * pos;
		
		float ka = uMaterial.x;
		float kd = uMaterial.y;
		float ks = uMaterial.z;
		float shininess = uMaterial.w;

		float lightRadius = 100.0;
		float lightDist = distance(uLightPos, pos.xyz);
		float lightAttn = clamp(1.0 - (lightDist*lightDist) / (lightRadius*lightRadius), 0.0, 1.0);

		vec3 ambient = ka * uColor.rgb;

		vec3 L = normalize(uLightPos - pos.xyz);
		vec3 N = normalize(norm.xyz);
		float lambert = max(dot(N, L), 0.0);
		vec3 diffuse = lightAttn * lambert * kd * uColor.rgb;

		vec3 specular = vec3(0.0, 0.0, 0.0);
		if (lambert > 0.0) {
			vec3 R = reflect(-L, N);
			vec3 V = normalize(uEyePos);
			specular = pow(max(dot(R, V), 0.0), shininess) * ks * uColor.rgb;
		}

		vColor = ambient + diffuse + specular;
	}	
`

var gouraudFragShaderCode = `
	precision highp float;

	varying vec3 vColor;

	void main(void) {
		gl_FragColor = vec4(vColor, 1.);
	}
`

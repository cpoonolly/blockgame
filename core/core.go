package core

import (
	// "fmt"
	"github.com/go-gl/mathgl/mgl32"
	"math"
)

func f32Abs(num float32) float32 {
	return float32(math.Abs(float64(num)))
}

func f32Round(num float32, decimalPlaces uint16) float32 {
	multiplier := math.Pow10(int(decimalPlaces))
	return float32(math.Round(float64(num)*multiplier) / multiplier)
}

func f32Max(num1, num2 float32) float32 {
	return float32(math.Max(float64(num1), float64(num2)))
}

func f32Min(num1, num2 float32) float32 {
	return float32(math.Min(float64(num1), float64(num2)))
}

func f32LimitBetween(num, min, max float32) float32 {
	return f32Min(f32Max(num, min), max)
}

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
	pos      mgl32.Vec3
	scale    mgl32.Vec3
	velocity mgl32.Vec3
	color    mgl32.Vec4
}

type camera interface {
	getViewMatrix() mgl32.Mat4
}

type arcballCamera struct {
	lookAt mgl32.Vec3
	zoom   float32
	yaw    float32
	up     mgl32.Vec3
}

func (camera *arcballCamera) getViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(camera.getEyePos(), camera.lookAt, camera.up)
}

func (camera *arcballCamera) getEyePos() mgl32.Vec3 {
	zoom := camera.zoom
	yaw := camera.yaw

	relPos := mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
	relPos = mgl32.HomogRotate3DY(yaw).Mul4x1(relPos)
	relPos = mgl32.Scale3D(1+(zoom/2), 1+(zoom/2), 1+(zoom/2)).Mul4x1(relPos)
	relPos = mgl32.Translate3D(0.0, zoom, 0.0).Mul4x1(relPos)

	return camera.lookAt.Add(relPos.Vec3())
}

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
	// GameInputEditModeBlockCreate creates a new block
	GameInputEditModeBlockCreate
	// GameInputEditModeBlockDelete deletes a block
	GameInputEditModeBlockDelete
)

// player moves 1 unit per second
const playerSpeed float32 = 10

// camera move .5 units per second
const cameraSpeed float32 = 100

// gravity not acceleration based but based on pure veloctiy
const gravitySpeed float32 = 10

// Game represents a game
type Game struct {
	gl          GlContext
	blockShader ShaderProgram
	blockMesh   Mesh

	projMatrix      mgl32.Mat4
	modelViewMatrix mgl32.Mat4
	normalMatrix    mgl32.Mat4
	color           mgl32.Vec4

	playerBlock    *block
	worldBlocks    []*block
	camera         camera
	isEditMode     bool
	editorNewBlock *block

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

	game.playerBlock = new(block)
	game.playerBlock.scale = mgl32.Vec3{0.5, 0.5, 0.5}
	game.playerBlock.color = mgl32.Vec4{0.9, 0.9, 0.9, 1.0}

	// generate world blocks
	game.worldBlocks = make([]*block, 3)

	// world block 1
	game.worldBlocks[0] = new(block)
	game.worldBlocks[0].pos = mgl32.Vec3{3.0, 0.0, 0.0}
	game.worldBlocks[0].scale = mgl32.Vec3{1.0, 1.0, 1.0}
	game.worldBlocks[0].color = mgl32.Vec4{0.1, 1.0, 0.1, 1.0}

	// world block 2
	game.worldBlocks[1] = new(block)
	game.worldBlocks[1].pos = mgl32.Vec3{-5.0, 0.0, 0.0}
	game.worldBlocks[1].scale = mgl32.Vec3{2.0, 2.0, 2.0}
	game.worldBlocks[1].color = mgl32.Vec4{1.0, 0.1, 0.1, 1.0}

	// world block 3
	game.worldBlocks[2] = new(block)
	game.worldBlocks[2].pos = mgl32.Vec3{0.0, -3.0, 0.0}
	game.worldBlocks[2].scale = mgl32.Vec3{10.0, 1.0, 10.0}
	game.worldBlocks[2].color = mgl32.Vec4{0.1, 0.1, 1.0, 1.0}

	// create a camera
	arcballCamera := new(arcballCamera)
	arcballCamera.up = mgl32.Vec3{0.0, 1.0, 0.0}
	game.camera = arcballCamera

	game.isEditMode = true

	return game, nil
}

// Update updates the game models
func (game *Game) Update(dt float32, inputs map[GameInput]bool) {
	/*
		player := game.playerBlock
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
	*/

	if inputs[GameInputEditModeToggle] {
		game.isEditMode = !game.isEditMode
	}

	game.updatePlayerBlock(dt, inputs)
	game.updateCamera(dt, inputs)
}

func (game *Game) updatePlayerBlock(dt float32, inputs map[GameInput]bool) {
	player := game.playerBlock

	var vx, vy, vz float32
	if inputs[GameInputPlayerMoveLeft] {
		vx = playerSpeed
	} else if inputs[GameInputPlayerMoveRight] {
		vx = -1 * playerSpeed
	}
	if inputs[GameInputPlayerMoveForward] {
		vz = playerSpeed
	} else if inputs[GameInputPlayerMoveBack] {
		vz = -1 * playerSpeed
	}

	if !game.isEditMode {
		vy = -1 * gravitySpeed
	} else {
		if inputs[GameInputEditModeMoveUp] {
			vy = playerSpeed
		} else if inputs[GameInputEditModeMoveDown] {
			vy = -1 * playerSpeed
		}
	}

	player.velocity = mgl32.Vec3{vx, vy, vz}
	// game.Log += fmt.Sprintf("<br/>Player Velocity: (vx: %.2f\tvy: %.2f\tvz: %.2f)\n", player.velocity.X(), player.velocity.Y(), player.velocity.Z())

	collisions := game.checkForCollisions(dt, player, game.worldBlocks)
	if !game.isEditMode {
		player.pos = game.processCollisions(dt, player, collisions)
	} else {
		player.pos = player.pos.Add(player.velocity.Mul(dt / 1000))
	}
}

func (game *Game) updateCamera(dt float32, inputs map[GameInput]bool) {
	camera := game.camera.(*arcballCamera)
	player := game.playerBlock

	var dyaw float32
	if inputs[GameInputCameraRotateLeft] {
		dyaw = -1 * cameraSpeed / 1000.0
	} else if inputs[GameInputCameraRotateRight] {
		dyaw = cameraSpeed / 1000.0
	}

	var dzoom float32
	if inputs[GameInputCameraZoomIn] {
		dzoom = -1 * cameraSpeed / 1000.0
	} else if inputs[GameInputCameraZoomOut] {
		dzoom = cameraSpeed / 1000.0
	}

	camera.lookAt = player.pos
	camera.yaw = f32LimitBetween(camera.yaw+dyaw, 1.5, 2.5)
	camera.zoom = f32LimitBetween(camera.zoom+dzoom, 1.0, 5.0)
	// fmt.Printf("camera: (yaw: %.5f, zoom: %.5f)\n", camera.yaw, camera.zoom)
}

func (game *Game) checkForCollisions(dt float32, blk *block, collidables []*block) []*block {
	dLeft := dt / 1000 * f32Max(0.0, blk.velocity.X())
	dRight := dt / 1000 * f32Max(0.0, -1*blk.velocity.X())
	dUp := dt / 1000 * f32Max(0.0, blk.velocity.Y())
	dDown := dt / 1000 * f32Max(0.0, -1*blk.velocity.Y())
	dForward := dt / 1000 * f32Max(0.0, blk.velocity.Z())
	dBackward := dt / 1000 * f32Max(0.0, -1*blk.velocity.Z())

	blkLeft := blk.pos.X() + blk.scale.X()
	blkRight := blk.pos.X() - blk.scale.X()
	blkTop := blk.pos.Y() + blk.scale.Y()
	blkBottom := blk.pos.Y() - blk.scale.Y()
	blkFront := blk.pos.Z() + blk.scale.Z()
	blkBack := blk.pos.Z() - blk.scale.Z()

	var collisions []*block

	for _, collidable := range collidables {
		collidableLeft := collidable.pos.X() + collidable.scale.X()
		if blkRight-dRight >= collidableLeft {
			continue
		}

		collidableRight := collidable.pos.X() - collidable.scale.X()
		if blkLeft+dLeft <= collidableRight {
			continue
		}

		collidableTop := collidable.pos.Y() + collidable.scale.Y()
		if blkBottom-dDown >= collidableTop {
			continue
		}

		collidableBottom := collidable.pos.Y() - collidable.scale.Y()
		if blkTop+dUp <= collidableBottom {
			continue
		}

		collidableFront := collidable.pos.Z() + collidable.scale.Z()
		if blkBack-dBackward >= collidableFront {
			continue
		}

		collidableBack := collidable.pos.Z() - collidable.scale.Z()
		if blkFront+dForward <= collidableBack {
			continue
		}

		// fmt.Printf("collision detected - block %d\n", blockNum)
		collisions = append(collisions, collidable)
	}

	return collisions
}

func (game *Game) processCollisions(dt float32, blk *block, collisions []*block) mgl32.Vec3 {
	dLeft := dt / 1000 * f32Max(0.0, blk.velocity.X())
	dRight := dt / 1000 * f32Max(0.0, -1*blk.velocity.X())
	dUp := dt / 1000 * f32Max(0.0, blk.velocity.Y())
	dDown := dt / 1000 * f32Max(0.0, -1*blk.velocity.Y())
	dForward := dt / 1000 * f32Max(0.0, blk.velocity.Z())
	dBackward := dt / 1000 * f32Max(0.0, -1*blk.velocity.Z())

	blkLeft := blk.pos.X() + blk.scale.X()
	blkRight := blk.pos.X() - blk.scale.X()
	blkTop := blk.pos.Y() + blk.scale.Y()
	blkBottom := blk.pos.Y() - blk.scale.Y()
	blkFront := blk.pos.Z() + blk.scale.Z()
	blkBack := blk.pos.Z() - blk.scale.Z()

	for _, collision := range collisions {
		collisionLeft := collision.pos.X() + collision.scale.X()
		collisionRight := collision.pos.X() - collision.scale.X()
		collisionTop := collision.pos.Y() + collision.scale.Y()
		collisionBottom := collision.pos.Y() - collision.scale.Y()
		collisionFront := collision.pos.Z() + collision.scale.Z()
		collisionBack := collision.pos.Z() - collision.scale.Z()

		// fmt.Println("Processing Collision:")
		// fmt.Printf("collision: left: %.5f\tright: %.5f\ttop: %.5f,\tbottom: %.5f,\tfront: %.5f\tback: %.5f\n", collisionLeft, collisionRight, collisionTop, collisionBottom, collisionFront, collisionBack)
		// fmt.Printf("blk: left: %.5f\tright: %.5f\ttop: %.5f,\tbottom: %.5f,\tfront: %.5f\tback: %.5f\n", blkLeft, blkRight, blkTop, blkBottom, blkFront, blkBack)
		// fmt.Printf("dLeft: %.5f\tdRight: %.5f\tdUp: %.5f\tdDown: %.5f\tdForward: %.5f\tdBackward: %.5f\n", dLeft, dRight, dUp, dDown, dForward, dBackward)

		// find the axis that collides later in time and adjust it
		var timeOfXAxisCollision, timeOfYAxisCollision, timeOfZAxisCollision float32
		var distanceToCollisionX, distanceToCollisionY, distanceToCollisionZ float32

		if dLeft > 0 {
			distanceToCollisionX = collisionRight - blkLeft
			timeOfXAxisCollision = dt * distanceToCollisionX / dLeft
		}

		if dRight > 0 {
			distanceToCollisionX = blkRight - collisionLeft
			timeOfXAxisCollision = dt * distanceToCollisionX / dRight
		}

		if dUp > 0 {
			distanceToCollisionY = collisionBottom - blkTop
			timeOfYAxisCollision = dt * distanceToCollisionY / dUp
		}

		if dDown > 0 {
			distanceToCollisionY = blkBottom - collisionTop
			timeOfYAxisCollision = dt * distanceToCollisionY / dDown
		}

		if dForward > 0 {
			distanceToCollisionZ = collisionBack - blkFront
			timeOfZAxisCollision = dt * distanceToCollisionZ / dForward
		}

		if dBackward > 0 {
			distanceToCollisionZ = blkBack - collisionFront
			timeOfZAxisCollision = dt * distanceToCollisionZ / dBackward
		}

		// fmt.Printf("timeOfXAxisCollision: %.5f\ttimeOfYAxisCollision: %.5f\ttimeOfZAxisCollision: %.5f\n", timeOfXAxisCollision, timeOfYAxisCollision, timeOfZAxisCollision)
		// fmt.Printf("distanceToCollisionX: %.5f\tdistanceToCollisionY: %.5f\tdistanceToCollisionZ: %.5f\n", distanceToCollisionX, distanceToCollisionY, distanceToCollisionZ)
		if timeOfXAxisCollision >= timeOfYAxisCollision && timeOfXAxisCollision >= timeOfZAxisCollision {
			if dLeft > dRight {
				// fmt.Println("collision - left side")
				dLeft = distanceToCollisionX
			} else {
				// fmt.Println("collision - right side")
				dRight = distanceToCollisionX
			}
		}

		if timeOfYAxisCollision >= timeOfXAxisCollision && timeOfYAxisCollision >= timeOfZAxisCollision {
			if dUp > dDown {
				// fmt.Println("collision - top side")
				dUp = distanceToCollisionY
			} else {
				// fmt.Println("collision - bottom side")
				dDown = distanceToCollisionY
			}
		}

		if timeOfZAxisCollision >= timeOfXAxisCollision && timeOfZAxisCollision >= timeOfYAxisCollision {
			if dForward > dBackward {
				// fmt.Println("collision - front side")
				dForward = distanceToCollisionZ
			} else {
				// fmt.Println("collision - back side")
				dBackward = distanceToCollisionZ
			}
		}
	}

	newPos := blk.pos.Add(mgl32.Vec3{dLeft - dRight, dUp - dDown, dForward - dBackward})
	newPos[0] = f32Round(newPos[0], 2)
	newPos[1] = f32Round(newPos[1], 2)
	newPos[2] = f32Round(newPos[2], 2)

	/*
		if len(collisions) > 0 {
			newPosLeft := newPos.X() + blk.scale.X()
			newPosRight := newPos.X() - blk.scale.X()
			newPosTop := newPos.Y() + blk.scale.Y()
			newPosBottom := newPos.Y() - blk.scale.Y()
			newPosFront := newPos.Z() + blk.scale.Z()
			newPosBack := newPos.Z() - blk.scale.Z()

			fmt.Println("Collisions Processed:")
			fmt.Printf("dLeft: %.5f\tdRight: %.5f\tdUp: %.5f\tdDown: %.5f\tdForward: %.5f\tdBackward: %.5f\n", dLeft, dRight, dUp, dDown, dForward, dBackward)
			fmt.Printf("left: %.5f\tright: %.5f\ttop: %.5f,\tbottom: %.5f,\tfront: %.5f\tback: %.5f\n", newPosLeft, newPosRight, newPosTop, newPosBottom, newPosFront, newPosBack)
		}
	*/

	return newPos
}

// Render renders the frame
func (game *Game) Render() {
	if err := game.gl.ClearScreen(0.0, 0.0, 0.0); err != nil {
		panic(err)
	}

	viewMatrix := game.camera.getViewMatrix()

	// Render player
	if err := game.renderBlock(game.playerBlock, viewMatrix); err != nil {
		panic(err)
	}

	// Render world
	for _, block := range game.worldBlocks {
		if err := game.renderBlock(block, viewMatrix); err != nil {
			panic(err)
		}
	}
}

func (game *Game) renderBlock(block *block, viewMatrix mgl32.Mat4) error {
	scaleMatrix := mgl32.Scale3D(block.scale.X(), block.scale.Y(), block.scale.Z())
	translateMatrix := mgl32.Translate3D(block.pos.X(), block.pos.Y(), block.pos.Z())

	modelMatrix := mgl32.Ident4().Mul4(translateMatrix).Mul4(scaleMatrix)

	// not magic - shader is initialized with pointers to these values as uniforms
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

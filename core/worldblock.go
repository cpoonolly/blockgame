package core

import (
	"github.com/go-gl/mathgl/mgl32"
)

var worldBlockColorDefault = mgl32.Vec4{0.7, 0.7, 0.7, 1.0}
var worldBlockColorHighlighted = mgl32.Vec4{.99, .84, .20, 1.0}

type worldBlock struct {
	pos   mgl32.Vec3
	scale mgl32.Vec3
	color mgl32.Vec4
}

func (worldBlock *worldBlock) left() float32 {
	return worldBlock.pos.X() + worldBlock.scale.X()
}

func (worldBlock *worldBlock) right() float32 {
	return worldBlock.pos.X() - worldBlock.scale.X()
}

func (worldBlock *worldBlock) top() float32 {
	return worldBlock.pos.Y() + worldBlock.scale.Y()
}

func (worldBlock *worldBlock) bottom() float32 {
	return worldBlock.pos.Y() - worldBlock.scale.Y()
}

func (worldBlock *worldBlock) front() float32 {
	return worldBlock.pos.Z() + worldBlock.scale.Z()
}

func (worldBlock *worldBlock) back() float32 {
	return worldBlock.pos.Z() - worldBlock.scale.Z()
}

func (worldBlock *worldBlock) update(game *Game, dt float32, inputs map[GameInput]bool) {
	if game.IsEditModeEnabled && checkForStaticOnStaticCollision(game.player, worldBlock) {
		worldBlock.color = worldBlockColorHighlighted
	} else {
		worldBlock.color = worldBlockColorDefault
	}
}

func (worldBlock *worldBlock) render(game *Game, viewMatrix mgl32.Mat4) error {
	scaleMatrix := mgl32.Scale3D(worldBlock.scale.X(), worldBlock.scale.Y(), worldBlock.scale.Z())
	translateMatrix := mgl32.Translate3D(worldBlock.pos.X(), worldBlock.pos.Y(), worldBlock.pos.Z())

	modelMatrix := mgl32.Ident4().Mul4(translateMatrix).Mul4(scaleMatrix)

	// not magic - shader is initialized with pointers to these values as uniforms
	game.modelViewMatrix = viewMatrix.Mul4(modelMatrix)
	game.normalMatrix = game.modelViewMatrix.Inv().Transpose()
	game.color = worldBlock.color
	game.material = mgl32.Vec4{0.1, 0.6, 0.0, 20.0}
	game.lightPos = viewMatrix.Mul4x1(game.player.pos.Vec4(1.0)).Vec3()

	shader := game.gouraudShader
	if worldBlock.scale.X() > 3.0 || worldBlock.scale.Y() > 3.0 || worldBlock.scale.Z() > 3.0 {
		shader = game.phongShader
	}

	if err := game.gl.RenderTriangles(game.blockMesh, shader); err != nil {
		return err
	}

	return nil
}

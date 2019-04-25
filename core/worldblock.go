package core

import (
	"github.com/go-gl/mathgl/mgl32"
)

type worldBlock struct {
	id    uint32
	pos   mgl32.Vec3
	scale mgl32.Vec3
	color mgl32.Vec4
}

func (worldBlock *worldBlock) position() mgl32.Vec3 {
	return worldBlock.pos
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

func (worldBlock *worldBlock) render(game *Game, viewMatrix mgl32.Mat4) error {
	scaleMatrix := mgl32.Scale3D(worldBlock.scale.X(), worldBlock.scale.Y(), worldBlock.scale.Z())
	translateMatrix := mgl32.Translate3D(worldBlock.pos.X(), worldBlock.pos.Y(), worldBlock.pos.Z())

	modelMatrix := mgl32.Ident4().Mul4(translateMatrix).Mul4(scaleMatrix)

	// not magic - shader is initialized with pointers to these values as uniforms
	game.modelViewMatrix = viewMatrix.Mul4(modelMatrix)
	game.normalMatrix = game.modelViewMatrix.Inv().Transpose()
	game.color = worldBlock.color

	if err := game.gl.RenderTriangles(game.blockMesh, game.blockShader); err != nil {
		return err
	}

	return nil
}

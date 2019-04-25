package core

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
)

const enemyAcceleration float32 = playerAcceleration * .75

type enemy struct {
	start mgl32.Vec3
	pos   mgl32.Vec3
	scale mgl32.Vec3
	vel   mgl32.Vec3
	color mgl32.Vec4
}

func (enemy *enemy) position() mgl32.Vec3 {
	return enemy.pos
}

func (enemy *enemy) velocity() mgl32.Vec3 {
	return enemy.vel
}

func (enemy *enemy) left() float32 {
	return enemy.pos.X() + enemy.scale.X()
}

func (enemy *enemy) right() float32 {
	return enemy.pos.X() - enemy.scale.X()
}

func (enemy *enemy) top() float32 {
	return enemy.pos.Y() + enemy.scale.Y()
}

func (enemy *enemy) bottom() float32 {
	return enemy.pos.Y() - enemy.scale.Y()
}

func (enemy *enemy) front() float32 {
	return enemy.pos.Z() + enemy.scale.Z()
}

func (enemy *enemy) back() float32 {
	return enemy.pos.Z() - enemy.scale.Z()
}

func (enemy *enemy) update(game *Game, dt float32, inputs map[GameInput]bool) {
	playerPos := game.player.pos
	enemyPos := enemy.pos

	if game.IsEditModeEnabled {
		return
	}

	var dvx, dvy, dvz float32
	if playerPos.X() > enemyPos.X() {
		dvx = enemyAcceleration
	} else if playerPos.X() < enemyPos.X() {
		dvx = -1 * enemyAcceleration
	} else if enemy.vel.X() != 0 {
		dvx = -1 * enemy.vel.X() / f32Abs(enemy.vel.X()) * dampening
	}

	if playerPos.Z() > enemyPos.Z() {
		dvz = enemyAcceleration
	} else if playerPos.Z() < enemyPos.Z() {
		dvz = -1 * enemyAcceleration
	} else if enemy.vel.Z() != 0 {
		dvz = -1 * enemy.vel.Z() / f32Abs(enemy.vel.Z()) * dampening
	}

	dvy = -1 * gravityAcceleration

	enemy.vel = enemy.vel.Add(mgl32.Vec3{dvx, dvy, dvz})
	enemy.vel[0] = f32LimitBetween(enemy.vel[0], -1*maxVelocity, maxVelocity)
	enemy.vel[1] = f32LimitBetween(enemy.vel[1], -1*maxVelocity, maxVelocity)
	enemy.vel[2] = f32LimitBetween(enemy.vel[2], -1*maxVelocity, maxVelocity)

	dPos := enemy.vel.Mul(dt / 1000)
	for _, worldBlock := range game.worldBlocks {
		if checkForStaticCollision(dt, dPos, enemy, worldBlock) {
			dPos = processStaticCollision(dPos, getStaticCollisionDetails(dt, dPos, enemy, worldBlock))
		}
	}

	enemy.pos = enemy.pos.Add(dPos)
}

func (enemy *enemy) render(game *Game, viewMatrix mgl32.Mat4) error {
	scaleMatrix := mgl32.Scale3D(enemy.scale.X(), enemy.scale.Y(), enemy.scale.Z())
	translateMatrix := mgl32.Translate3D(enemy.pos.X(), enemy.pos.Y(), enemy.pos.Z())

	modelMatrix := mgl32.Ident4().Mul4(translateMatrix).Mul4(scaleMatrix)

	// not magic - shader is initialized with pointers to these values as uniforms
	game.modelViewMatrix = viewMatrix.Mul4(modelMatrix)
	game.normalMatrix = game.modelViewMatrix.Inv().Transpose()
	game.color = enemy.color

	if err := game.gl.RenderTriangles(game.blockMesh, game.blockShader); err != nil {
		return err
	}

	return nil
}

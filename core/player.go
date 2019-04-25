package core

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
)

type player struct {
	pos   mgl32.Vec3
	scale mgl32.Vec3
	vel   mgl32.Vec3
	color mgl32.Vec4
}

func (player *player) position() mgl32.Vec3 {
	return player.pos
}

func (player *player) velocity() mgl32.Vec3 {
	return player.vel
}

func (player *player) left() float32 {
	return player.pos.X() + player.scale.X()
}

func (player *player) right() float32 {
	return player.pos.X() - player.scale.X()
}

func (player *player) top() float32 {
	return player.pos.Y() + player.scale.Y()
}

func (player *player) bottom() float32 {
	return player.pos.Y() - player.scale.Y()
}

func (player *player) front() float32 {
	return player.pos.Z() + player.scale.Z()
}

func (player *player) back() float32 {
	return player.pos.Z() - player.scale.Z()
}

func (player *player) update(game *Game, dt float32, inputs map[GameInput]bool) {
	var dvx, dvy, dvz float32
	if inputs[GameInputPlayerMoveLeft] {
		dvx = playerAcceleration
	} else if inputs[GameInputPlayerMoveRight] {
		dvx = -1 * playerAcceleration
	} else if player.vel.X() != 0 {
		dvx = -1 * player.vel.X() / f32Abs(player.vel.X()) * dampening
	}

	if inputs[GameInputPlayerMoveForward] {
		dvz = playerAcceleration
	} else if inputs[GameInputPlayerMoveBack] {
		dvz = -1 * playerAcceleration
	} else if player.vel.Z() != 0 {
		dvz = -1 * player.vel.Z() / f32Abs(player.vel.Z()) * dampening
	}

	if !game.IsEditModeEnabled {
		dvy = -1 * gravityAcceleration
	} else {
		if inputs[GameInputEditModeMoveUp] {
			dvy = playerAcceleration
		} else if inputs[GameInputEditModeMoveDown] {
			dvy = -1 * playerAcceleration
		} else if player.vel.Y() != 0 {
			dvy = -1 * player.vel.Y() / f32Abs(player.vel.Y()) * dampening
		}
	}

	player.vel = player.vel.Add(mgl32.Vec3{dvx, dvy, dvz})
	player.vel[0] = f32LimitBetween(player.vel[0], -1*maxVelocity, maxVelocity)
	player.vel[1] = f32LimitBetween(player.vel[1], -1*maxVelocity, maxVelocity)
	player.vel[2] = f32LimitBetween(player.vel[2], -1*maxVelocity, maxVelocity)

	if game.IsEditModeEnabled {
		game.Log += fmt.Sprintf("<br/>Player Velocity: (vx: %.2f\tvy: %.2f\tvz: %.2f)\n", player.vel.X(), player.vel.Y(), player.vel.Z())
	}

	dPos := game.player.vel.Mul(dt / 1000)
	if !game.IsEditModeEnabled {
		for _, worldBlock := range game.worldBlocks {
			if checkForStaticCollision(dt, dPos, player, worldBlock) {
				dPos = processStaticCollision(dPos, getStaticCollisionDetails(dt, dPos, game.player, worldBlock))
			}
		}
	}

	player.pos = player.pos.Add(dPos)
}

func (player *player) render(game *Game, viewMatrix mgl32.Mat4) error {
	scaleMatrix := mgl32.Scale3D(player.scale.X(), player.scale.Y(), player.scale.Z())
	translateMatrix := mgl32.Translate3D(player.pos.X(), player.pos.Y(), player.pos.Z())

	modelMatrix := mgl32.Ident4().Mul4(translateMatrix).Mul4(scaleMatrix)

	// not magic - shader is initialized with pointers to these values as uniforms
	game.modelViewMatrix = viewMatrix.Mul4(modelMatrix)
	game.normalMatrix = game.modelViewMatrix.Inv().Transpose()
	game.color = player.color

	if err := game.gl.RenderTriangles(game.blockMesh, game.blockShader); err != nil {
		return err
	}

	return nil
}

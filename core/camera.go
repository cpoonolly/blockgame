package core

import (
	"github.com/go-gl/mathgl/mgl32"
)

const cameraSpeed float32 = 100

type camera interface {
	gameUpdatable
	getViewMatrix() mgl32.Mat4
}

type arcballCamera struct {
	lookAt mgl32.Vec3
	zoom   float32
	yaw    float32
	up     mgl32.Vec3
}

func (camera *arcballCamera) getEyePos() mgl32.Vec3 {
	zoom := camera.zoom
	yaw := camera.yaw

	relPos := mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
	relPos = mgl32.HomogRotate3DY(yaw).Mul4x1(relPos)
	relPos = mgl32.Scale3D(1+(zoom/1.5), 1+(zoom/1.5), 1+(zoom/1.5)).Mul4x1(relPos)
	relPos = mgl32.Translate3D(0.0, zoom, 0.0).Mul4x1(relPos)

	return camera.lookAt.Add(relPos.Vec3())
}

func (camera *arcballCamera) getViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(camera.getEyePos(), camera.lookAt, camera.up)
}

func (camera *arcballCamera) update(game *Game, dt float32, inputs map[GameInput]bool) {
	player := game.player

	var dyaw float32
	if inputs[GameInputCameraRotateLeft] {
		dyaw = cameraSpeed / 1000.0
	} else if inputs[GameInputCameraRotateRight] {
		dyaw = -1 * cameraSpeed / 1000.0
	}

	var dzoom float32
	if inputs[GameInputCameraZoomIn] {
		dzoom = -1 * cameraSpeed / 1000.0
	} else if inputs[GameInputCameraZoomOut] {
		dzoom = cameraSpeed / 1000.0
	}

	camera.lookAt = player.pos
	camera.yaw = f32LimitBetween(camera.yaw+dyaw, 1.5, 3.5)
	camera.zoom = f32LimitBetween(camera.zoom+dzoom, 1.0, 5.0)
}

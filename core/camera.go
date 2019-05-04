package core

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
)

const cameraSpeed float32 = 100
const cameraRotateSpeed float32 = 5000

type camera interface {
	gameUpdatable
	getViewMatrix() mgl32.Mat4
}

type arcballCamera struct {
	zoom   float32
	yaw    float32
	lookAt mgl32.Vec3
	up     mgl32.Vec3
	eyePos mgl32.Vec3
}

func (camera *arcballCamera) getViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(camera.eyePos, camera.lookAt, camera.up)
}

func (camera *arcballCamera) update(game *Game, dt float32, inputs map[GameInput]bool) {
	player := game.player

	var dyaw float32
	if inputs[GameInputCameraRotateLeft] {
		dyaw = cameraRotateSpeed / 1000.0
	} else if inputs[GameInputCameraRotateRight] {
		dyaw = -1 * cameraRotateSpeed / 1000.0
	}

	var dzoom float32
	if inputs[GameInputCameraZoomIn] {
		dzoom = -1 * cameraSpeed / 1000.0
	} else if inputs[GameInputCameraZoomOut] {
		dzoom = cameraSpeed / 1000.0
	}

	camera.lookAt = player.pos
	camera.yaw = camera.yaw + dyaw    // f32LimitBetween(, 1.5, 3.5)
	camera.zoom = camera.zoom + dzoom // f32LimitBetween(, 1.0, 5.0)
	camera.eyePos = computeEyePos(camera)

	if game.IsEditModeEnabled {
		game.Log += fmt.Sprintf("<br/>Camera: (zoom: %.2f\tyaw: %.2f)\n", camera.zoom, camera.yaw)
	}
}

func computeEyePos(camera *arcballCamera) mgl32.Vec3 {
	zoom := camera.zoom
	yaw := camera.yaw

	relPos := mgl32.Vec4{0.0, 1.0, 1.0, 1.0}
	relPos = mgl32.HomogRotate3DY(mgl32.DegToRad(yaw)).Mul4x1(relPos)
	relPos = mgl32.Scale3D(5.0+(.3*zoom), 5.0+(.3*zoom), 5.0+(.3*zoom)).Mul4x1(relPos)
	relPos = mgl32.Translate3D(0.0, zoom, 0.0).Mul4x1(relPos)

	return camera.lookAt.Add(relPos.Vec3())
}

package core

import (
	// "fmt"
	"github.com/go-gl/mathgl/mgl32"
)

type staticCollidable interface {
	position() mgl32.Vec3
	left() float32
	right() float32
	top() float32
	bottom() float32
	front() float32
	back() float32
}

type dynamicCollidable interface {
	staticCollidable
	gameUpdatable
	velocity() mgl32.Vec3
}

type staticCollisionDetails struct {
	timeOfXAxisCollision, timeOfYAxisCollision, timeOfZAxisCollision float32
	distanceToCollisionX, distanceToCollisionY, distanceToCollisionZ float32
}

func checkForStaticCollision(dt float32, dPos mgl32.Vec3, dynamic dynamicCollidable, static staticCollidable) bool {
	dRight := f32Max(0.0, -1*dPos.X())
	if dynamic.right()-dRight >= static.left() {
		return false
	}

	dLeft := f32Max(0.0, dPos.X())
	if dynamic.left()+dLeft <= static.right() {
		return false
	}

	dDown := f32Max(0.0, -1*dPos.Y())
	if dynamic.bottom()-dDown >= static.top() {
		return false
	}

	dUp := f32Max(0.0, dPos.Y())
	if dynamic.top()+dUp <= static.bottom() {
		return false
	}

	dBackward := f32Max(0.0, -1*dPos.Z())
	if dynamic.back()-dBackward >= static.front() {
		return false
	}

	dForward := f32Max(0.0, dPos.Z())
	if dynamic.front()+dForward <= static.back() {
		return false
	}

	return true
}

func getStaticCollisionDetails(dt float32, dPos mgl32.Vec3, dynamic dynamicCollidable, static staticCollidable) staticCollisionDetails {
	var collisionDetails staticCollisionDetails

	if dPos.X() > 0 {
		collisionDetails.distanceToCollisionX = static.right() - dynamic.left()
		collisionDetails.timeOfXAxisCollision = dt * collisionDetails.distanceToCollisionX / dPos.X()
	} else if dPos.X() < 0 {
		collisionDetails.distanceToCollisionX = static.left() - dynamic.right()
		collisionDetails.timeOfXAxisCollision = dt * collisionDetails.distanceToCollisionX / dPos.X()
	}

	if dPos.Y() > 0 {
		collisionDetails.distanceToCollisionY = static.bottom() - dynamic.top()
		collisionDetails.timeOfYAxisCollision = dt * collisionDetails.distanceToCollisionY / dPos.Y()
	} else if dPos.Y() < 0 {
		collisionDetails.distanceToCollisionY = static.top() - dynamic.bottom()
		collisionDetails.timeOfYAxisCollision = dt * collisionDetails.distanceToCollisionY / dPos.Y()
	}

	if dPos.Z() > 0 {
		collisionDetails.distanceToCollisionZ = static.back() - dynamic.front()
		collisionDetails.timeOfZAxisCollision = dt * collisionDetails.distanceToCollisionZ / dPos.Z()
	} else if dPos.Z() < 0 {
		collisionDetails.distanceToCollisionZ = static.front() - dynamic.back()
		collisionDetails.timeOfZAxisCollision = dt * collisionDetails.distanceToCollisionZ / dPos.Z()
	}

	return collisionDetails
}

func processStaticCollision(dPos mgl32.Vec3, details staticCollisionDetails) mgl32.Vec3 {
	// find the axis that collides later in time and adjust it

	if details.timeOfXAxisCollision >= details.timeOfYAxisCollision && details.timeOfXAxisCollision >= details.timeOfZAxisCollision {
		dPos[0] = details.distanceToCollisionX // if dX >= 0 then left side else right side
	}

	if details.timeOfYAxisCollision >= details.timeOfXAxisCollision && details.timeOfYAxisCollision >= details.timeOfZAxisCollision {
		dPos[1] = details.distanceToCollisionY
	}

	if details.timeOfZAxisCollision >= details.timeOfXAxisCollision && details.timeOfZAxisCollision >= details.timeOfYAxisCollision {
		dPos[2] = details.distanceToCollisionZ
	}

	return dPos
}

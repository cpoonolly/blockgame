package core

import (
	// "fmt"
	"github.com/go-gl/mathgl/mgl32"
)

type collidable interface {
	left() float32
	right() float32
	top() float32
	bottom() float32
	front() float32
	back() float32
}

// checks to see if 2 static collidables are colliding
func checkForStaticOnStaticCollision(static1, static2 collidable) bool {
	return static1.right() < static2.left() &&
		static1.left() > static2.right() &&
		static1.bottom() < static2.top() &&
		static1.top() > static2.bottom() &&
		static1.back() < static2.front() &&
		static1.front() > static2.back()
}

// checks for a collision between a "dynamic" (moving) collidable (dPos = it's change in position) & a static collidable
func checkForDynamicOnStaticCollision(dPos mgl32.Vec3, dynamic, static collidable) bool {
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

// processes a collision between a "dynamic" (moving) collidable (dPos = it's change in position) & a static collidable
func processDynamicOnStaticCollisionDetails(dt float32, dPos mgl32.Vec3, dynamic, static collidable) mgl32.Vec3 {
	/*
		Very simple collision correction:

		STEP 1
		Find out axis/side where the collision occurs. A collision occurs when blocks "intersect" on all 3 xyz axis. The last axis to intersect
		during a collision tells you the axis/side/face on which the collision occurs. If more than one axis "intersect" last - then it's a
		corner/edge collision


		STEP 2
		Adjust dPos only on the axis of collision to right before the collision happens. This creates a "slide" effect.
	*/

	var timeOfXAxisIntersection, timeOfYAxisIntersection, timeOfZAxisIntersection float32
	var distanceToCollisionX, distanceToCollisionY, distanceToCollisionZ float32

	if dPos.X() > 0 {
		distanceToCollisionX = static.right() - dynamic.left()
		timeOfXAxisIntersection = dt * distanceToCollisionX / dPos.X()
	} else if dPos.X() < 0 {
		distanceToCollisionX = static.left() - dynamic.right()
		timeOfXAxisIntersection = dt * distanceToCollisionX / dPos.X()
	}

	if dPos.Y() > 0 {
		distanceToCollisionY = static.bottom() - dynamic.top()
		timeOfYAxisIntersection = dt * distanceToCollisionY / dPos.Y()
	} else if dPos.Y() < 0 {
		distanceToCollisionY = static.top() - dynamic.bottom()
		timeOfYAxisIntersection = dt * distanceToCollisionY / dPos.Y()
	}

	if dPos.Z() > 0 {
		distanceToCollisionZ = static.back() - dynamic.front()
		timeOfZAxisIntersection = dt * distanceToCollisionZ / dPos.Z()
	} else if dPos.Z() < 0 {
		distanceToCollisionZ = static.front() - dynamic.back()
		timeOfZAxisIntersection = dt * distanceToCollisionZ / dPos.Z()
	}

	if timeOfXAxisIntersection >= timeOfYAxisIntersection && timeOfXAxisIntersection >= timeOfZAxisIntersection {
		dPos[0] = distanceToCollisionX // if dX >= 0 then left side else right side
	}

	if timeOfYAxisIntersection >= timeOfXAxisIntersection && timeOfYAxisIntersection >= timeOfZAxisIntersection {
		dPos[1] = distanceToCollisionY
	}

	if timeOfZAxisIntersection >= timeOfXAxisIntersection && timeOfZAxisIntersection >= timeOfYAxisIntersection {
		dPos[2] = distanceToCollisionZ
	}

	return dPos
}

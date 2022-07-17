package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
)

// MoveDirection specifies camera move direction.
type MoveDirection int

// Move constants.
const (
	MoveForward MoveDirection = iota
	MoveBack
	MoveLeft
	MoveRight
)

// RotateDirection specifies camera rotation direction.
type RotateDirection int

// Rotate constants.
const (
	RotateUp RotateDirection = iota
	RotateDown
	RotateLeft
	RotateRight
)

// RollDirection specifies camera roll direction.
type RollDirection int

// Roll constants.
const (
	RollLeft RollDirection = iota
	RollRight
)

// PerspectiveCamera is a camera that uses perspective projection.
type PerspectiveCamera struct {
	eye        mgl32.Vec3
	target     mgl32.Vec3
	up         mgl32.Vec3
	fovRadians float32
	zoom       float32
	ratio      float32
}

// NewPerspectiveCamera creates a new camera.
func NewPerspectiveCamera(eye, target, up mgl32.Vec3, fov, zoom, ratio float32) *PerspectiveCamera {
	return &PerspectiveCamera{
		eye:        eye,
		target:     target,
		up:         up,
		fovRadians: fov,
		zoom:       zoom,
		ratio:      ratio,
	}
}

// Projection returns the projection matrix.
func (c *PerspectiveCamera) Projection() mgl32.Mat4 {
	return mgl32.Perspective(c.fovRadians*c.zoom, c.ratio, 1, 1000)
}

// View returns the view matrix.
func (c *PerspectiveCamera) View() mgl32.Mat4 {
	return mgl32.LookAtV(c.eye, c.target, c.up)
}

// Move the camera in the specified direction.
func (c *PerspectiveCamera) Move(direction MoveDirection, amount float32) {
	var dir mgl32.Vec3

	switch direction {
	case MoveForward:
		dir = c.DirAxis().Mul(amount)
	case MoveBack:
		dir = c.DirAxis().Mul(-amount)
	case MoveRight:
		dir = c.PitchAxis().Mul(amount)
	case MoveLeft:
		dir = c.PitchAxis().Mul(-amount)
	}

	c.target = c.target.Add(dir)
	c.eye = c.eye.Add(dir)
}

// Rotate the camera in the specified direction.
func (c *PerspectiveCamera) Rotate(direction RotateDirection, amount float32) {
	switch direction {
	case RotateUp:
		c.target = Point3D(c.target).RotateAroundPoint(c.eye, c.PitchAxis(), amount)

	case RotateDown:
		c.target = Point3D(c.target).RotateAroundPoint(c.eye, c.PitchAxis(), -amount)

	case RotateLeft:
		c.target = Point3D(c.target).RotateAroundPoint(c.eye, c.up, amount)

	case RotateRight:
		c.target = Point3D(c.target).RotateAroundPoint(c.eye, c.up, -amount)
	}
}

// Roll camera left or right.
func (c *PerspectiveCamera) Roll(direction RollDirection, amount float32) {
	switch direction {
	case RollLeft:
		c.up = Vector3D(c.up).RotateAroundAxis(c.DirAxis(), amount)
	case RollRight:
		c.up = Vector3D(c.up).RotateAroundAxis(c.DirAxis(), -amount)
	}
}

// DirAxis returns the direction axis.
func (c *PerspectiveCamera) DirAxis() mgl32.Vec3 {
	return c.target.Sub(c.eye).Normalize()
}

// PitchAxis returns the pitch axis.
func (c *PerspectiveCamera) PitchAxis() mgl32.Vec3 {
	return c.target.Sub(c.eye).Cross(c.up).Normalize()
}

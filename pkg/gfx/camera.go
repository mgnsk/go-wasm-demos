package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
)

type MoveDirection int

const (
	MoveForward MoveDirection = iota
	MoveBack
	MoveLeft
	MoveRight
)

type RotateDirection int

const (
	RotateUp RotateDirection = iota
	RotateDown
	RotateLeft
	RotateRight
)

type RollDirection int

const (
	RollLeft RollDirection = iota
	RollRight
)

type Camera struct {
	eye        mgl32.Vec3
	target     mgl32.Vec3
	up         mgl32.Vec3
	fovRadians float32
	zoom       float32
	ratio      float32
}

func NewCamera(eye, target, up mgl32.Vec3, fov, zoom, ratio float32) *Camera {
	return &Camera{
		eye:        eye,
		target:     target,
		up:         up,
		fovRadians: fov,
		zoom:       zoom,
		ratio:      ratio,
	}
}

func (c *Camera) Projection() mgl32.Mat4 {
	return mgl32.Perspective(c.fovRadians*c.zoom, c.ratio, 1, 1000)
}

func (c *Camera) View() mgl32.Mat4 {
	return mgl32.LookAtV(c.eye, c.target, c.up)
}

func (c *Camera) Move(direction MoveDirection) {
	switch direction {
	case MoveForward:
		c.target = c.target.Add(c.dirAxis().Mul(0.1))
		c.eye = c.eye.Add(c.dirAxis().Mul(0.1))
	case MoveBack:
		c.target = c.target.Add(c.dirAxis().Mul(-0.1))
		c.eye = c.eye.Add(c.dirAxis().Mul(-0.1))
	case MoveRight:
		c.target = c.target.Add(c.pitchAxis().Mul(0.1))
		c.eye = c.eye.Add(c.pitchAxis().Mul(0.1))
	case MoveLeft:
		c.target = c.target.Sub(c.pitchAxis().Mul(0.1))
		c.eye = c.eye.Sub(c.pitchAxis().Mul(0.1))
	}
}

func (c *Camera) dirAxis() mgl32.Vec3 {
	return c.target.Sub(c.eye).Normalize()
}

func (c *Camera) pitchAxis() mgl32.Vec3 {
	return c.dirAxis().Cross(c.up).Normalize()
}

// Rotate the camera in the specified direction.
func (c *Camera) Rotate(direction RotateDirection) {
	switch direction {
	case RotateUp:
		c.target = Point3D(c.target).RotateAroundPoint(c.eye, c.pitchAxis(), 0.01)

	case RotateDown:
		c.target = Point3D(c.target).RotateAroundPoint(c.eye, c.pitchAxis(), -0.01)

	case RotateLeft:
		c.target = Point3D(c.target).RotateAroundPoint(c.eye, c.up, 0.01)

	case RotateRight:
		c.target = Point3D(c.target).RotateAroundPoint(c.eye, c.up, -0.01)

	}
}

// Roll camera left or right.
func (c *Camera) Roll(direction RollDirection) {
	switch direction {
	case RollLeft:
		c.up = Vector3D(c.up).RotateAroundAxis(c.dirAxis(), 0.1)
	case RollRight:
		c.up = Vector3D(c.up).RotateAroundAxis(c.dirAxis(), -0.1)
	}
}

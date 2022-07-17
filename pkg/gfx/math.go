package gfx

import (
	"github.com/chewxy/math32"
	"github.com/go-gl/mathgl/mgl32"
)

// Point3D is a point in 3D space.
type Point3D mgl32.Vec3

// RotateAroundPoint rotates point p around middle point with direction vector
// dir (must be unit vector) by an angle.
// Formula from here: https://sites.google.com/site/glennmurray/Home/rotation-matrices-and-formulas
func (p Point3D) RotateAroundPoint(middle mgl32.Vec3, axis mgl32.Vec3, angle float32) mgl32.Vec3 {
	a := middle.X()
	b := middle.Y()
	c := middle.Z()
	u := axis.X()
	v := axis.Y()
	w := axis.Z()

	angleCos := math32.Cos(angle)
	angleSin := math32.Sin(angle)

	newX := (a*(v*v+w*w)-u*(b*v+c*w-u*p[0]-v*p[1]-w*p[2]))*(1-angleCos) + p[0]*angleCos + (-c*v+b*w-w*p[1]+v*p[2])*angleSin
	newY := (b*(u*u+w*w)-v*(a*u+c*w-u*p[0]-v*p[1]-w*p[2]))*(1-angleCos) + p[1]*angleCos + (c*u-a*w+w*p[0]-u*p[2])*angleSin
	newZ := (c*(u*u+v*v)-w*(a*u+b*v-u*p[0]-v*p[1]-w*p[2]))*(1-angleCos) + p[2]*angleCos + (-b*u+a*v-v*p[0]+u*p[1])*angleSin

	return mgl32.Vec3{newX, newY, newZ}
}

// Vector3D is a 3D vector.
type Vector3D mgl32.Vec3

// RotateAroundAxis rotates a vector around an axis by radians degrees
// using  Rodrigues' rotation formula: https://en.wikipedia.org/wiki/Rodrigues%27_rotation_formula
func (v Vector3D) RotateAroundAxis(axis mgl32.Vec3, angle float32) mgl32.Vec3 {
	add1 := mgl32.Vec3(v).Mul(math32.Cos(angle))
	add2 := mgl32.Vec3(v).Cross(axis).Mul(math32.Sin(angle))
	dot := mgl32.Vec3(v).Dot(axis)
	add3 := axis.Mul(dot * (1 - math32.Cos(angle)))
	return add1.Add(add2).Add(add3)
}

package render3d

import "math"

// Vec3 is a 3D vector
type Vec3 struct {
	X, Y, Z float64
}

func V3(x, y, z float64) Vec3 { return Vec3{x, y, z} }

func (v Vec3) Add(o Vec3) Vec3    { return Vec3{v.X + o.X, v.Y + o.Y, v.Z + o.Z} }
func (v Vec3) Sub(o Vec3) Vec3    { return Vec3{v.X - o.X, v.Y - o.Y, v.Z - o.Z} }
func (v Vec3) Scale(s float64) Vec3 { return Vec3{v.X * s, v.Y * s, v.Z * s} }
func (v Vec3) Dot(o Vec3) float64 { return v.X*o.X + v.Y*o.Y + v.Z*o.Z }
func (v Vec3) Cross(o Vec3) Vec3 {
	return Vec3{v.Y*o.Z - v.Z*o.Y, v.Z*o.X - v.X*o.Z, v.X*o.Y - v.Y*o.X}
}
func (v Vec3) Len() float64       { return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z) }
func (v Vec3) Normalize() Vec3 {
	l := v.Len()
	if l < 1e-10 {
		return Vec3{}
	}
	return Vec3{v.X / l, v.Y / l, v.Z / l}
}
func (v Vec3) Lerp(o Vec3, t float64) Vec3 {
	return Vec3{v.X + (o.X-v.X)*t, v.Y + (o.Y-v.Y)*t, v.Z + (o.Z-v.Z)*t}
}

// Vec4 for homogeneous coords
type Vec4 struct {
	X, Y, Z, W float64
}

// Mat4 is a 4x4 matrix (column-major)
type Mat4 [16]float64

func Mat4Identity() Mat4 {
	var m Mat4
	m[0], m[5], m[10], m[15] = 1, 1, 1, 1
	return m
}

func Mat4Translate(tx, ty, tz float64) Mat4 {
	m := Mat4Identity()
	m[12], m[13], m[14] = tx, ty, tz
	return m
}

func Mat4Scale(sx, sy, sz float64) Mat4 {
	m := Mat4Identity()
	m[0], m[5], m[10] = sx, sy, sz
	return m
}

func Mat4RotateX(angle float64) Mat4 {
	c, s := math.Cos(angle), math.Sin(angle)
	m := Mat4Identity()
	m[5], m[6] = c, s
	m[9], m[10] = -s, c
	return m
}

func Mat4RotateY(angle float64) Mat4 {
	c, s := math.Cos(angle), math.Sin(angle)
	m := Mat4Identity()
	m[0], m[2] = c, -s
	m[8], m[10] = s, c
	return m
}

func Mat4RotateZ(angle float64) Mat4 {
	c, s := math.Cos(angle), math.Sin(angle)
	m := Mat4Identity()
	m[0], m[1] = c, s
	m[4], m[5] = -s, c
	return m
}

func Mat4Ortho(left, right, bottom, top, near, far float64) Mat4 {
	var m Mat4
	m[0] = 2 / (right - left)
	m[5] = 2 / (top - bottom)
	m[10] = -2 / (far - near)
	m[12] = -(right + left) / (right - left)
	m[13] = -(top + bottom) / (top - bottom)
	m[14] = -(far + near) / (far - near)
	m[15] = 1
	return m
}

// Mul multiplies two matrices
func (a Mat4) Mul(b Mat4) Mat4 {
	var r Mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				r[j*4+i] += a[k*4+i] * b[j*4+k]
			}
		}
	}
	return r
}

// MulVec4 multiplies matrix by vec4
func (m Mat4) MulVec4(v Vec4) Vec4 {
	return Vec4{
		m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12]*v.W,
		m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13]*v.W,
		m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14]*v.W,
		m[3]*v.X + m[7]*v.Y + m[11]*v.Z + m[15]*v.W,
	}
}

// TransformPoint transforms a 3D point (w=1)
func (m Mat4) TransformPoint(v Vec3) Vec3 {
	r := m.MulVec4(Vec4{v.X, v.Y, v.Z, 1})
	if r.W != 0 {
		return Vec3{r.X / r.W, r.Y / r.W, r.Z / r.W}
	}
	return Vec3{r.X, r.Y, r.Z}
}

// TransformDir transforms a direction (w=0)
func (m Mat4) TransformDir(v Vec3) Vec3 {
	r := m.MulVec4(Vec4{v.X, v.Y, v.Z, 0})
	return Vec3{r.X, r.Y, r.Z}
}

// LookAt creates a view matrix
func Mat4LookAt(eye, center, up Vec3) Mat4 {
	f := center.Sub(eye).Normalize()
	s := f.Cross(up).Normalize()
	u := s.Cross(f)
	var m Mat4
	m[0], m[4], m[8] = s.X, s.Y, s.Z
	m[1], m[5], m[9] = u.X, u.Y, u.Z
	m[2], m[6], m[10] = -f.X, -f.Y, -f.Z
	m[12] = -s.Dot(eye)
	m[13] = -u.Dot(eye)
	m[14] = f.Dot(eye)
	m[15] = 1
	return m
}

// RGBA color
type Color3 struct {
	R, G, B float64
}

func (c Color3) Scale(s float64) Color3 {
	return Color3{c.R * s, c.G * s, c.B * s}
}

func (c Color3) Add(o Color3) Color3 {
	return Color3{
		math.Min(c.R+o.R, 1),
		math.Min(c.G+o.G, 1),
		math.Min(c.B+o.B, 1),
	}
}

func (c Color3) Mul(o Color3) Color3 {
	return Color3{c.R * o.R, c.G * o.G, c.B * o.B}
}

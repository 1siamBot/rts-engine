package render3d

import "math"

// Camera3D implements an isometric camera with orthographic projection
type Camera3D struct {
	// Camera target (world position to look at)
	TargetX, TargetY float64

	// Zoom: how many world units fit on screen
	Zoom float64

	// Screen dimensions
	ScreenW, ScreenH int

	// Isometric angles
	Pitch float64 // ~35.264° for true isometric
	Yaw   float64 // 45° for classic isometric

	// Computed matrices
	view       Mat4
	proj       Mat4
	viewProj   Mat4
	dirty      bool

	// Edge scrolling
	EdgeScroll bool
	EdgeSize   int
}

// NewCamera3D creates an isometric camera
func NewCamera3D(screenW, screenH int) *Camera3D {
	c := &Camera3D{
		Zoom:       20, // 20 world units visible
		ScreenW:    screenW,
		ScreenH:    screenH,
		Pitch:      35.264 * math.Pi / 180, // true isometric angle
		Yaw:        45 * math.Pi / 180,
		EdgeScroll: true,
		EdgeSize:   20,
		dirty:      true,
	}
	return c
}

// CenterOn centers camera on world position
func (c *Camera3D) CenterOn(wx, wy float64) {
	c.TargetX = wx
	c.TargetY = wy
	c.dirty = true
}

// Pan moves the camera in screen-relative direction
func (c *Camera3D) Pan(dx, dy float64) {
	// Convert screen delta to world delta based on yaw
	cosY := math.Cos(c.Yaw)
	sinY := math.Sin(c.Yaw)
	scale := c.Zoom / float64(c.ScreenW) // world units per pixel
	c.TargetX += (dx*cosY + dy*sinY) * scale
	c.TargetY += (-dx*sinY + dy*cosY) * scale
	c.dirty = true
}

// ZoomAt zooms toward a screen point
func (c *Camera3D) ZoomAt(delta float64, screenX, screenY int) {
	// Get world pos before zoom
	wx, wy := c.ScreenToWorld(screenX, screenY)
	c.Zoom *= 1 - delta*0.05
	if c.Zoom < 5 {
		c.Zoom = 5
	}
	if c.Zoom > 80 {
		c.Zoom = 80
	}
	c.dirty = true
	// Get world pos after zoom and adjust
	wx2, wy2 := c.ScreenToWorld(screenX, screenY)
	c.TargetX += wx - wx2
	c.TargetY += wy - wy2
	c.dirty = true
}

func (c *Camera3D) update() {
	if !c.dirty {
		return
	}
	c.dirty = false

	// Camera position: offset from target along isometric direction
	dist := 100.0 // arbitrary distance for ortho (doesn't affect size)
	eyeX := c.TargetX + dist*math.Sin(c.Yaw)*math.Cos(c.Pitch)
	eyeY := dist * math.Sin(c.Pitch)
	eyeZ := c.TargetY + dist*math.Cos(c.Yaw)*math.Cos(c.Pitch)

	eye := V3(eyeX, eyeY, eyeZ)
	center := V3(c.TargetX, 0, c.TargetY)
	up := V3(0, 1, 0)

	c.view = Mat4LookAt(eye, center, up)

	// Orthographic projection
	aspect := float64(c.ScreenW) / float64(c.ScreenH)
	halfW := c.Zoom / 2
	halfH := halfW / aspect
	c.proj = Mat4Ortho(-halfW, halfW, -halfH, halfH, 0.1, 500)

	c.viewProj = c.proj.Mul(c.view)
}

// ViewProj returns the combined view-projection matrix
func (c *Camera3D) ViewProj() Mat4 {
	c.update()
	return c.viewProj
}

// Project3DToScreen converts a 3D world point to screen coordinates
func (c *Camera3D) Project3DToScreen(wx, wy, wz float64) (int, int, float64) {
	c.update()
	// World space: X = east, Y = up, Z = south
	clip := c.viewProj.TransformPoint(V3(wx, wy, wz))
	// clip is in NDC [-1,1]
	sx := (clip.X*0.5 + 0.5) * float64(c.ScreenW)
	sy := (1 - (clip.Y*0.5 + 0.5)) * float64(c.ScreenH)
	return int(sx), int(sy), clip.Z
}

// WorldToScreen converts tile coords to screen (convenience, Y=0)
func (c *Camera3D) WorldToScreen(tileX, tileY float64) (int, int) {
	sx, sy, _ := c.Project3DToScreen(tileX, 0, tileY)
	return sx, sy
}

// ScreenToWorld converts screen coords to world XZ plane (Y=0)
func (c *Camera3D) ScreenToWorld(sx, sy int) (float64, float64) {
	c.update()
	// NDC
	ndcX := (float64(sx)/float64(c.ScreenW))*2 - 1
	ndcY := (1 - float64(sy)/float64(c.ScreenH))*2 - 1

	// Inverse view-proj to get ray
	// For ortho, we can compute directly
	// Unproject two points on near/far plane
	invVP := c.invertViewProj()
	near := invVP.TransformPoint(V3(ndcX, ndcY, -1))
	far := invVP.TransformPoint(V3(ndcX, ndcY, 1))

	// Intersect with Y=0 plane
	dir := far.Sub(near)
	if math.Abs(dir.Y) < 1e-10 {
		return near.X, near.Z
	}
	t := -near.Y / dir.Y
	return near.X + dir.X*t, near.Z + dir.Z*t
}

func (c *Camera3D) invertViewProj() Mat4 {
	return invertMat4(c.viewProj)
}

// VisibleTileRange returns approximate tile range visible on screen
func (c *Camera3D) VisibleTileRange(mapW, mapH int) (minX, minY, maxX, maxY int) {
	// Sample screen corners
	corners := [][2]int{{0, 0}, {c.ScreenW, 0}, {0, c.ScreenH}, {c.ScreenW, c.ScreenH}}
	minXf, minYf := math.MaxFloat64, math.MaxFloat64
	maxXf, maxYf := -math.MaxFloat64, -math.MaxFloat64
	for _, co := range corners {
		wx, wy := c.ScreenToWorld(co[0], co[1])
		if wx < minXf {
			minXf = wx
		}
		if wx > maxXf {
			maxXf = wx
		}
		if wy < minYf {
			minYf = wy
		}
		if wy > maxYf {
			maxYf = wy
		}
	}
	pad := 3
	minX = int(math.Floor(minXf)) - pad
	minY = int(math.Floor(minYf)) - pad
	maxX = int(math.Ceil(maxXf)) + pad
	maxY = int(math.Ceil(maxYf)) + pad
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= mapW {
		maxX = mapW - 1
	}
	if maxY >= mapH {
		maxY = mapH - 1
	}
	return
}

// SetMapBounds is a compatibility stub
func (c *Camera3D) SetMapBounds(w, h, tw, th int) {
	// No-op for 3D camera
}

// Simple 4x4 matrix inversion (Cramer's rule)
func invertMat4(m Mat4) Mat4 {
	var inv Mat4
	inv[0] = m[5]*m[10]*m[15] - m[5]*m[11]*m[14] - m[9]*m[6]*m[15] + m[9]*m[7]*m[14] + m[13]*m[6]*m[11] - m[13]*m[7]*m[10]
	inv[4] = -m[4]*m[10]*m[15] + m[4]*m[11]*m[14] + m[8]*m[6]*m[15] - m[8]*m[7]*m[14] - m[12]*m[6]*m[11] + m[12]*m[7]*m[10]
	inv[8] = m[4]*m[9]*m[15] - m[4]*m[11]*m[13] - m[8]*m[5]*m[15] + m[8]*m[7]*m[13] + m[12]*m[5]*m[11] - m[12]*m[7]*m[9]
	inv[12] = -m[4]*m[9]*m[14] + m[4]*m[10]*m[13] + m[8]*m[5]*m[14] - m[8]*m[6]*m[13] - m[12]*m[5]*m[10] + m[12]*m[6]*m[9]
	inv[1] = -m[1]*m[10]*m[15] + m[1]*m[11]*m[14] + m[9]*m[2]*m[15] - m[9]*m[3]*m[14] - m[13]*m[2]*m[11] + m[13]*m[3]*m[10]
	inv[5] = m[0]*m[10]*m[15] - m[0]*m[11]*m[14] - m[8]*m[2]*m[15] + m[8]*m[3]*m[14] + m[12]*m[2]*m[11] - m[12]*m[3]*m[10]
	inv[9] = -m[0]*m[9]*m[15] + m[0]*m[11]*m[13] + m[8]*m[1]*m[15] - m[8]*m[3]*m[13] - m[12]*m[1]*m[11] + m[12]*m[3]*m[9]
	inv[13] = m[0]*m[9]*m[14] - m[0]*m[10]*m[13] - m[8]*m[1]*m[14] + m[8]*m[2]*m[13] + m[12]*m[1]*m[10] - m[12]*m[2]*m[9]
	inv[2] = m[1]*m[6]*m[15] - m[1]*m[7]*m[14] - m[5]*m[2]*m[15] + m[5]*m[3]*m[14] + m[13]*m[2]*m[7] - m[13]*m[3]*m[6]
	inv[6] = -m[0]*m[6]*m[15] + m[0]*m[7]*m[14] + m[4]*m[2]*m[15] - m[4]*m[3]*m[14] - m[12]*m[2]*m[7] + m[12]*m[3]*m[6]
	inv[10] = m[0]*m[5]*m[15] - m[0]*m[7]*m[13] - m[4]*m[1]*m[15] + m[4]*m[3]*m[13] + m[12]*m[1]*m[7] - m[12]*m[3]*m[5]
	inv[14] = -m[0]*m[5]*m[14] + m[0]*m[6]*m[13] + m[4]*m[1]*m[14] - m[4]*m[2]*m[13] - m[12]*m[1]*m[6] + m[12]*m[2]*m[5]
	inv[3] = -m[1]*m[6]*m[11] + m[1]*m[7]*m[10] + m[5]*m[2]*m[11] - m[5]*m[3]*m[10] - m[9]*m[2]*m[7] + m[9]*m[3]*m[6]
	inv[7] = m[0]*m[6]*m[11] - m[0]*m[7]*m[10] - m[4]*m[2]*m[11] + m[4]*m[3]*m[10] + m[8]*m[2]*m[7] - m[8]*m[3]*m[6]
	inv[11] = -m[0]*m[5]*m[11] + m[0]*m[7]*m[9] + m[4]*m[1]*m[11] - m[4]*m[3]*m[9] - m[8]*m[1]*m[7] + m[8]*m[3]*m[5]
	inv[15] = m[0]*m[5]*m[10] - m[0]*m[6]*m[9] - m[4]*m[1]*m[10] + m[4]*m[2]*m[9] + m[8]*m[1]*m[6] - m[8]*m[2]*m[5]

	det := m[0]*inv[0] + m[1]*inv[4] + m[2]*inv[8] + m[3]*inv[12]
	if math.Abs(det) < 1e-10 {
		return Mat4Identity()
	}
	det = 1 / det
	for i := range inv {
		inv[i] *= det
	}
	return inv
}

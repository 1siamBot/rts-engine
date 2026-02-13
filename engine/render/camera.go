package render

import "math"

// Camera represents the viewport into the isometric world
type Camera struct {
	X, Y       float64 // camera center position (world coords)
	Zoom       float64 // zoom level (1.0 = default)
	MinZoom    float64
	MaxZoom    float64
	ScreenW    int // viewport width in pixels
	ScreenH    int // viewport height in pixels
	Speed      float64 // pan speed (pixels per second)
	EdgeScroll bool    // enable edge scrolling
	EdgeSize   int     // edge scroll trigger zone in pixels

	// Map bounds for clamping
	MapWidth   int
	MapHeight  int
	TileWidth  int
	TileHeight int
}

// NewCamera creates a camera with default settings
func NewCamera(screenW, screenH int) *Camera {
	return &Camera{
		X:         0,
		Y:         0,
		Zoom:      1.0,
		MinZoom:   0.25,
		MaxZoom:   3.0,
		ScreenW:   screenW,
		ScreenH:   screenH,
		Speed:     500,
		EdgeScroll: true,
		EdgeSize:  20,
		TileWidth: 64,
		TileHeight: 32,
	}
}

// SetMapBounds sets the map size for camera clamping
func (c *Camera) SetMapBounds(w, h, tw, th int) {
	c.MapWidth = w
	c.MapHeight = h
	c.TileWidth = tw
	c.TileHeight = th
}

// Pan moves the camera by pixel delta
func (c *Camera) Pan(dx, dy float64) {
	c.X += dx / c.Zoom
	c.Y += dy / c.Zoom
	c.clamp()
}

// SetZoom sets zoom level with clamping
func (c *Camera) SetZoom(z float64) {
	c.Zoom = math.Max(c.MinZoom, math.Min(c.MaxZoom, z))
}

// ZoomAt zooms toward a screen point
func (c *Camera) ZoomAt(delta float64, screenX, screenY int) {
	// Convert screen point to world before zoom
	wx, wy := c.ScreenToWorld(screenX, screenY)
	c.SetZoom(c.Zoom + delta)
	// Convert same screen point to world after zoom
	wx2, wy2 := c.ScreenToWorld(screenX, screenY)
	// Adjust camera to keep the point stationary
	c.X += wx - wx2
	c.Y += wy - wy2
	c.clamp()
}

// CenterOn centers the camera on a world position
func (c *Camera) CenterOn(wx, wy float64) {
	// Convert world to iso screen
	tw := float64(c.TileWidth)
	th := float64(c.TileHeight)
	c.X = (wx - wy) * (tw / 2)
	c.Y = (wx + wy) * (th / 2)
	c.clamp()
}

// WorldToScreen converts world iso position to screen pixel position
func (c *Camera) WorldToScreen(wx, wy float64) (int, int) {
	tw := float64(c.TileWidth)
	th := float64(c.TileHeight)
	// World to iso
	isoX := (wx - wy) * (tw / 2)
	isoY := (wx + wy) * (th / 2)
	// Apply camera offset and zoom
	sx := (isoX-c.X)*c.Zoom + float64(c.ScreenW)/2
	sy := (isoY-c.Y)*c.Zoom + float64(c.ScreenH)/2
	return int(sx), int(sy)
}

// ScreenToWorld converts screen pixel to world tile coords
func (c *Camera) ScreenToWorld(sx, sy int) (float64, float64) {
	tw := float64(c.TileWidth)
	th := float64(c.TileHeight)
	// Remove camera offset and zoom
	isoX := (float64(sx)-float64(c.ScreenW)/2)/c.Zoom + c.X
	isoY := (float64(sy)-float64(c.ScreenH)/2)/c.Zoom + c.Y
	// Iso to world
	wx := isoX/tw + isoY/th
	wy := isoY/th - isoX/tw
	return wx, wy
}

// VisibleTileRange returns the range of tiles visible on screen
func (c *Camera) VisibleTileRange(mapW, mapH int) (minX, minY, maxX, maxY int) {
	// Get world coords of screen corners
	wx0, wy0 := c.ScreenToWorld(0, 0)
	wx1, wy1 := c.ScreenToWorld(c.ScreenW, 0)
	wx2, wy2 := c.ScreenToWorld(0, c.ScreenH)
	wx3, wy3 := c.ScreenToWorld(c.ScreenW, c.ScreenH)

	// Find bounding box in tile space
	minXf := math.Min(math.Min(wx0, wx1), math.Min(wx2, wx3))
	minYf := math.Min(math.Min(wy0, wy1), math.Min(wy2, wy3))
	maxXf := math.Max(math.Max(wx0, wx1), math.Max(wx2, wx3))
	maxYf := math.Max(math.Max(wy0, wy1), math.Max(wy2, wy3))

	// Add padding and clamp
	pad := 2
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

func (c *Camera) clamp() {
	// Optional: clamp camera to map bounds
	// For now, allow free movement
}

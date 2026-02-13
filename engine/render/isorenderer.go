package render

import (
	"image/color"

	"github.com/1siamBot/rts-engine/engine/maplib"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// TerrainColors maps terrain types to colors (placeholder until real sprites)
var TerrainColors = map[maplib.TerrainType]color.RGBA{
	maplib.TerrainGrass:     {34, 139, 34, 255},     // forest green
	maplib.TerrainDirt:      {139, 119, 101, 255},    // brown
	maplib.TerrainSand:      {238, 214, 175, 255},    // sandy
	maplib.TerrainWater:     {30, 144, 255, 255},     // blue
	maplib.TerrainDeepWater: {0, 0, 139, 255},        // dark blue
	maplib.TerrainRock:      {128, 128, 128, 255},    // gray
	maplib.TerrainCliff:     {105, 105, 105, 255},    // dark gray
	maplib.TerrainRoad:      {169, 169, 169, 255},    // light gray
	maplib.TerrainBridge:    {139, 90, 43, 255},      // wood brown
	maplib.TerrainOre:       {255, 215, 0, 255},      // gold
	maplib.TerrainGem:       {0, 255, 255, 255},      // cyan
	maplib.TerrainSnow:      {245, 245, 255, 255},    // white
	maplib.TerrainUrban:     {192, 192, 192, 255},    // silver
	maplib.TerrainForest:    {0, 100, 0, 255},        // dark green
}

// IsoRenderer handles isometric map rendering
type IsoRenderer struct {
	Camera    *Camera
	TileCache map[maplib.TerrainType]*ebiten.Image
	Sprites   *SpriteManager
}

// NewIsoRenderer creates a new isometric renderer
func NewIsoRenderer(screenW, screenH int) *IsoRenderer {
	cam := NewCamera(screenW, screenH)
	r := &IsoRenderer{
		Camera:    cam,
		TileCache: make(map[maplib.TerrainType]*ebiten.Image),
		Sprites:   NewSpriteManager(),
	}
	return r
}

// GetTileImage returns (or creates) a cached tile image for a terrain type (default variant)
func (r *IsoRenderer) GetTileImage(terrain maplib.TerrainType, tw, th int) *ebiten.Image {
	if img, ok := r.TileCache[terrain]; ok {
		return img
	}

	// Try to use HD sprite, scaled to tile size
	if spriteImg, ok := r.Sprites.TerrainDefault[terrain]; ok {
		// Scale sprite to match tile dimensions
		sw := spriteImg.Bounds().Dx()
		sh := spriteImg.Bounds().Dy()
		if sw == tw && sh == th {
			r.TileCache[terrain] = spriteImg
			return spriteImg
		}
		// Scale
		scaled := ebiten.NewImage(tw, th)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(tw)/float64(sw), float64(th)/float64(sh))
		scaled.DrawImage(spriteImg, op)
		r.TileCache[terrain] = scaled
		return scaled
	}

	// Fallback: colored diamond
	img := ebiten.NewImage(tw, th)
	clr, ok := TerrainColors[terrain]
	if !ok {
		clr = color.RGBA{255, 0, 255, 255}
	}

	hw := float32(tw) / 2
	hh := float32(th) / 2

	var path vector.Path
	path.MoveTo(hw, 0)
	path.LineTo(float32(tw), hh)
	path.LineTo(hw, float32(th))
	path.LineTo(0, hh)
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(clr.R) / 255
		vs[i].ColorG = float32(clr.G) / 255
		vs[i].ColorB = float32(clr.B) / 255
		vs[i].ColorA = float32(clr.A) / 255
	}

	whiteImg := ebiten.NewImage(3, 3)
	whiteImg.Fill(color.White)
	img.DrawTriangles(vs, is, whiteImg, nil)

	vector.StrokeLine(img, hw, 0, float32(tw), hh, 1, color.RGBA{0, 0, 0, 80}, false)
	vector.StrokeLine(img, float32(tw), hh, hw, float32(th), 1, color.RGBA{0, 0, 0, 80}, false)
	vector.StrokeLine(img, hw, float32(th), 0, hh, 1, color.RGBA{0, 0, 0, 80}, false)
	vector.StrokeLine(img, 0, hh, hw, 0, 1, color.RGBA{0, 0, 0, 80}, false)

	r.TileCache[terrain] = img
	return img
}

// DrawMap renders the visible portion of the tile map
func (r *IsoRenderer) DrawMap(screen *ebiten.Image, tm *maplib.TileMap) {
	tw := tm.TileWidth
	th := tm.TileHeight

	r.Camera.SetMapBounds(tm.Width, tm.Height, tw, th)

	minX, minY, maxX, maxY := r.Camera.VisibleTileRange(tm.Width, tm.Height)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			tile := tm.At(x, y)
			if tile == nil {
				continue
			}

			// Get screen position
			sx, sy := r.Camera.WorldToScreen(float64(x), float64(y))

			// Adjust for tile height (elevation)
			sy -= int(tile.Height) * (th / 4)

			// Center the tile on the grid point
			sx -= tw / 2
			// sy is already at top of diamond due to iso math

			// Use variant sprite based on tile position (deterministic)
			var tileImg *ebiten.Image
			variantImg := r.Sprites.GetTerrainVariant(tile.Terrain, x, y)
			if variantImg != nil {
				sw := variantImg.Bounds().Dx()
				sh := variantImg.Bounds().Dy()
				if sw == tw && sh == th {
					tileImg = variantImg
				} else {
					tileImg = r.GetTileImage(tile.Terrain, tw, th)
				}
			} else {
				tileImg = r.GetTileImage(tile.Terrain, tw, th)
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(sx), float64(sy))

			// Apply zoom
			// (zoom is already factored into WorldToScreen)

			screen.DrawImage(tileImg, op)
		}
	}
}

// DrawGrid draws the isometric grid overlay
func (r *IsoRenderer) DrawGrid(screen *ebiten.Image, tm *maplib.TileMap) {
	tw := tm.TileWidth
	th := tm.TileHeight
	minX, minY, maxX, maxY := r.Camera.VisibleTileRange(tm.Width, tm.Height)

	gridColor := color.RGBA{255, 255, 255, 30}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			sx, sy := r.Camera.WorldToScreen(float64(x), float64(y))
			hw := float32(tw) / 2
			hh := float32(th) / 2
			cx := float32(sx)
			cy := float32(sy) + hh

			vector.StrokeLine(screen, cx, cy-hh, cx+hw, cy, 1, gridColor, false)
			vector.StrokeLine(screen, cx+hw, cy, cx, cy+hh, 1, gridColor, false)
		}
	}
}

// DrawSelectionBox draws a selection rectangle on screen
func (r *IsoRenderer) DrawSelectionBox(screen *ebiten.Image, x1, y1, x2, y2 int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	selColor := color.RGBA{0, 255, 0, 128}
	fillColor := color.RGBA{0, 255, 0, 30}

	// Fill
	fillImg := ebiten.NewImage(x2-x1, y2-y1)
	fillImg.Fill(fillColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x1), float64(y1))
	screen.DrawImage(fillImg, op)

	// Border
	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y1), 1, selColor, false)
	vector.StrokeLine(screen, float32(x2), float32(y1), float32(x2), float32(y2), 1, selColor, false)
	vector.StrokeLine(screen, float32(x2), float32(y2), float32(x1), float32(y2), 1, selColor, false)
	vector.StrokeLine(screen, float32(x1), float32(y2), float32(x1), float32(y1), 1, selColor, false)
}

// DrawMinimap draws a minimap in the corner
func (r *IsoRenderer) DrawMinimap(screen *ebiten.Image, tm *maplib.TileMap, posX, posY, size int) {
	minimap := ebiten.NewImage(size, size)
	minimap.Fill(color.RGBA{0, 0, 0, 180})

	scaleX := float64(size) / float64(tm.Width)
	scaleY := float64(size) / float64(tm.Height)

	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			tile := tm.At(x, y)
			if tile == nil {
				continue
			}
			clr, ok := TerrainColors[tile.Terrain]
			if !ok {
				clr = color.RGBA{128, 128, 128, 255}
			}

			px := float32(float64(x) * scaleX)
			py := float32(float64(y) * scaleY)
			pw := float32(scaleX) + 1
			ph := float32(scaleY) + 1

			vector.DrawFilledRect(minimap, px, py, pw, ph, clr, false)
		}
	}

	// Draw camera viewport indicator
	wx0, wy0 := r.Camera.ScreenToWorld(0, 0)
	wx1, wy1 := r.Camera.ScreenToWorld(r.Camera.ScreenW, r.Camera.ScreenH)

	vx0 := float32(wx0 * scaleX)
	vy0 := float32(wy0 * scaleY)
	vx1 := float32(wx1 * scaleX)
	vy1 := float32(wy1 * scaleY)

	viewColor := color.RGBA{255, 255, 255, 200}
	vector.StrokeLine(minimap, vx0, vy0, vx1, vy0, 1, viewColor, false)
	vector.StrokeLine(minimap, vx1, vy0, vx1, vy1, 1, viewColor, false)
	vector.StrokeLine(minimap, vx1, vy1, vx0, vy1, 1, viewColor, false)
	vector.StrokeLine(minimap, vx0, vy1, vx0, vy0, 1, viewColor, false)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(posX), float64(posY))
	screen.DrawImage(minimap, op)
}

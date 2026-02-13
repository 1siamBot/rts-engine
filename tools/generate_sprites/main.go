package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
)

func main() {
	base := filepath.Join("assets")
	tilesDir := filepath.Join(base, "tiles")
	spritesDir := filepath.Join(base, "sprites")
	os.MkdirAll(tilesDir, 0755)
	os.MkdirAll(spritesDir, 0755)

	// HD tile dimensions
	tw, th := 128, 64

	// Terrain tiles
	generateTile(tilesDir, "grass", tw, th, grassTile)
	generateTile(tilesDir, "dirt", tw, th, dirtTile)
	generateTile(tilesDir, "sand", tw, th, sandTile)
	generateTile(tilesDir, "water", tw, th, waterTile)
	generateTile(tilesDir, "deep_water", tw, th, deepWaterTile)
	generateTile(tilesDir, "rock", tw, th, rockTile)
	generateTile(tilesDir, "cliff", tw, th, cliffTile)
	generateTile(tilesDir, "road", tw, th, roadTile)
	generateTile(tilesDir, "bridge", tw, th, bridgeTile)
	generateTile(tilesDir, "ore", tw, th, oreTile)
	generateTile(tilesDir, "gem", tw, th, gemTile)
	generateTile(tilesDir, "snow", tw, th, snowTile)
	generateTile(tilesDir, "urban", tw, th, urbanTile)
	generateTile(tilesDir, "forest", tw, th, forestTile)

	// Building sprites (drawn on larger canvases)
	generateSprite(spritesDir, "construction_yard", 128, 96, constructionYardSprite)
	generateSprite(spritesDir, "power_plant", 96, 80, powerPlantSprite)
	generateSprite(spritesDir, "barracks", 96, 80, barracksSprite)
	generateSprite(spritesDir, "war_factory", 128, 96, warFactorySprite)
	generateSprite(spritesDir, "refinery", 128, 96, refinerySprite)

	// Unit sprites
	generateSprite(spritesDir, "infantry", 32, 32, infantrySprite)
	generateSprite(spritesDir, "tank", 48, 48, tankSprite)
	generateSprite(spritesDir, "harvester", 48, 48, harvesterSprite)
	generateSprite(spritesDir, "mcv", 56, 56, mcvSprite)

	fmt.Println("✅ All sprites generated in assets/")
}

func generateTile(dir, name string, w, h int, fn func(*image.RGBA, int, int)) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fn(img, w, h)
	savePNG(filepath.Join(dir, name+".png"), img)
}

func generateSprite(dir, name string, w, h int, fn func(*image.RGBA, int, int)) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fn(img, w, h)
	savePNG(filepath.Join(dir, name+".png"), img)
}

func savePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, img)
	fmt.Println("  →", path)
}

// isInDiamond checks if (px,py) is inside the isometric diamond of size (w,h)
func isInDiamond(px, py, w, h int) bool {
	cx := float64(w) / 2
	cy := float64(h) / 2
	dx := math.Abs(float64(px) - cx)
	dy := math.Abs(float64(py) - cy)
	return (dx/cx + dy/cy) <= 1.0
}

// diamondEdgeDist returns distance from edge (0=edge, 1=center)
func diamondEdgeDist(px, py, w, h int) float64 {
	cx := float64(w) / 2
	cy := float64(h) / 2
	dx := math.Abs(float64(px) - cx)
	dy := math.Abs(float64(py) - cy)
	d := dx/cx + dy/cy
	if d > 1 {
		return 0
	}
	return 1.0 - d
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	t = math.Max(0, math.Min(1, t))
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: uint8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

func noise(x, y int) float64 {
	return rand.Float64()*0.3 - 0.15
}

func perlinish(x, y, scale int) float64 {
	fx := float64(x) / float64(scale)
	fy := float64(y) / float64(scale)
	return (math.Sin(fx*2.7+fy*1.3) + math.Sin(fx*1.1-fy*3.2) + math.Sin(fx*4.5+fy*0.7)) / 6.0
}

func setPixelBlend(img *image.RGBA, x, y int, c color.RGBA) {
	if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
		return
	}
	existing := img.RGBAAt(x, y)
	if existing.A == 0 {
		img.SetRGBA(x, y, c)
		return
	}
	alpha := float64(c.A) / 255.0
	img.SetRGBA(x, y, color.RGBA{
		R: uint8(float64(existing.R)*(1-alpha) + float64(c.R)*alpha),
		G: uint8(float64(existing.G)*(1-alpha) + float64(c.G)*alpha),
		B: uint8(float64(existing.B)*(1-alpha) + float64(c.B)*alpha),
		A: 255,
	})
}

func fillDiamond(img *image.RGBA, w, h int, baseColor color.RGBA, noiseAmt float64) {
	rng := rand.New(rand.NewSource(int64(baseColor.R)*1000 + int64(baseColor.G)))
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			dist := diamondEdgeDist(px, py, w, h)
			n := (rng.Float64()*2 - 1) * noiseAmt
			// Gradient: slightly darker at edges, lighter toward center-top
			gradientT := float64(py) / float64(h)
			bright := 1.0 + (0.15 - gradientT*0.3) + n
			c := color.RGBA{
				R: clampU8(float64(baseColor.R) * bright),
				G: clampU8(float64(baseColor.G) * bright),
				B: clampU8(float64(baseColor.B) * bright),
				A: 255,
			}
			// Anti-alias edges
			if dist < 0.03 {
				c.A = uint8(dist / 0.03 * 255)
			}
			img.SetRGBA(px, py, c)
		}
	}
	// Draw subtle outline
	drawDiamondOutline(img, w, h, color.RGBA{0, 0, 0, 50})
}

func drawDiamondOutline(img *image.RGBA, w, h int, c color.RGBA) {
	hw := float64(w) / 2
	hh := float64(h) / 2
	// Top to right
	drawLineAA(img, int(hw), 0, w-1, int(hh), c)
	// Right to bottom
	drawLineAA(img, w-1, int(hh), int(hw), h-1, c)
	// Bottom to left
	drawLineAA(img, int(hw), h-1, 0, int(hh), c)
	// Left to top
	drawLineAA(img, 0, int(hh), int(hw), 0, c)
}

func drawLineAA(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := math.Abs(float64(x1 - x0))
	dy := math.Abs(float64(y1 - y0))
	steps := int(math.Max(dx, dy))
	if steps == 0 {
		return
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := float64(x0) + t*float64(x1-x0)
		y := float64(y0) + t*float64(y1-y0)
		setPixelBlend(img, int(x), int(y), c)
	}
}

func clampU8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func fillRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			setPixelBlend(img, px, py, c)
		}
	}
}

func fillCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for py := cy - r; py <= cy+r; py++ {
		for px := cx - r; px <= cx+r; px++ {
			dx := float64(px - cx)
			dy := float64(py - cy)
			d := math.Sqrt(dx*dx + dy*dy)
			if d <= float64(r) {
				alpha := c.A
				if d > float64(r)-1 {
					alpha = uint8(float64(alpha) * (float64(r) - d))
				}
				setPixelBlend(img, px, py, color.RGBA{c.R, c.G, c.B, alpha})
			}
		}
	}
}

func fillEllipse(img *image.RGBA, cx, cy, rx, ry int, c color.RGBA) {
	for py := cy - ry; py <= cy+ry; py++ {
		for px := cx - rx; px <= cx+rx; px++ {
			dx := float64(px-cx) / float64(rx)
			dy := float64(py-cy) / float64(ry)
			d := dx*dx + dy*dy
			if d <= 1.0 {
				alpha := c.A
				if d > 0.85 {
					alpha = uint8(float64(alpha) * (1.0 - d) / 0.15)
				}
				setPixelBlend(img, px, py, color.RGBA{c.R, c.G, c.B, alpha})
			}
		}
	}
}

// fillTriangle fills a triangle
func fillTriangle(img *image.RGBA, x0, y0, x1, y1, x2, y2 int, c color.RGBA) {
	minX := min3(x0, x1, x2)
	maxX := max3(x0, x1, x2)
	minY := min3(y0, y1, y2)
	maxY := max3(y0, y1, y2)
	for py := minY; py <= maxY; py++ {
		for px := minX; px <= maxX; px++ {
			if pointInTriangle(px, py, x0, y0, x1, y1, x2, y2) {
				setPixelBlend(img, px, py, c)
			}
		}
	}
}

func pointInTriangle(px, py, x0, y0, x1, y1, x2, y2 int) bool {
	d1 := sign(px, py, x0, y0, x1, y1)
	d2 := sign(px, py, x1, y1, x2, y2)
	d3 := sign(px, py, x2, y2, x0, y0)
	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !(hasNeg && hasPos)
}

func sign(px, py, x0, y0, x1, y1 int) float64 {
	return float64((px-x1)*(y0-y1) - (x0-x1)*(py-y1))
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

func max3(a, b, c int) int {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}

// ============= TERRAIN TILES =============

func grassTile(img *image.RGBA, w, h int) {
	base := color.RGBA{45, 160, 45, 255}
	fillDiamond(img, w, h, base, 0.12)
	// Add grass blades
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 30; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) && diamondEdgeDist(px, py, w, h) > 0.1 {
			bright := uint8(30 + rng.Intn(40))
			c := color.RGBA{bright, 120 + uint8(rng.Intn(60)), bright, 120}
			setPixelBlend(img, px, py, c)
			setPixelBlend(img, px, py-1, c)
		}
	}
}

func dirtTile(img *image.RGBA, w, h int) {
	base := color.RGBA{140, 110, 70, 255}
	fillDiamond(img, w, h, base, 0.15)
	rng := rand.New(rand.NewSource(43))
	for i := 0; i < 40; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) {
			c := color.RGBA{100 + uint8(rng.Intn(60)), 80 + uint8(rng.Intn(40)), 40 + uint8(rng.Intn(30)), 80}
			setPixelBlend(img, px, py, c)
		}
	}
}

func sandTile(img *image.RGBA, w, h int) {
	base := color.RGBA{220, 195, 150, 255}
	fillDiamond(img, w, h, base, 0.08)
	rng := rand.New(rand.NewSource(44))
	for i := 0; i < 25; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) {
			c := color.RGBA{200 + uint8(rng.Intn(40)), 180 + uint8(rng.Intn(30)), 130 + uint8(rng.Intn(30)), 60}
			setPixelBlend(img, px, py, c)
		}
	}
}

func waterTile(img *image.RGBA, w, h int) {
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			dist := diamondEdgeDist(px, py, w, h)
			wave := math.Sin(float64(px)*0.15+float64(py)*0.1) * 15
			r := clampU8(30 + wave)
			g := clampU8(120 + wave*0.5 + dist*30)
			b := clampU8(210 + wave*0.3)
			a := uint8(255)
			if dist < 0.03 {
				a = uint8(dist / 0.03 * 255)
			}
			img.SetRGBA(px, py, color.RGBA{r, g, b, a})
		}
	}
	// Specular highlights
	rng := rand.New(rand.NewSource(45))
	for i := 0; i < 8; i++ {
		px := w/4 + rng.Intn(w/2)
		py := h/4 + rng.Intn(h/2)
		if isInDiamond(px, py, w, h) {
			setPixelBlend(img, px, py, color.RGBA{200, 230, 255, 80})
			setPixelBlend(img, px+1, py, color.RGBA{200, 230, 255, 50})
		}
	}
	drawDiamondOutline(img, w, h, color.RGBA{0, 40, 80, 60})
}

func deepWaterTile(img *image.RGBA, w, h int) {
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			dist := diamondEdgeDist(px, py, w, h)
			wave := math.Sin(float64(px)*0.2+float64(py)*0.15) * 10
			r := clampU8(10 + wave*0.5)
			g := clampU8(40 + wave*0.3 + dist*20)
			b := clampU8(140 + wave)
			a := uint8(255)
			if dist < 0.03 {
				a = uint8(dist / 0.03 * 255)
			}
			img.SetRGBA(px, py, color.RGBA{r, g, b, a})
		}
	}
	drawDiamondOutline(img, w, h, color.RGBA{0, 10, 60, 60})
}

func rockTile(img *image.RGBA, w, h int) {
	base := color.RGBA{130, 130, 130, 255}
	fillDiamond(img, w, h, base, 0.2)
	// Rock cracks
	rng := rand.New(rand.NewSource(46))
	for i := 0; i < 5; i++ {
		x := w/4 + rng.Intn(w/2)
		y := h/4 + rng.Intn(h/2)
		for j := 0; j < 8; j++ {
			if isInDiamond(x, y, w, h) {
				setPixelBlend(img, x, y, color.RGBA{80, 80, 80, 120})
			}
			x += rng.Intn(3) - 1
			y += rng.Intn(3) - 1
		}
	}
}

func cliffTile(img *image.RGBA, w, h int) {
	// Top face lighter, bottom face darker for 3D effect
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			dist := diamondEdgeDist(px, py, w, h)
			t := float64(py) / float64(h)
			r := clampU8(100 - t*40)
			g := clampU8(95 - t*40)
			b := clampU8(90 - t*40)
			a := uint8(255)
			if dist < 0.03 {
				a = uint8(dist / 0.03 * 255)
			}
			img.SetRGBA(px, py, color.RGBA{r, g, b, a})
		}
	}
	drawDiamondOutline(img, w, h, color.RGBA{40, 40, 40, 100})
}

func roadTile(img *image.RGBA, w, h int) {
	base := color.RGBA{160, 155, 150, 255}
	fillDiamond(img, w, h, base, 0.05)
	// Lane markings (dashed center line)
	for py := 0; py < h; py++ {
		px := w / 2
		if isInDiamond(px, py, w, h) && (py/4)%2 == 0 {
			setPixelBlend(img, px, py, color.RGBA{230, 220, 100, 180})
			setPixelBlend(img, px-1, py, color.RGBA{230, 220, 100, 120})
		}
	}
}

func bridgeTile(img *image.RGBA, w, h int) {
	base := color.RGBA{150, 100, 55, 255}
	fillDiamond(img, w, h, base, 0.1)
	// Wood plank lines
	for i := 0; i < 6; i++ {
		y := h/8 + i*(h/7)
		for px := 0; px < w; px++ {
			if isInDiamond(px, y, w, h) {
				setPixelBlend(img, px, y, color.RGBA{100, 65, 30, 100})
			}
		}
	}
}

func oreTile(img *image.RGBA, w, h int) {
	base := color.RGBA{80, 70, 50, 255}
	fillDiamond(img, w, h, base, 0.15)
	// Golden ore sparkles
	rng := rand.New(rand.NewSource(47))
	for i := 0; i < 20; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) && diamondEdgeDist(px, py, w, h) > 0.1 {
			bright := uint8(180 + rng.Intn(75))
			fillCircle(img, px, py, 1+rng.Intn(2), color.RGBA{bright, bright - 30, 0, 200})
		}
	}
}

func gemTile(img *image.RGBA, w, h int) {
	base := color.RGBA{40, 60, 80, 255}
	fillDiamond(img, w, h, base, 0.1)
	rng := rand.New(rand.NewSource(48))
	for i := 0; i < 12; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) && diamondEdgeDist(px, py, w, h) > 0.1 {
			c := color.RGBA{0, uint8(200 + rng.Intn(55)), uint8(200 + rng.Intn(55)), 220}
			fillCircle(img, px, py, 2, c)
			setPixelBlend(img, px, py-1, color.RGBA{200, 255, 255, 120}) // highlight
		}
	}
}

func snowTile(img *image.RGBA, w, h int) {
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			dist := diamondEdgeDist(px, py, w, h)
			n := perlinish(px, py, 20) * 15
			r := clampU8(240 + n)
			g := clampU8(242 + n)
			b := clampU8(250 + n*0.5)
			a := uint8(255)
			if dist < 0.03 {
				a = uint8(dist / 0.03 * 255)
			}
			// Subtle blue shadows at bottom
			if float64(py) > float64(h)*0.6 {
				t := (float64(py)/float64(h) - 0.6) / 0.4
				b = clampU8(float64(b) - t*10)
				r = clampU8(float64(r) - t*15)
			}
			img.SetRGBA(px, py, color.RGBA{r, g, b, a})
		}
	}
	drawDiamondOutline(img, w, h, color.RGBA{180, 190, 210, 40})
}

func urbanTile(img *image.RGBA, w, h int) {
	base := color.RGBA{170, 170, 175, 255}
	fillDiamond(img, w, h, base, 0.06)
	// Grid pattern
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			if px%16 == 0 || py%8 == 0 {
				setPixelBlend(img, px, py, color.RGBA{140, 140, 145, 60})
			}
		}
	}
}

func forestTile(img *image.RGBA, w, h int) {
	// Green base
	base := color.RGBA{30, 100, 30, 255}
	fillDiamond(img, w, h, base, 0.15)
	// Tree tops (small circles)
	rng := rand.New(rand.NewSource(49))
	positions := [][2]int{}
	for i := 0; i < 8; i++ {
		px := w/4 + rng.Intn(w/2)
		py := h/4 + rng.Intn(h/2)
		if isInDiamond(px, py, w, h) && diamondEdgeDist(px, py, w, h) > 0.15 {
			positions = append(positions, [2]int{px, py})
		}
	}
	for _, p := range positions {
		// Shadow
		fillCircle(img, p[0]+2, p[1]+2, 6, color.RGBA{10, 40, 10, 100})
		// Tree canopy
		green := uint8(80 + rng.Intn(80))
		fillCircle(img, p[0], p[1], 5+rng.Intn(3), color.RGBA{20, green, 15, 220})
		// Highlight
		fillCircle(img, p[0]-1, p[1]-2, 2, color.RGBA{60, green + 40, 30, 100})
	}
}

// ============= BUILDING SPRITES =============

func constructionYardSprite(img *image.RGBA, w, h int) {
	// Shadow
	fillEllipse(img, w/2+3, h-12, w/2-8, 10, color.RGBA{0, 0, 0, 60})
	// Main structure - isometric box
	drawIsoBox(img, w/2, h/2+5, 50, 30, 25,
		color.RGBA{50, 130, 180, 255},  // top
		color.RGBA{35, 100, 150, 255},  // left
		color.RGBA{25, 80, 120, 255})   // right
	// Gear/star icon on top
	fillCircle(img, w/2, h/2-5, 8, color.RGBA{255, 220, 50, 230})
	fillCircle(img, w/2, h/2-5, 5, color.RGBA{255, 240, 100, 200})
	// Antenna
	drawLineAA(img, w/2+15, h/2-15, w/2+15, h/2-30, color.RGBA{180, 180, 180, 255})
	fillCircle(img, w/2+15, h/2-31, 2, color.RGBA{255, 50, 50, 255})
	// Border
	drawLineAA(img, w/2, h/2-20, w/2+50, h/2+5, color.RGBA{80, 180, 220, 120})
	drawLineAA(img, w/2, h/2-20, w/2-50, h/2+5, color.RGBA{80, 180, 220, 120})
}

func powerPlantSprite(img *image.RGBA, w, h int) {
	// Shadow
	fillEllipse(img, w/2+3, h-10, w/2-10, 8, color.RGBA{0, 0, 0, 60})
	// Main building
	drawIsoBox(img, w/2, h/2+8, 35, 25, 20,
		color.RGBA{180, 160, 60, 255},
		color.RGBA{150, 130, 40, 255},
		color.RGBA{120, 100, 30, 255})
	// Cooling tower (cylinder shape)
	fillEllipse(img, w/2-10, h/2-8, 10, 6, color.RGBA{200, 200, 200, 255})
	fillRect(img, w/2-20, h/2-18, 20, 12, color.RGBA{190, 190, 190, 255})
	fillEllipse(img, w/2-10, h/2-19, 10, 5, color.RGBA{210, 210, 215, 255})
	// Lightning bolt icon
	fillTriangle(img, w/2+5, h/2-5, w/2+12, h/2-5, w/2+8, h/2+3, color.RGBA{255, 230, 50, 230})
	fillTriangle(img, w/2+6, h/2+1, w/2+13, h/2+1, w/2+10, h/2+9, color.RGBA{255, 210, 30, 230})
}

func barracksSprite(img *image.RGBA, w, h int) {
	fillEllipse(img, w/2+3, h-10, w/2-10, 8, color.RGBA{0, 0, 0, 60})
	drawIsoBox(img, w/2, h/2+8, 35, 25, 18,
		color.RGBA{60, 140, 60, 255},
		color.RGBA{40, 110, 40, 255},
		color.RGBA{30, 85, 30, 255})
	// Door
	fillRect(img, w/2-4, h/2+5, 8, 12, color.RGBA{30, 60, 30, 255})
	// Flag pole
	drawLineAA(img, w/2+15, h/2+8, w/2+15, h/2-18, color.RGBA{160, 160, 160, 255})
	// Flag
	fillTriangle(img, w/2+15, h/2-18, w/2+25, h/2-14, w/2+15, h/2-10, color.RGBA{220, 40, 40, 230})
	// Star on building
	fillCircle(img, w/2, h/2-2, 4, color.RGBA{255, 255, 200, 180})
}

func warFactorySprite(img *image.RGBA, w, h int) {
	fillEllipse(img, w/2+3, h-12, w/2-8, 10, color.RGBA{0, 0, 0, 60})
	// Large industrial building
	drawIsoBox(img, w/2, h/2+5, 50, 30, 22,
		color.RGBA{120, 120, 130, 255},
		color.RGBA{90, 90, 100, 255},
		color.RGBA{70, 70, 80, 255})
	// Garage door
	fillRect(img, w/2-12, h/2+2, 24, 16, color.RGBA{40, 40, 50, 255})
	// Stripes on door
	for i := 0; i < 3; i++ {
		fillRect(img, w/2-10, h/2+4+i*5, 20, 2, color.RGBA{180, 160, 40, 200})
	}
	// Smokestack
	fillRect(img, w/2+18, h/2-20, 6, 15, color.RGBA{100, 100, 100, 255})
	fillCircle(img, w/2+21, h/2-22, 4, color.RGBA{80, 80, 80, 150})
}

func refinerySprite(img *image.RGBA, w, h int) {
	fillEllipse(img, w/2+3, h-12, w/2-8, 10, color.RGBA{0, 0, 0, 60})
	drawIsoBox(img, w/2, h/2+5, 45, 28, 20,
		color.RGBA{160, 140, 80, 255},
		color.RGBA{130, 110, 60, 255},
		color.RGBA{100, 85, 45, 255})
	// Conveyor belt visual
	for i := 0; i < 5; i++ {
		x := w/2 - 20 + i*10
		fillRect(img, x, h/2+10, 8, 3, color.RGBA{80, 80, 80, 200})
		fillCircle(img, x+4, h/2+11, 2, color.RGBA{100, 100, 100, 200})
	}
	// Ore symbol
	fillCircle(img, w/2, h/2-3, 6, color.RGBA{255, 200, 30, 220})
	fillCircle(img, w/2, h/2-3, 3, color.RGBA{255, 230, 100, 180})
	// Silo
	fillRect(img, w/2+16, h/2-15, 10, 18, color.RGBA{140, 120, 70, 255})
	fillEllipse(img, w/2+21, h/2-15, 5, 3, color.RGBA{160, 140, 80, 255})
}

func drawIsoBox(img *image.RGBA, cx, cy, halfW, halfH, height int,
	topColor, leftColor, rightColor color.RGBA) {
	// Top face (diamond)
	topY := cy - height
	for py := topY - halfH; py <= topY+halfH; py++ {
		for px := cx - halfW; px <= cx+halfW; px++ {
			dx := math.Abs(float64(px-cx)) / float64(halfW)
			dy := math.Abs(float64(py-topY)) / float64(halfH)
			if dx+dy <= 1.0 {
				setPixelBlend(img, px, py, topColor)
			}
		}
	}
	// Left face
	for py := topY; py <= cy; py++ {
		t := float64(py-topY) / float64(height)
		for px := cx - halfW; px <= cx; px++ {
			dxMax := float64(halfW) * (1.0 - float64(py-topY)/float64(halfH+height))
			if float64(px-cx+halfW) <= dxMax+float64(halfW) && px <= cx {
				darken := 1.0 - t*0.15
				c := color.RGBA{
					clampU8(float64(leftColor.R) * darken),
					clampU8(float64(leftColor.G) * darken),
					clampU8(float64(leftColor.B) * darken),
					leftColor.A,
				}
				setPixelBlend(img, px, py, c)
			}
		}
	}
	// Right face
	for py := topY; py <= cy; py++ {
		t := float64(py-topY) / float64(height)
		for px := cx; px <= cx+halfW; px++ {
			dxMax := float64(halfW) * (1.0 - float64(py-topY)/float64(halfH+height))
			if float64(px-cx) <= dxMax {
				darken := 1.0 - t*0.1
				c := color.RGBA{
					clampU8(float64(rightColor.R) * darken),
					clampU8(float64(rightColor.G) * darken),
					clampU8(float64(rightColor.B) * darken),
					rightColor.A,
				}
				setPixelBlend(img, px, py, c)
			}
		}
	}
}

// ============= UNIT SPRITES =============

func infantrySprite(img *image.RGBA, w, h int) {
	cx, cy := w/2, h/2
	// Shadow
	fillEllipse(img, cx+1, cy+10, 6, 3, color.RGBA{0, 0, 0, 60})
	// Body
	fillRect(img, cx-3, cy-2, 6, 10, color.RGBA{50, 100, 50, 255})
	// Head
	fillCircle(img, cx, cy-5, 4, color.RGBA{200, 170, 140, 255})
	// Helmet
	fillCircle(img, cx, cy-7, 4, color.RGBA{60, 90, 60, 250})
	// Weapon (rifle)
	drawLineAA(img, cx+3, cy, cx+8, cy-6, color.RGBA{80, 80, 80, 220})
	// Legs
	drawLineAA(img, cx-2, cy+8, cx-4, cy+12, color.RGBA{40, 80, 40, 255})
	drawLineAA(img, cx+2, cy+8, cx+4, cy+12, color.RGBA{40, 80, 40, 255})
	// Boots
	fillRect(img, cx-5, cy+11, 3, 2, color.RGBA{40, 35, 30, 255})
	fillRect(img, cx+3, cy+11, 3, 2, color.RGBA{40, 35, 30, 255})
}

func tankSprite(img *image.RGBA, w, h int) {
	cx, cy := w/2, h/2
	// Shadow
	fillEllipse(img, cx+2, cy+14, 18, 8, color.RGBA{0, 0, 0, 50})
	// Tracks
	fillEllipse(img, cx, cy+6, 20, 8, color.RGBA{50, 50, 45, 255})
	fillEllipse(img, cx, cy+6, 18, 6, color.RGBA{70, 70, 60, 255})
	// Track detail lines
	for i := -3; i <= 3; i++ {
		x := cx + i*5
		drawLineAA(img, x, cy+1, x, cy+11, color.RGBA{55, 55, 50, 200})
	}
	// Hull
	drawIsoBox(img, cx, cy+2, 14, 8, 6,
		color.RGBA{90, 110, 80, 255},
		color.RGBA{70, 85, 60, 255},
		color.RGBA{60, 75, 50, 255})
	// Turret
	fillEllipse(img, cx, cy-4, 8, 5, color.RGBA{80, 100, 70, 255})
	fillEllipse(img, cx, cy-5, 7, 4, color.RGBA{95, 115, 80, 255})
	// Gun barrel
	fillRect(img, cx-1, cy-14, 3, 10, color.RGBA{60, 60, 55, 255})
	// Muzzle
	fillRect(img, cx-2, cy-15, 5, 2, color.RGBA{70, 70, 65, 255})
	// Commander hatch
	fillCircle(img, cx+2, cy-4, 2, color.RGBA{75, 90, 65, 255})
}

func harvesterSprite(img *image.RGBA, w, h int) {
	cx, cy := w/2, h/2
	// Shadow
	fillEllipse(img, cx+2, cy+14, 18, 8, color.RGBA{0, 0, 0, 50})
	// Wheels/tracks
	fillEllipse(img, cx, cy+8, 18, 6, color.RGBA{50, 50, 45, 255})
	// Body
	drawIsoBox(img, cx, cy+3, 16, 10, 10,
		color.RGBA{200, 160, 40, 255},
		color.RGBA{170, 130, 30, 255},
		color.RGBA{140, 105, 25, 255})
	// Container/bucket (on top)
	fillRect(img, cx-10, cy-10, 20, 8, color.RGBA{180, 140, 30, 255})
	fillRect(img, cx-10, cy-10, 20, 2, color.RGBA{210, 170, 50, 255})
	// Scoop arm
	drawLineAA(img, cx-12, cy-6, cx-18, cy+2, color.RGBA{140, 140, 130, 255})
	drawLineAA(img, cx-18, cy+2, cx-14, cy+6, color.RGBA{140, 140, 130, 255})
	// Cab window
	fillRect(img, cx+4, cy-8, 6, 4, color.RGBA{140, 200, 220, 200})
}

func mcvSprite(img *image.RGBA, w, h int) {
	cx, cy := w/2, h/2
	// Shadow
	fillEllipse(img, cx+3, cy+18, 22, 10, color.RGBA{0, 0, 0, 50})
	// Large wheels
	fillEllipse(img, cx, cy+10, 22, 8, color.RGBA{45, 45, 40, 255})
	fillEllipse(img, cx, cy+10, 20, 6, color.RGBA{60, 60, 55, 255})
	// Body - large truck
	drawIsoBox(img, cx, cy+2, 20, 12, 14,
		color.RGBA{80, 60, 180, 255},
		color.RGBA{60, 45, 150, 255},
		color.RGBA{50, 35, 120, 255})
	// Deploy mechanism on top
	fillRect(img, cx-8, cy-16, 16, 6, color.RGBA{100, 80, 200, 255})
	fillRect(img, cx-6, cy-18, 12, 4, color.RGBA{120, 100, 220, 255})
	// Dish/antenna
	fillCircle(img, cx, cy-20, 4, color.RGBA{180, 180, 190, 255})
	fillCircle(img, cx, cy-20, 2, color.RGBA{140, 140, 160, 255})
	drawLineAA(img, cx, cy-16, cx, cy-20, color.RGBA{160, 160, 170, 255})
	// Cab
	fillRect(img, cx+8, cy-8, 8, 8, color.RGBA{70, 55, 160, 255})
	fillRect(img, cx+9, cy-7, 6, 4, color.RGBA{140, 180, 210, 200}) // window
	// MCV label star
	fillCircle(img, cx-4, cy-6, 3, color.RGBA{255, 220, 50, 200})
}

// Ensure draw is imported (used internally)
var _ = draw.Src

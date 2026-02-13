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

// Ensure draw is imported
var _ = draw.Src

func main() {
	base := filepath.Join("assets")
	tilesDir := filepath.Join(base, "tiles")
	spritesDir := filepath.Join(base, "sprites")
	effectsDir := filepath.Join(base, "effects")
	os.MkdirAll(tilesDir, 0755)
	os.MkdirAll(spritesDir, 0755)
	os.MkdirAll(effectsDir, 0755)

	tw, th := 128, 64

	// ========== TERRAIN TILES (3 variants each) ==========
	terrainGenerators := map[string]func(*image.RGBA, int, int, int){
		"grass":      grassTile,
		"dirt":       dirtTile,
		"sand":       sandTile,
		"water":      waterTile,
		"deep_water": deepWaterTile,
		"rock":       rockTile,
		"cliff":      cliffTile,
		"road":       roadTile,
		"bridge":     bridgeTile,
		"ore":        oreTile,
		"gem":        gemTile,
		"snow":       snowTile,
		"urban":      urbanTile,
		"forest":     forestTile,
		"concrete":   concreteTile,
	}

	for name, gen := range terrainGenerators {
		for v := 0; v < 3; v++ {
			img := image.NewRGBA(image.Rect(0, 0, tw, th))
			gen(img, tw, th, v)
			savePNG(filepath.Join(tilesDir, fmt.Sprintf("%s_%d.png", name, v)), img)
		}
		// Also save variant 0 as the default (backward compat)
		img := image.NewRGBA(image.Rect(0, 0, tw, th))
		gen(img, tw, th, 0)
		savePNG(filepath.Join(tilesDir, name+".png"), img)
	}

	// ========== BUILDING SPRITES ==========
	type buildingDef struct {
		name string
		w, h int
		fn   func(*image.RGBA, int, int, color.RGBA)
	}
	buildings := []buildingDef{
		{"construction_yard", 160, 120, constructionYardSprite},
		{"power_plant", 128, 100, powerPlantSprite},
		{"barracks", 128, 100, barracksSprite},
		{"war_factory", 160, 120, warFactorySprite},
		{"refinery", 160, 120, refinerySprite},
		{"radar", 128, 110, radarSprite},
		{"turret", 96, 80, turretSprite},
		{"wall", 64, 48, wallSprite},
	}

	factionColors := map[string]color.RGBA{
		"allied": {60, 120, 200, 255},
		"soviet": {200, 50, 50, 255},
	}

	for _, b := range buildings {
		for faction, fc := range factionColors {
			// Complete state
			img := image.NewRGBA(image.Rect(0, 0, b.w, b.h))
			b.fn(img, b.w, b.h, fc)
			savePNG(filepath.Join(spritesDir, fmt.Sprintf("%s_%s.png", b.name, faction)), img)

			// Construction frames (3 stages)
			for stage := 0; stage < 3; stage++ {
				img := image.NewRGBA(image.Rect(0, 0, b.w, b.h))
				drawConstructionStage(img, b.w, b.h, stage, fc)
				savePNG(filepath.Join(spritesDir, fmt.Sprintf("%s_%s_build_%d.png", b.name, faction, stage)), img)
			}

			// Damaged overlay
			img2 := image.NewRGBA(image.Rect(0, 0, b.w, b.h))
			b.fn(img2, b.w, b.h, fc)
			applyDamageOverlay(img2, b.w, b.h)
			savePNG(filepath.Join(spritesDir, fmt.Sprintf("%s_%s_damaged.png", b.name, faction)), img2)
		}
		// Also save default (allied) as plain name for backward compat
		img := image.NewRGBA(image.Rect(0, 0, b.w, b.h))
		b.fn(img, b.w, b.h, factionColors["allied"])
		savePNG(filepath.Join(spritesDir, b.name+".png"), img)
	}

	// ========== UNIT SPRITES (8 directions × 3 frames) ==========
	type unitDef struct {
		name string
		w, h int
		fn   func(*image.RGBA, int, int, int, int) // img, w, h, direction, frame
	}
	units := []unitDef{
		{"infantry", 40, 40, infantrySprite},
		{"tank", 56, 56, tankSprite},
		{"harvester", 56, 56, harvesterSprite},
		{"mcv", 64, 64, mcvSprite},
		{"engineer", 40, 40, engineerSprite},
		{"attack_dog", 40, 40, attackDogSprite},
		{"apocalypse_tank", 64, 64, apocalypseTankSprite},
		{"v3_rocket", 56, 56, v3RocketSprite},
	}

	for _, u := range units {
		for dir := 0; dir < 8; dir++ {
			for frame := 0; frame < 3; frame++ {
				img := image.NewRGBA(image.Rect(0, 0, u.w, u.h))
				u.fn(img, u.w, u.h, dir, frame)
				savePNG(filepath.Join(spritesDir, fmt.Sprintf("%s_d%d_f%d.png", u.name, dir, frame)), img)
			}
		}
		// Default sprite (dir=0, frame=0) for backward compat
		img := image.NewRGBA(image.Rect(0, 0, u.w, u.h))
		u.fn(img, u.w, u.h, 2, 0) // south-facing default
		savePNG(filepath.Join(spritesDir, u.name+".png"), img)
	}

	// ========== VISUAL EFFECTS ==========
	// Explosion (8 frames)
	for f := 0; f < 8; f++ {
		img := image.NewRGBA(image.Rect(0, 0, 64, 64))
		explosionFrame(img, 64, 64, f)
		savePNG(filepath.Join(effectsDir, fmt.Sprintf("explosion_%d.png", f)), img)
	}
	// Muzzle flash (3 frames)
	for f := 0; f < 3; f++ {
		img := image.NewRGBA(image.Rect(0, 0, 32, 32))
		muzzleFlashFrame(img, 32, 32, f)
		savePNG(filepath.Join(effectsDir, fmt.Sprintf("muzzle_%d.png", f)), img)
	}
	// Smoke puffs (4 frames)
	for f := 0; f < 4; f++ {
		img := image.NewRGBA(image.Rect(0, 0, 32, 32))
		smokeFrame(img, 32, 32, f)
		savePNG(filepath.Join(effectsDir, fmt.Sprintf("smoke_%d.png", f)), img)
	}
	// Ore sparkle (4 frames)
	for f := 0; f < 4; f++ {
		img := image.NewRGBA(image.Rect(0, 0, 16, 16))
		oreSparkleFrame(img, 16, 16, f)
		savePNG(filepath.Join(effectsDir, fmt.Sprintf("ore_sparkle_%d.png", f)), img)
	}
	// Selection circle
	img := image.NewRGBA(image.Rect(0, 0, 64, 32))
	selectionCircle(img, 64, 32)
	savePNG(filepath.Join(effectsDir, "selection_circle.png"), img)
	// Rally point flag
	img = image.NewRGBA(image.Rect(0, 0, 16, 32))
	rallyFlag(img, 16, 32)
	savePNG(filepath.Join(effectsDir, "rally_flag.png"), img)

	fmt.Println("✅ All HD sprites generated in assets/")
}

// ===================== UTILITY FUNCTIONS =====================

func savePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, img)
	fmt.Println("  →", path)
}

func isInDiamond(px, py, w, h int) bool {
	cx := float64(w) / 2
	cy := float64(h) / 2
	dx := math.Abs(float64(px) - cx)
	dy := math.Abs(float64(py) - cy)
	return (dx/cx + dy/cy) <= 1.0
}

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

func clampU8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	t = clampF(t, 0, 1)
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: uint8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

func setPixelBlend(img *image.RGBA, x, y int, c color.RGBA) {
	if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
		return
	}
	if c.A == 0 {
		return
	}
	existing := img.RGBAAt(x, y)
	if existing.A == 0 {
		img.SetRGBA(x, y, c)
		return
	}
	alpha := float64(c.A) / 255.0
	img.SetRGBA(x, y, color.RGBA{
		R: clampU8(float64(existing.R)*(1-alpha) + float64(c.R)*alpha),
		G: clampU8(float64(existing.G)*(1-alpha) + float64(c.G)*alpha),
		B: clampU8(float64(existing.B)*(1-alpha) + float64(c.B)*alpha),
		A: 255,
	})
}

func setPixelAdditive(img *image.RGBA, x, y int, c color.RGBA) {
	if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
		return
	}
	existing := img.RGBAAt(x, y)
	alpha := float64(c.A) / 255.0
	img.SetRGBA(x, y, color.RGBA{
		R: clampU8(float64(existing.R) + float64(c.R)*alpha),
		G: clampU8(float64(existing.G) + float64(c.G)*alpha),
		B: clampU8(float64(existing.B) + float64(c.B)*alpha),
		A: clampU8(float64(existing.A) + float64(c.A)*0.5),
	})
}

// Perlin-like noise for texturing
func perlinish(x, y, scale int) float64 {
	fx := float64(x) / float64(scale)
	fy := float64(y) / float64(scale)
	return (math.Sin(fx*2.7+fy*1.3) + math.Sin(fx*1.1-fy*3.2) + math.Sin(fx*4.5+fy*0.7)) / 6.0
}

func perlinish2(x, y, scale int, seed float64) float64 {
	fx := float64(x)/float64(scale) + seed
	fy := float64(y)/float64(scale) + seed*0.7
	return (math.Sin(fx*3.1+fy*1.7) + math.Sin(fx*0.8-fy*2.9) + math.Sin(fx*5.3+fy*0.4) + math.Cos(fx*1.4+fy*4.1)) / 8.0
}

// ===== Drawing primitives =====

func drawLineAA(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := math.Abs(float64(x1 - x0))
	dy := math.Abs(float64(y1 - y0))
	steps := int(math.Max(dx, dy))
	if steps == 0 {
		setPixelBlend(img, x0, y0, c)
		return
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := float64(x0) + t*float64(x1-x0)
		y := float64(y0) + t*float64(y1-y0)
		// Sub-pixel AA
		ix, iy := int(x), int(y)
		fx, fy := x-float64(ix), y-float64(iy)
		setPixelBlend(img, ix, iy, color.RGBA{c.R, c.G, c.B, clampU8(float64(c.A) * (1 - fx) * (1 - fy))})
		setPixelBlend(img, ix+1, iy, color.RGBA{c.R, c.G, c.B, clampU8(float64(c.A) * fx * (1 - fy))})
		setPixelBlend(img, ix, iy+1, color.RGBA{c.R, c.G, c.B, clampU8(float64(c.A) * (1 - fx) * fy)})
		setPixelBlend(img, ix+1, iy+1, color.RGBA{c.R, c.G, c.B, clampU8(float64(c.A) * fx * fy)})
	}
}

func drawThickLineAA(img *image.RGBA, x0, y0, x1, y1 int, thickness float64, c color.RGBA) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	length := math.Sqrt(dx*dx + dy*dy)
	if length == 0 {
		return
	}
	nx := -dy / length * thickness / 2
	ny := dx / length * thickness / 2
	for t := -thickness / 2; t <= thickness/2; t += 0.5 {
		ox := int(nx * t / (thickness / 2))
		oy := int(ny * t / (thickness / 2))
		drawLineAA(img, x0+ox, y0+oy, x1+ox, y1+oy, c)
	}
}

func fillRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			setPixelBlend(img, px, py, c)
		}
	}
}

func fillRectGradientV(img *image.RGBA, x, y, w, h int, top, bot color.RGBA) {
	for py := y; py < y+h; py++ {
		t := float64(py-y) / float64(h)
		c := lerpColor(top, bot, t)
		for px := x; px < x+w; px++ {
			setPixelBlend(img, px, py, c)
		}
	}
}

func fillCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for py := cy - r - 1; py <= cy+r+1; py++ {
		for px := cx - r - 1; px <= cx+r+1; px++ {
			dx := float64(px - cx)
			dy := float64(py - cy)
			d := math.Sqrt(dx*dx + dy*dy)
			if d <= float64(r)+0.5 {
				alpha := c.A
				if d > float64(r)-0.5 {
					alpha = clampU8(float64(c.A) * (float64(r) + 0.5 - d))
				}
				setPixelBlend(img, px, py, color.RGBA{c.R, c.G, c.B, alpha})
			}
		}
	}
}

func fillCircleGradient(img *image.RGBA, cx, cy, r int, center, edge color.RGBA) {
	for py := cy - r - 1; py <= cy+r+1; py++ {
		for px := cx - r - 1; px <= cx+r+1; px++ {
			dx := float64(px - cx)
			dy := float64(py - cy)
			d := math.Sqrt(dx*dx + dy*dy)
			if d <= float64(r)+0.5 {
				t := d / float64(r)
				c := lerpColor(center, edge, t)
				if d > float64(r)-0.5 {
					c.A = clampU8(float64(c.A) * (float64(r) + 0.5 - d))
				}
				setPixelBlend(img, px, py, c)
			}
		}
	}
}

func fillEllipse(img *image.RGBA, cx, cy, rx, ry int, c color.RGBA) {
	for py := cy - ry - 1; py <= cy+ry+1; py++ {
		for px := cx - rx - 1; px <= cx+rx+1; px++ {
			dx := float64(px-cx) / float64(rx)
			dy := float64(py-cy) / float64(ry)
			d := dx*dx + dy*dy
			if d <= 1.0 {
				alpha := c.A
				if d > 0.85 {
					alpha = clampU8(float64(c.A) * (1.0 - d) / 0.15)
				}
				setPixelBlend(img, px, py, color.RGBA{c.R, c.G, c.B, alpha})
			}
		}
	}
}

func fillEllipseGradient(img *image.RGBA, cx, cy, rx, ry int, center, edge color.RGBA) {
	for py := cy - ry - 1; py <= cy+ry+1; py++ {
		for px := cx - rx - 1; px <= cx+rx+1; px++ {
			dx := float64(px-cx) / float64(rx)
			dy := float64(py-cy) / float64(ry)
			d := dx*dx + dy*dy
			if d <= 1.0 {
				t := math.Sqrt(d)
				c := lerpColor(center, edge, t)
				if d > 0.85 {
					c.A = clampU8(float64(c.A) * (1.0 - d) / 0.15)
				}
				setPixelBlend(img, px, py, c)
			}
		}
	}
}

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
	d1 := triSign(px, py, x0, y0, x1, y1)
	d2 := triSign(px, py, x1, y1, x2, y2)
	d3 := triSign(px, py, x2, y2, x0, y0)
	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !(hasNeg && hasPos)
}

func triSign(px, py, x0, y0, x1, y1 int) float64 {
	return float64((px-x1)*(y0-y1) - (x0-x1)*(py-y1))
}

func min3(a, b, c int) int {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}

func max3(a, b, c int) int {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}

func drawDiamondOutline(img *image.RGBA, w, h int, c color.RGBA) {
	hw := float64(w) / 2
	hh := float64(h) / 2
	drawLineAA(img, int(hw), 0, w-1, int(hh), c)
	drawLineAA(img, w-1, int(hh), int(hw), h-1, c)
	drawLineAA(img, int(hw), h-1, 0, int(hh), c)
	drawLineAA(img, 0, int(hh), int(hw), 0, c)
}

// fillDiamondTextured fills a diamond with base color + noise + per-pixel lighting
func fillDiamondTextured(img *image.RGBA, w, h int, baseColor color.RGBA, noiseAmt float64, seed int64) {
	rng := rand.New(rand.NewSource(seed))
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			dist := diamondEdgeDist(px, py, w, h)
			n := (rng.Float64()*2 - 1) * noiseAmt
			// Top-left lighting: brighter top-left, darker bottom-right
			lightX := float64(px) / float64(w)
			lightY := float64(py) / float64(h)
			lighting := 1.0 + 0.12*(1.0-lightX) + 0.08*(1.0-lightY) - 0.15
			// Perlin-style large-scale variation
			p := perlinish(px, py, 20) * 0.1
			bright := lighting + n + p
			c := color.RGBA{
				R: clampU8(float64(baseColor.R) * bright),
				G: clampU8(float64(baseColor.G) * bright),
				B: clampU8(float64(baseColor.B) * bright),
				A: 255,
			}
			if dist < 0.04 {
				c.A = clampU8(dist / 0.04 * 255)
			}
			img.SetRGBA(px, py, c)
		}
	}
	drawDiamondOutline(img, w, h, color.RGBA{0, 0, 0, 40})
}

// drawIsoBox draws a full isometric box with proper geometry
func drawIsoBox(img *image.RGBA, cx, cy, halfW, halfH, height int,
	topColor, leftColor, rightColor color.RGBA) {

	topY := cy - height
	// Left face
	for py := topY; py <= cy; py++ {
		progress := float64(py-topY) / float64(cy-topY+1)
		leftEdge := cx - int(float64(halfW)*(1.0-progress*0.5))
		for px := leftEdge; px <= cx; px++ {
			darken := 1.0 - progress*0.2
			c := color.RGBA{
				clampU8(float64(leftColor.R) * darken),
				clampU8(float64(leftColor.G) * darken),
				clampU8(float64(leftColor.B) * darken),
				leftColor.A,
			}
			setPixelBlend(img, px, py, c)
		}
	}
	// Right face
	for py := topY; py <= cy; py++ {
		progress := float64(py-topY) / float64(cy-topY+1)
		rightEdge := cx + int(float64(halfW)*(1.0-progress*0.5))
		for px := cx; px <= rightEdge; px++ {
			darken := 1.0 - progress*0.15
			c := color.RGBA{
				clampU8(float64(rightColor.R) * darken),
				clampU8(float64(rightColor.G) * darken),
				clampU8(float64(rightColor.B) * darken),
				rightColor.A,
			}
			setPixelBlend(img, px, py, c)
		}
	}
	// Top face (diamond)
	for py := topY - halfH; py <= topY+halfH; py++ {
		for px := cx - halfW; px <= cx+halfW; px++ {
			dx := math.Abs(float64(px-cx)) / float64(halfW)
			dy := math.Abs(float64(py-topY)) / float64(halfH)
			if dx+dy <= 1.0 {
				// Add subtle lighting gradient on top face
				lightT := (float64(px-cx+halfW)/float64(2*halfW))*0.15 + (float64(py-topY+halfH)/float64(2*halfH))*0.1
				c := color.RGBA{
					clampU8(float64(topColor.R) * (1.05 - lightT)),
					clampU8(float64(topColor.G) * (1.05 - lightT)),
					clampU8(float64(topColor.B) * (1.05 - lightT)),
					topColor.A,
				}
				setPixelBlend(img, px, py, c)
			}
		}
	}
}

// drawIsoBoxDetailed draws box with panel lines and detail
func drawIsoBoxDetailed(img *image.RGBA, cx, cy, halfW, halfH, height int,
	topColor, leftColor, rightColor color.RGBA, panelLines bool) {
	drawIsoBox(img, cx, cy, halfW, halfH, height, topColor, leftColor, rightColor)
	if panelLines {
		topY := cy - height
		// Horizontal panel lines on left face
		for i := 1; i < 4; i++ {
			py := topY + i*(cy-topY)/4
			drawLineAA(img, cx-halfW/2, py, cx, py, color.RGBA{0, 0, 0, 30})
		}
		// Horizontal panel lines on right face
		for i := 1; i < 4; i++ {
			py := topY + i*(cy-topY)/4
			drawLineAA(img, cx, py, cx+halfW/2, py, color.RGBA{0, 0, 0, 25})
		}
	}
}

// rotatePoint rotates a 2D point around origin by angle (radians)
func rotatePoint(x, y, angle float64) (float64, float64) {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return x*cos - y*sin, x*sin + y*cos
}

// directionAngle returns angle in radians for 8-direction index
// 0=E, 1=SE, 2=S, 3=SW, 4=W, 5=NW, 6=N, 7=NE
func directionAngle(dir int) float64 {
	return float64(dir) * math.Pi / 4.0
}

// ===================== TERRAIN TILES =====================

func grassTile(img *image.RGBA, w, h, variant int) {
	seed := float64(variant)*100 + 42
	bases := []color.RGBA{{52, 120, 42, 255}, {48, 115, 46, 255}, {56, 125, 38, 255}}
	diamondReal(img, w, h, bases[variant%3], 0.3, seed)

	rng := rand.New(rand.NewSource(int64(seed)))
	// Realistic grass blades with color gradient
	for i := 0; i < 120; i++ {
		px := rng.Intn(w); py := rng.Intn(h)
		if !inDiamond(px, py, w, h) || diamondDist(px, py, w, h) < 0.08 { continue }
		green := uint8(70 + rng.Intn(80))
		darkGreen := uint8(30 + rng.Intn(30))
		length := 2 + rng.Intn(4)
		lean := (rng.Float64() - 0.5) * 2
		for j := 0; j < length; j++ {
			lx := px + int(lean*float64(j)/2)
			t := float64(j) / float64(length)
			gv := cu8(float64(darkGreen)*(1-t) + float64(green)*t)
			alpha := uint8(140 + rng.Intn(80))
			spxBlend(img, lx, py-j, color.RGBA{cu8(float64(gv) * 0.3), gv, cu8(float64(gv) * 0.2), alpha})
		}
	}
	// Wildflowers
	if variant > 0 {
		colors := []color.RGBA{{255, 240, 100, 180}, {255, 180, 200, 180}, {200, 170, 255, 180}}
		for i := 0; i < 2+variant*2; i++ {
			px := rng.Intn(w); py := rng.Intn(h)
			if inDiamond(px, py, w, h) && diamondDist(px, py, w, h) > 0.15 {
				fc := colors[rng.Intn(len(colors))]
				spxBlend(img, px, py, fc)
				spxBlend(img, px+1, py, color.RGBA{fc.R, fc.G, fc.B, fc.A / 2})
			}
		}
	}
	// Pebbles
	for i := 0; i < 3; i++ {
		px := rng.Intn(w); py := rng.Intn(h)
		if inDiamond(px, py, w, h) && diamondDist(px, py, w, h) > 0.1 {
			gray := uint8(100 + rng.Intn(50))
			fCircle(img, px, py, 1, color.RGBA{gray, gray - 5, gray - 10, 140})
		}
	}
}

func dirtTile(img *image.RGBA, w, h, variant int) {
	seed := float64(variant)*100 + 143
	bases := []color.RGBA{{110, 85, 55, 255}, {100, 78, 48, 255}, {115, 90, 60, 255}}
	diamondReal(img, w, h, bases[variant%3], 0.5, seed)

	rng := rand.New(rand.NewSource(int64(seed)))
	// Pebbles with 3D shading
	for i := 0; i < 30+variant*8; i++ {
		px := rng.Intn(w); py := rng.Intn(h)
		if !inDiamond(px, py, w, h) || diamondDist(px, py, w, h) < 0.06 { continue }
		gray := uint8(80 + rng.Intn(60))
		r := rng.Intn(2) + 1
		fCircle(img, px, py, r, color.RGBA{gray, gray - 8, gray - 15, uint8(140 + rng.Intn(80))})
		spxBlend(img, px-1, py-1, color.RGBA{gray + 30, gray + 25, gray + 20, 60})
	}
	// Cracks with depth
	for i := 0; i < 3+variant; i++ {
		x := w/4 + rng.Intn(w/2); y := h/4 + rng.Intn(h/2)
		for j := 0; j < 8+rng.Intn(10); j++ {
			if inDiamond(x, y, w, h) {
				spxBlend(img, x, y, color.RGBA{50, 35, 20, uint8(100 + rng.Intn(80))})
				spxBlend(img, x, y+1, color.RGBA{30, 20, 10, 40})
			}
			x += rng.Intn(3) - 1; y += rng.Intn(3) - 1
		}
	}
}

func sandTile(img *image.RGBA, w, h, variant int) {
	seeds := []int64{44, 3333, 5555}
	seed := seeds[variant%3]
	rng := rand.New(rand.NewSource(seed))
	bases := []color.RGBA{
		{225, 200, 155, 255},
		{215, 190, 145, 255},
		{230, 205, 160, 255},
	}
	fillDiamondTextured(img, w, h, bases[variant%3], 0.07, seed)

	// Dune ripple patterns
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			// Dune wave pattern
			wave := math.Sin(float64(px)*0.08+float64(py)*0.15+float64(variant)*2.0) * 0.5
			wave += math.Sin(float64(px)*0.03-float64(py)*0.06) * 0.3
			if wave > 0.3 {
				alpha := clampU8((wave - 0.3) * 200)
				setPixelBlend(img, px, py, color.RGBA{240, 220, 175, alpha})
			} else if wave < -0.3 {
				alpha := clampU8((-wave - 0.3) * 150)
				setPixelBlend(img, px, py, color.RGBA{190, 165, 120, alpha})
			}
		}
	}

	// Tiny dots of sand grain detail
	for i := 0; i < 40; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) {
			bright := uint8(200 + rng.Intn(55))
			setPixelBlend(img, px, py, color.RGBA{bright, bright - 20, bright - 50, 60})
		}
	}
}

func waterTile(img *image.RGBA, w, h, variant int) {
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			dist := diamondEdgeDist(px, py, w, h)
			phase := float64(variant) * 1.5
			// Multi-frequency waves
			wave1 := math.Sin(float64(px)*0.12+float64(py)*0.08+phase) * 18
			wave2 := math.Sin(float64(px)*0.06-float64(py)*0.15+phase*0.7) * 10
			wave3 := math.Sin(float64(px)*0.2+float64(py)*0.03+phase*1.3) * 5
			wave := wave1 + wave2 + wave3

			// Depth gradient (deeper toward center)
			depthT := clampF(dist*1.5, 0, 1)
			deepBlue := color.RGBA{15, 60, 150, 255}
			shallowBlue := color.RGBA{40, 140, 210, 255}
			base := lerpColor(shallowBlue, deepBlue, depthT*0.6)

			r := clampU8(float64(base.R) + wave*0.4)
			g := clampU8(float64(base.G) + wave*0.6)
			b := clampU8(float64(base.B) + wave*0.3)
			a := uint8(255)
			if dist < 0.04 {
				a = clampU8(dist / 0.04 * 255)
			}
			img.SetRGBA(px, py, color.RGBA{r, g, b, a})
		}
	}

	rng := rand.New(rand.NewSource(int64(45 + variant*100)))
	// Foam edges near diamond border
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			dist := diamondEdgeDist(px, py, w, h)
			if dist < 0.15 && dist > 0.03 {
				foamNoise := perlinish2(px, py, 8, float64(variant))
				if foamNoise > 0.05 {
					alpha := clampU8((0.15 - dist) / 0.12 * 120 * foamNoise * 5)
					setPixelBlend(img, px, py, color.RGBA{220, 240, 255, alpha})
				}
			}
		}
	}

	// Specular highlights (reflections)
	for i := 0; i < 15; i++ {
		px := w/4 + rng.Intn(w/2)
		py := h/4 + rng.Intn(h/2)
		if isInDiamond(px, py, w, h) && diamondEdgeDist(px, py, w, h) > 0.1 {
			size := 1 + rng.Intn(3)
			for dx := -size; dx <= size; dx++ {
				alpha := clampU8(float64(90-rng.Intn(30)) * (1.0 - float64(abs(dx))/float64(size+1)))
				setPixelBlend(img, px+dx, py, color.RGBA{200, 230, 255, alpha})
			}
		}
	}
	drawDiamondOutline(img, w, h, color.RGBA{0, 30, 80, 50})
}

func deepWaterTile(img *image.RGBA, w, h, variant int) {
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			dist := diamondEdgeDist(px, py, w, h)
			phase := float64(variant) * 1.2
			wave := math.Sin(float64(px)*0.15+float64(py)*0.12+phase)*12 +
				math.Sin(float64(px)*0.08-float64(py)*0.2+phase*0.5)*8
			r := clampU8(8 + wave*0.3)
			g := clampU8(30 + wave*0.4 + dist*15)
			b := clampU8(120 + wave*0.8)
			a := uint8(255)
			if dist < 0.04 {
				a = clampU8(dist / 0.04 * 255)
			}
			img.SetRGBA(px, py, color.RGBA{r, g, b, a})
		}
	}
	drawDiamondOutline(img, w, h, color.RGBA{0, 10, 50, 50})
}

func rockTile(img *image.RGBA, w, h, variant int) {
	seeds := []int64{46, 4444, 6666}
	seed := seeds[variant%3]
	rng := rand.New(rand.NewSource(seed))
	bases := []color.RGBA{
		{135, 130, 125, 255},
		{125, 125, 130, 255},
		{140, 135, 120, 255},
	}
	fillDiamondTextured(img, w, h, bases[variant%3], 0.2, seed)

	// Rock texture with cracks
	for i := 0; i < 8+variant*3; i++ {
		x := w/4 + rng.Intn(w/2)
		y := h/4 + rng.Intn(h/2)
		length := 6 + rng.Intn(12)
		for j := 0; j < length; j++ {
			if isInDiamond(x, y, w, h) {
				setPixelBlend(img, x, y, color.RGBA{70, 65, 60, uint8(80 + rng.Intn(80))})
				setPixelBlend(img, x+1, y, color.RGBA{70, 65, 60, uint8(40 + rng.Intn(40))})
			}
			x += rng.Intn(3) - 1
			y += rng.Intn(3) - 1
		}
	}

	// Highlights on upper-left edges of rocks
	for i := 0; i < 6; i++ {
		px := w/4 + rng.Intn(w/2)
		py := h/4 + rng.Intn(h/4)
		if isInDiamond(px, py, w, h) {
			setPixelBlend(img, px, py, color.RGBA{180, 180, 180, 60})
		}
	}
}

func cliffTile(img *image.RGBA, w, h, variant int) {
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			dist := diamondEdgeDist(px, py, w, h)
			t := float64(py) / float64(h)
			n := perlinish2(px, py, 12, float64(variant*7)) * 20
			r := clampU8(110 - t*50 + n)
			g := clampU8(105 - t*50 + n)
			b := clampU8(100 - t*50 + n*0.8)
			a := uint8(255)
			if dist < 0.04 {
				a = clampU8(dist / 0.04 * 255)
			}
			img.SetRGBA(px, py, color.RGBA{r, g, b, a})
		}
	}
	// Cliff face striations
	rng := rand.New(rand.NewSource(int64(200 + variant)))
	for i := 0; i < 5; i++ {
		y := h/3 + rng.Intn(h/3)
		for px := 0; px < w; px++ {
			if isInDiamond(px, y, w, h) {
				setPixelBlend(img, px, y, color.RGBA{80, 75, 70, 50})
			}
		}
	}
	drawDiamondOutline(img, w, h, color.RGBA{40, 40, 40, 80})
}

func roadTile(img *image.RGBA, w, h, variant int) {
	bases := []color.RGBA{
		{95, 95, 100, 255},
		{90, 90, 95, 255},
		{100, 100, 105, 255},
	}
	fillDiamondTextured(img, w, h, bases[variant%3], 0.06, int64(100+variant))

	// Asphalt texture (speckles)
	rng := rand.New(rand.NewSource(int64(500 + variant)))
	for i := 0; i < 60; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) {
			gray := uint8(70 + rng.Intn(50))
			setPixelBlend(img, px, py, color.RGBA{gray, gray, gray + 5, 40})
		}
	}

	// Center lane marking (dashed yellow)
	for py := 0; py < h; py++ {
		px := w / 2
		if isInDiamond(px, py, w, h) {
			segment := py / 6
			if segment%2 == 0 {
				setPixelBlend(img, px, py, color.RGBA{240, 220, 80, 200})
				setPixelBlend(img, px-1, py, color.RGBA{240, 220, 80, 140})
				setPixelBlend(img, px+1, py, color.RGBA{240, 220, 80, 140})
			}
		}
	}

	// Wear patterns
	if variant > 0 {
		for i := 0; i < 3; i++ {
			px := w/3 + rng.Intn(w/3)
			py := h/3 + rng.Intn(h/3)
			if isInDiamond(px, py, w, h) {
				fillCircle(img, px, py, 2+rng.Intn(3), color.RGBA{110, 110, 115, 40})
			}
		}
	}
}

func bridgeTile(img *image.RGBA, w, h, variant int) {
	base := color.RGBA{155, 105, 60, 255}
	fillDiamondTextured(img, w, h, base, 0.12, int64(300+variant))

	rng := rand.New(rand.NewSource(int64(301 + variant)))

	// Wood plank lines with grain
	for i := 0; i < 8; i++ {
		y := h/10 + i*(h*8/10)/7
		for px := 0; px < w; px++ {
			if isInDiamond(px, y, w, h) {
				setPixelBlend(img, px, y, color.RGBA{90, 55, 25, 120})
				// Wood grain
				if rng.Intn(3) == 0 {
					setPixelBlend(img, px, y-1, color.RGBA{130, 85, 40, 40})
				}
			}
		}
	}
	// Nail heads
	for i := 0; i < 4+variant; i++ {
		px := w/4 + rng.Intn(w/2)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) {
			fillCircle(img, px, py, 1, color.RGBA{80, 80, 90, 180})
		}
	}
}

func oreTile(img *image.RGBA, w, h, variant int) {
	base := color.RGBA{85, 75, 55, 255}
	fillDiamondTextured(img, w, h, base, 0.15, int64(47+variant*100))

	rng := rand.New(rand.NewSource(int64(47 + variant*100)))

	// Metallic veins
	for i := 0; i < 3+variant; i++ {
		x := w/4 + rng.Intn(w/2)
		y := h/4 + rng.Intn(h/2)
		length := 8 + rng.Intn(15)
		for j := 0; j < length; j++ {
			if isInDiamond(x, y, w, h) {
				setPixelBlend(img, x, y, color.RGBA{180, 150, 50, 160})
				setPixelBlend(img, x+1, y, color.RGBA{160, 130, 40, 100})
			}
			x += rng.Intn(3) - 1
			y += rng.Intn(3) - 1
		}
	}

	// Crystal/ore chunks with glow
	for i := 0; i < 15+variant*5; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) && diamondEdgeDist(px, py, w, h) > 0.1 {
			bright := uint8(200 + rng.Intn(55))
			size := 1 + rng.Intn(3)
			// Glow
			fillCircleGradient(img, px, py, size+2,
				color.RGBA{bright, bright - 40, 0, 80},
				color.RGBA{bright, bright - 40, 0, 0})
			// Crystal
			fillCircle(img, px, py, size, color.RGBA{bright, bright - 30, 10, 220})
			// Sparkle highlight
			setPixelBlend(img, px-1, py-1, color.RGBA{255, 255, 200, 180})
		}
	}
}

func gemTile(img *image.RGBA, w, h, variant int) {
	base := color.RGBA{45, 55, 85, 255}
	fillDiamondTextured(img, w, h, base, 0.12, int64(48+variant*100))

	rng := rand.New(rand.NewSource(int64(48 + variant*100)))

	// Gem crystals (larger, more detailed)
	gemColors := []color.RGBA{
		{30, 220, 255, 240},
		{50, 200, 240, 240},
		{0, 240, 220, 240},
	}
	for i := 0; i < 10+variant*3; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) && diamondEdgeDist(px, py, w, h) > 0.1 {
			gc := gemColors[rng.Intn(len(gemColors))]
			size := 2 + rng.Intn(2)
			// Glow aura
			fillCircleGradient(img, px, py, size+3,
				color.RGBA{gc.R / 2, gc.G / 2, gc.B / 2, 100},
				color.RGBA{gc.R / 4, gc.G / 4, gc.B / 4, 0})
			// Crystal body
			fillCircle(img, px, py, size, gc)
			// Facet highlights
			setPixelBlend(img, px-1, py-1, color.RGBA{220, 255, 255, 180})
			setPixelBlend(img, px, py-1, color.RGBA{200, 250, 255, 120})
		}
	}
}

func snowTile(img *image.RGBA, w, h, variant int) {
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			dist := diamondEdgeDist(px, py, w, h)
			n := perlinish2(px, py, 18, float64(variant*5)) * 12
			r := clampU8(238 + n)
			g := clampU8(240 + n)
			b := clampU8(248 + n*0.5)
			a := uint8(255)
			if dist < 0.04 {
				a = clampU8(dist / 0.04 * 255)
			}
			// Blue shadows at bottom-right (top-left light)
			lightFactor := 1.0 - float64(px)/float64(w)*0.06 - float64(py)/float64(h)*0.08
			r = clampU8(float64(r) * lightFactor)
			g = clampU8(float64(g) * lightFactor)
			// b stays bright for blue tint in shadows
			img.SetRGBA(px, py, color.RGBA{r, g, b, a})
		}
	}

	// Sparkle dots
	rng := rand.New(rand.NewSource(int64(700 + variant)))
	for i := 0; i < 10; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) && diamondEdgeDist(px, py, w, h) > 0.1 {
			setPixelBlend(img, px, py, color.RGBA{255, 255, 255, uint8(100 + rng.Intn(100))})
		}
	}
	drawDiamondOutline(img, w, h, color.RGBA{180, 190, 210, 35})
}

func urbanTile(img *image.RGBA, w, h, variant int) {
	base := color.RGBA{168, 168, 173, 255}
	fillDiamondTextured(img, w, h, base, 0.06, int64(800+variant))

	// Grid pattern (concrete panels)
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			if px%16 == 0 || py%8 == 0 {
				setPixelBlend(img, px, py, color.RGBA{140, 140, 145, 50})
			}
			// Subtle wear pattern
			if (px+variant*5)%32 < 2 && py%16 < 2 {
				setPixelBlend(img, px, py, color.RGBA{155, 155, 160, 30})
			}
		}
	}

	// Drain grate or detail based on variant
	if variant == 1 {
		cx, cy := w/2, h/2
		fillRect(img, cx-3, cy-2, 6, 4, color.RGBA{60, 60, 65, 180})
		for i := 0; i < 3; i++ {
			drawLineAA(img, cx-2+i*2, cy-1, cx-2+i*2, cy+1, color.RGBA{40, 40, 45, 200})
		}
	}
}

func forestTile(img *image.RGBA, w, h, variant int) {
	base := color.RGBA{32, 95, 32, 255}
	fillDiamondTextured(img, w, h, base, 0.15, int64(49+variant*100))

	rng := rand.New(rand.NewSource(int64(49 + variant*100)))

	// Undergrowth (scattered small bushes)
	for i := 0; i < 20; i++ {
		px := rng.Intn(w)
		py := rng.Intn(h)
		if isInDiamond(px, py, w, h) && diamondEdgeDist(px, py, w, h) > 0.1 {
			green := uint8(60 + rng.Intn(50))
			fillCircle(img, px, py, 2+rng.Intn(2), color.RGBA{15, green, 10, uint8(80 + rng.Intn(60))})
		}
	}

	// Tree canopies (detailed with shadow and highlight)
	numTrees := 5 + variant*2
	for i := 0; i < numTrees; i++ {
		px := w/5 + rng.Intn(w*3/5)
		py := h/5 + rng.Intn(h*3/5)
		if !isInDiamond(px, py, w, h) || diamondEdgeDist(px, py, w, h) < 0.15 {
			continue
		}
		treeR := 5 + rng.Intn(4)

		// Shadow (offset bottom-right)
		fillCircle(img, px+3, py+3, treeR, color.RGBA{8, 30, 8, 100})

		// Trunk (visible below canopy)
		fillRect(img, px-1, py+treeR-3, 3, 5, color.RGBA{80, 55, 30, 200})

		// Main canopy
		green := uint8(70 + rng.Intn(80))
		fillCircle(img, px, py, treeR, color.RGBA{20, green, 15, 230})

		// Secondary canopy layer
		fillCircle(img, px-1, py-1, treeR-1, color.RGBA{25, green + 15, 20, 180})

		// Highlight on top-left
		fillCircle(img, px-2, py-2, treeR/2, color.RGBA{50, green + 40, 35, 120})

		// Leaf detail dots
		for j := 0; j < 5; j++ {
			lx := px - treeR + rng.Intn(treeR*2)
			ly := py - treeR + rng.Intn(treeR*2)
			dx := float64(lx-px) / float64(treeR)
			dy := float64(ly-py) / float64(treeR)
			if dx*dx+dy*dy < 1.0 {
				leafGreen := uint8(60 + rng.Intn(100))
				setPixelBlend(img, lx, ly, color.RGBA{15, leafGreen, 10, uint8(80 + rng.Intn(80))})
			}
		}
	}
}

func concreteTile(img *image.RGBA, w, h, variant int) {
	base := color.RGBA{160, 160, 165, 255}
	fillDiamondTextured(img, w, h, base, 0.05, int64(900+variant))

	// Panel lines in diamond pattern
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			if !isInDiamond(px, py, w, h) {
				continue
			}
			gridX := (px + variant*7) % 32
			gridY := (py + variant*3) % 16
			if gridX == 0 || gridY == 0 {
				setPixelBlend(img, px, py, color.RGBA{130, 130, 135, 70})
			}
		}
	}

	// Floor plate details
	rng := rand.New(rand.NewSource(int64(901 + variant)))
	if variant == 1 {
		// Hazard stripes
		for py := 0; py < h; py++ {
			for px := 0; px < w; px++ {
				if !isInDiamond(px, py, w, h) {
					continue
				}
				dist := diamondEdgeDist(px, py, w, h)
				if dist < 0.12 && dist > 0.04 {
					if (px+py)%8 < 4 {
						setPixelBlend(img, px, py, color.RGBA{220, 180, 30, 100})
					}
				}
			}
		}
	}
	if variant == 2 {
		// Oil stain
		cx := w/2 + rng.Intn(20) - 10
		cy := h/2 + rng.Intn(10) - 5
		fillEllipse(img, cx, cy, 8+rng.Intn(5), 4+rng.Intn(3), color.RGBA{50, 45, 40, 60})
	}
}

// ===================== BUILDING SPRITES =====================

func constructionYardSprite(img *image.RGBA, w, h int, factionColor color.RGBA) {
	cx, cy := w/2, h/2+10

	// Ground shadow
	fillEllipse(img, cx+4, h-14, w/2-12, 12, color.RGBA{0, 0, 0, 50})

	// Main foundation platform
	drawIsoBoxDetailed(img, cx, cy+5, 60, 35, 8,
		color.RGBA{80, 80, 85, 255},
		color.RGBA{60, 60, 65, 255},
		color.RGBA{50, 50, 55, 255}, true)

	// Main building
	darker := color.RGBA{
		clampU8(float64(factionColor.R) * 0.6),
		clampU8(float64(factionColor.G) * 0.6),
		clampU8(float64(factionColor.B) * 0.6), 255}
	darkest := color.RGBA{
		clampU8(float64(factionColor.R) * 0.4),
		clampU8(float64(factionColor.G) * 0.4),
		clampU8(float64(factionColor.B) * 0.4), 255}

	drawIsoBoxDetailed(img, cx, cy-5, 45, 28, 28,
		factionColor, darker, darkest, true)

	// Crane arm
	drawThickLineAA(img, cx+20, cy-35, cx+40, cy-50, 2, color.RGBA{200, 200, 50, 255})
	drawThickLineAA(img, cx+40, cy-50, cx+55, cy-35, 2, color.RGBA{200, 200, 50, 255})
	// Crane cable
	drawLineAA(img, cx+45, cy-48, cx+45, cy-25, color.RGBA{120, 120, 120, 200})
	// Crane base
	fillRect(img, cx+17, cy-38, 8, 5, color.RGBA{180, 180, 40, 255})

	// Scaffolding on right side
	for i := 0; i < 4; i++ {
		y := cy - 30 + i*8
		drawLineAA(img, cx+30, y, cx+38, y, color.RGBA{160, 160, 160, 180})
	}
	drawLineAA(img, cx+30, cy-30, cx+30, cy+2, color.RGBA{160, 160, 160, 180})
	drawLineAA(img, cx+38, cy-30, cx+38, cy+2, color.RGBA{160, 160, 160, 180})

	// Windows
	for i := 0; i < 3; i++ {
		wy := cy - 25 + i*9
		fillRect(img, cx-15+i*12, wy, 5, 4, color.RGBA{140, 200, 230, 200})
		// Window frame
		drawLineAA(img, cx-15+i*12, wy, cx-10+i*12, wy, color.RGBA{60, 60, 70, 150})
	}

	// Detail panels and vents
	fillRect(img, cx-30, cy-15, 8, 6, color.RGBA{50, 50, 55, 200})
	for i := 0; i < 3; i++ {
		drawLineAA(img, cx-29, cy-14+i*2, cx-23, cy-14+i*2, color.RGBA{70, 70, 75, 180})
	}

	// Antenna with blinking light
	drawLineAA(img, cx-20, cy-35, cx-20, cy-55, color.RGBA{180, 180, 185, 255})
	fillCircle(img, cx-20, cy-56, 2, color.RGBA{255, 40, 40, 255})
	// Glow around light
	fillCircleGradient(img, cx-20, cy-56, 5,
		color.RGBA{255, 80, 80, 80},
		color.RGBA{255, 40, 40, 0})

	// Faction emblem (gear icon)
	fillCircle(img, cx, cy-18, 7, color.RGBA{255, 220, 50, 220})
	fillCircle(img, cx, cy-18, 4, factionColor)

	// Warning stripes on edges
	for i := 0; i < 8; i++ {
		x := cx - 40 + i*10
		if i%2 == 0 {
			fillRect(img, x, cy+8, 5, 3, color.RGBA{220, 180, 30, 180})
		} else {
			fillRect(img, x, cy+8, 5, 3, color.RGBA{30, 30, 30, 180})
		}
	}

	// Small lights along building edge
	for i := 0; i < 4; i++ {
		lx := cx - 30 + i*20
		fillCircle(img, lx, cy-2, 1, color.RGBA{255, 240, 180, 200})
	}
}

func powerPlantSprite(img *image.RGBA, w, h int, factionColor color.RGBA) {
	cx, cy := w/2, h/2+12

	fillEllipse(img, cx+3, h-12, w/2-14, 10, color.RGBA{0, 0, 0, 50})

	// Main building
	darker := color.RGBA{clampU8(float64(factionColor.R) * 0.7), clampU8(float64(factionColor.G) * 0.7), clampU8(float64(factionColor.B) * 0.7), 255}
	darkest := color.RGBA{clampU8(float64(factionColor.R) * 0.5), clampU8(float64(factionColor.G) * 0.5), clampU8(float64(factionColor.B) * 0.5), 255}
	drawIsoBoxDetailed(img, cx+10, cy, 30, 20, 22,
		factionColor, darker, darkest, true)

	// Cooling tower (cylindrical)
	towerCx, towerCy := cx-15, cy-10
	// Tower body
	for py := towerCy - 20; py <= towerCy; py++ {
		t := float64(py-towerCy+20) / 20.0
		radius := int(10 + t*3) // slightly wider at bottom
		gray := clampU8(190 + (1.0-t)*20)
		for px := towerCx - radius; px <= towerCx+radius; px++ {
			d := float64(px-towerCx) / float64(radius)
			if d*d <= 1.0 {
				lighting := 1.0 - d*0.2
				setPixelBlend(img, px, py, color.RGBA{
					clampU8(float64(gray) * lighting),
					clampU8(float64(gray) * lighting),
					clampU8(float64(gray+5) * lighting), 255})
			}
		}
	}
	// Tower top (ellipse)
	fillEllipse(img, towerCx, towerCy-21, 10, 5, color.RGBA{200, 200, 205, 255})

	// Steam from tower
	rng := rand.New(rand.NewSource(55))
	for i := 0; i < 6; i++ {
		sx := towerCx - 5 + rng.Intn(10)
		sy := towerCy - 25 - rng.Intn(15)
		size := 3 + rng.Intn(4)
		alpha := uint8(60 - i*8)
		if alpha > 200 {
			alpha = 0
		}
		fillCircle(img, sx, sy, size, color.RGBA{220, 225, 230, alpha})
	}

	// Pipes connecting tower to building
	drawThickLineAA(img, towerCx+8, towerCy-5, cx, cy-5, 2, color.RGBA{150, 150, 155, 230})
	drawThickLineAA(img, towerCx+8, towerCy, cx, cy, 2, color.RGBA{140, 140, 145, 230})

	// Generator detail on building
	fillRect(img, cx+2, cy-15, 12, 8, color.RGBA{80, 80, 85, 220})
	// Coil pattern
	for i := 0; i < 4; i++ {
		fillCircle(img, cx+5+i*3, cy-11, 1, color.RGBA{200, 160, 50, 200})
	}

	// Lightning bolt emblem
	fillTriangle(img, cx+8, cy-22, cx+14, cy-22, cx+10, cy-14, color.RGBA{255, 230, 50, 240})
	fillTriangle(img, cx+9, cy-16, cx+15, cy-16, cx+12, cy-8, color.RGBA{255, 210, 30, 240})

	// Warning stripes
	for i := 0; i < 6; i++ {
		x := cx - 5 + i*5
		c := color.RGBA{220, 180, 30, 160}
		if i%2 == 1 {
			c = color.RGBA{30, 30, 30, 160}
		}
		fillRect(img, x, cy+5, 4, 3, c)
	}
}

func barracksSprite(img *image.RGBA, w, h int, factionColor color.RGBA) {
	cx, cy := w/2, h/2+12

	fillEllipse(img, cx+3, h-12, w/2-14, 10, color.RGBA{0, 0, 0, 50})

	// Sandbag wall at base
	for i := 0; i < 7; i++ {
		sx := cx - 30 + i*10
		fillEllipse(img, sx, cy+10, 5, 3, color.RGBA{160, 145, 100, 230})
		fillEllipse(img, sx, cy+9, 5, 2, color.RGBA{175, 160, 115, 200})
	}

	// Main building (military green tinted with faction)
	milGreen := color.RGBA{
		clampU8(float64(factionColor.R)*0.3 + 40),
		clampU8(float64(factionColor.G)*0.3 + 80),
		clampU8(float64(factionColor.B)*0.3 + 40), 255}
	darker := color.RGBA{clampU8(float64(milGreen.R) * 0.75), clampU8(float64(milGreen.G) * 0.75), clampU8(float64(milGreen.B) * 0.75), 255}
	darkest := color.RGBA{clampU8(float64(milGreen.R) * 0.55), clampU8(float64(milGreen.G) * 0.55), clampU8(float64(milGreen.B) * 0.55), 255}

	drawIsoBoxDetailed(img, cx, cy-2, 38, 24, 25,
		milGreen, darker, darkest, true)

	// Door
	fillRectGradientV(img, cx-6, cy-2, 12, 16,
		color.RGBA{35, 55, 35, 255},
		color.RGBA{25, 40, 25, 255})
	// Door frame
	drawLineAA(img, cx-7, cy-3, cx+6, cy-3, color.RGBA{100, 100, 100, 200})
	drawLineAA(img, cx-7, cy-3, cx-7, cy+14, color.RGBA{100, 100, 100, 200})
	drawLineAA(img, cx+6, cy-3, cx+6, cy+14, color.RGBA{100, 100, 100, 200})

	// Windows (3 across)
	for i := 0; i < 3; i++ {
		wx := cx - 25 + i*18
		wy := cy - 18
		fillRect(img, wx, wy, 7, 5, color.RGBA{120, 180, 210, 200})
		// Cross bar
		drawLineAA(img, wx+3, wy, wx+3, wy+4, color.RGBA{60, 60, 65, 180})
		drawLineAA(img, wx, wy+2, wx+6, wy+2, color.RGBA{60, 60, 65, 180})
	}

	// Flag pole
	drawLineAA(img, cx+28, cy+6, cx+28, cy-35, color.RGBA{170, 170, 175, 255})
	// Flag (faction colored)
	fillTriangle(img, cx+28, cy-35, cx+40, cy-30, cx+28, cy-25, factionColor)
	// Flag shading
	fillTriangle(img, cx+28, cy-30, cx+40, cy-30, cx+28, cy-25,
		color.RGBA{clampU8(float64(factionColor.R) * 0.7),
			clampU8(float64(factionColor.G) * 0.7),
			clampU8(float64(factionColor.B) * 0.7), 200})

	// Star emblem
	fillCircle(img, cx, cy-25, 5, factionColor)
	fillCircle(img, cx, cy-25, 3, color.RGBA{255, 255, 220, 200})
}

func warFactorySprite(img *image.RGBA, w, h int, factionColor color.RGBA) {
	cx, cy := w/2, h/2+10

	fillEllipse(img, cx+4, h-14, w/2-12, 12, color.RGBA{0, 0, 0, 50})

	// Large industrial building
	buildColor := color.RGBA{120, 125, 135, 255}
	drawIsoBoxDetailed(img, cx, cy, 58, 34, 30,
		buildColor,
		color.RGBA{95, 100, 110, 255},
		color.RGBA{75, 80, 90, 255}, true)

	// Garage door (large)
	fillRectGradientV(img, cx-18, cy-8, 36, 24,
		color.RGBA{45, 45, 55, 255},
		color.RGBA{35, 35, 42, 255})

	// Roller door horizontal lines
	for i := 0; i < 6; i++ {
		y := cy - 6 + i*4
		drawLineAA(img, cx-16, y, cx+16, y, color.RGBA{60, 60, 70, 200})
	}

	// Warning stripes on door frame
	for i := 0; i < 8; i++ {
		y := cy - 8 + i*3
		if i%2 == 0 {
			fillRect(img, cx-20, y, 3, 3, color.RGBA{220, 180, 30, 200})
		} else {
			fillRect(img, cx-20, y, 3, 3, color.RGBA{30, 30, 30, 200})
		}
		if i%2 == 0 {
			fillRect(img, cx+18, y, 3, 3, color.RGBA{220, 180, 30, 200})
		} else {
			fillRect(img, cx+18, y, 3, 3, color.RGBA{30, 30, 30, 200})
		}
	}

	// Vehicle ramp
	for i := 0; i < 5; i++ {
		y := cy + 16 + i
		rampW := 30 + i*2
		fillRect(img, cx-rampW/2, y, rampW, 1, color.RGBA{100, 100, 105, uint8(200 - i*30)})
	}

	// Smokestack
	fillRect(img, cx+28, cy-42, 8, 20, color.RGBA{110, 110, 115, 255})
	fillEllipse(img, cx+32, cy-43, 4, 2, color.RGBA{120, 120, 125, 255})
	// Smoke
	for i := 0; i < 4; i++ {
		fillCircle(img, cx+30+i*2, cy-48-i*5, 3+i, color.RGBA{180, 180, 185, uint8(60 - i*12)})
	}

	// Tools on wall (wrenches)
	drawLineAA(img, cx+35, cy-20, cx+38, cy-12, color.RGBA{160, 160, 165, 200})
	drawLineAA(img, cx+40, cy-20, cx+37, cy-12, color.RGBA{160, 160, 165, 200})

	// Faction color stripe
	fillRect(img, cx-50, cy-30, 100, 3, factionColor)

	// Mechanical detail on roof
	fillRect(img, cx-10, cy-32, 20, 4, color.RGBA{90, 90, 95, 220})
	fillCircle(img, cx-5, cy-30, 2, color.RGBA{80, 80, 85, 200})
	fillCircle(img, cx+5, cy-30, 2, color.RGBA{80, 80, 85, 200})
}

func refinerySprite(img *image.RGBA, w, h int, factionColor color.RGBA) {
	cx, cy := w/2, h/2+10

	fillEllipse(img, cx+4, h-14, w/2-12, 12, color.RGBA{0, 0, 0, 50})

	// Main processing building
	drawIsoBoxDetailed(img, cx-10, cy, 35, 22, 25,
		color.RGBA{170, 145, 85, 255},
		color.RGBA{140, 115, 65, 255},
		color.RGBA{110, 90, 50, 255}, true)

	// Silo (tall cylinder)
	siloCx, siloCy := cx+30, cy-5
	for py := siloCy - 30; py <= siloCy; py++ {
		_ = float64(py-siloCy+30) / 30.0
		r := 10
		for px := siloCx - r; px <= siloCx+r; px++ {
			d := float64(px-siloCx) / float64(r)
			if d*d <= 1.0 {
				lighting := 1.0 - d*0.25
				gray := clampU8(160 * lighting)
				setPixelBlend(img, px, py, color.RGBA{gray, clampU8(float64(gray) * 0.9), clampU8(float64(gray) * 0.7), 255})
			}
		}
	}
	fillEllipse(img, siloCx, siloCy-31, 10, 4, color.RGBA{175, 155, 100, 255})
	// Silo rings
	for i := 0; i < 3; i++ {
		y := siloCy - 25 + i*10
		drawLineAA(img, siloCx-10, y, siloCx+10, y, color.RGBA{120, 100, 60, 150})
	}

	// Conveyor belt
	for i := 0; i < 6; i++ {
		bx := cx - 30 + i*8
		by := cy + 8
		fillRect(img, bx, by, 6, 3, color.RGBA{80, 80, 85, 220})
		fillCircle(img, bx+3, by+1, 2, color.RGBA{110, 110, 115, 220})
	}

	// Smoke stacks (2)
	for _, offset := range []int{-20, -8} {
		fillRect(img, cx+offset, cy-40, 5, 15, color.RGBA{100, 100, 105, 255})
		fillEllipse(img, cx+offset+2, cy-41, 3, 2, color.RGBA{110, 110, 115, 255})
		// Smoke wisps
		fillCircle(img, cx+offset+1, cy-45, 3, color.RGBA{190, 190, 195, 40})
		fillCircle(img, cx+offset, cy-50, 4, color.RGBA{200, 200, 205, 25})
	}

	// Ore symbol
	fillCircle(img, cx-10, cy-18, 6, color.RGBA{255, 200, 30, 220})
	fillCircle(img, cx-10, cy-18, 3, color.RGBA{255, 240, 100, 180})

	// Processing equipment
	fillRect(img, cx-25, cy-12, 10, 8, color.RGBA{90, 90, 95, 200})
	drawLineAA(img, cx-20, cy-12, cx-20, cy-4, color.RGBA{70, 70, 75, 180})

	// Faction accent
	fillRect(img, cx-40, cy-28, 3, 28, factionColor)
}

func radarSprite(img *image.RGBA, w, h int, factionColor color.RGBA) {
	cx, cy := w/2, h/2+15

	fillEllipse(img, cx+3, h-12, w/2-14, 10, color.RGBA{0, 0, 0, 50})

	// Base building
	darker := color.RGBA{clampU8(float64(factionColor.R) * 0.6), clampU8(float64(factionColor.G) * 0.6), clampU8(float64(factionColor.B) * 0.6), 255}
	darkest := color.RGBA{clampU8(float64(factionColor.R) * 0.4), clampU8(float64(factionColor.G) * 0.4), clampU8(float64(factionColor.B) * 0.4), 255}
	drawIsoBoxDetailed(img, cx, cy, 35, 22, 20, factionColor, darker, darkest, true)

	// Radar tower (tall post)
	drawThickLineAA(img, cx, cy-20, cx, cy-60, 3, color.RGBA{160, 160, 165, 255})

	// Radar dish (large ellipse, tilted)
	fillEllipse(img, cx, cy-62, 18, 8, color.RGBA{190, 195, 200, 255})
	fillEllipse(img, cx, cy-62, 15, 6, color.RGBA{170, 175, 180, 240})
	// Dish feed horn
	drawLineAA(img, cx, cy-68, cx, cy-75, color.RGBA{150, 150, 155, 255})
	fillCircle(img, cx, cy-76, 2, color.RGBA{180, 180, 185, 255})

	// Tech panels on building
	for i := 0; i < 3; i++ {
		fillRect(img, cx-20+i*14, cy-14, 8, 6, color.RGBA{40, 60, 80, 200})
		// Screen glow
		fillRect(img, cx-19+i*14, cy-13, 6, 4, color.RGBA{60, 200, 100, 150})
	}

	// Blinking lights
	fillCircle(img, cx-10, cy-20, 1, color.RGBA{0, 255, 0, 220})
	fillCircle(img, cx+10, cy-20, 1, color.RGBA{255, 0, 0, 220})
}

func turretSprite(img *image.RGBA, w, h int, factionColor color.RGBA) {
	cx, cy := w/2, h/2+8

	fillEllipse(img, cx+2, h-10, w/2-12, 8, color.RGBA{0, 0, 0, 50})

	// Concrete base
	drawIsoBox(img, cx, cy+5, 20, 12, 6,
		color.RGBA{140, 140, 145, 255},
		color.RGBA{110, 110, 115, 255},
		color.RGBA{90, 90, 95, 255})

	// Turret body
	fillEllipse(img, cx, cy-5, 14, 10, color.RGBA{100, 105, 110, 255})
	fillEllipse(img, cx, cy-7, 12, 8, factionColor)

	// Gun barrels (twin)
	fillRect(img, cx-3, cy-28, 2, 20, color.RGBA{70, 70, 75, 255})
	fillRect(img, cx+2, cy-28, 2, 20, color.RGBA{70, 70, 75, 255})
	// Muzzle
	fillRect(img, cx-4, cy-30, 3, 3, color.RGBA{80, 80, 85, 255})
	fillRect(img, cx+1, cy-30, 3, 3, color.RGBA{80, 80, 85, 255})

	// Detail rivets
	for i := 0; i < 4; i++ {
		angle := float64(i) * math.Pi / 2
		rx := cx + int(8*math.Cos(angle))
		ry := cy - 6 + int(5*math.Sin(angle))
		fillCircle(img, rx, ry, 1, color.RGBA{80, 80, 85, 200})
	}
}

func wallSprite(img *image.RGBA, w, h int, factionColor color.RGBA) {
	cx, cy := w/2, h/2

	// Wall segment (isometric box)
	drawIsoBox(img, cx, cy, 25, 14, 16,
		color.RGBA{150, 150, 155, 255},
		color.RGBA{120, 120, 125, 255},
		color.RGBA{100, 100, 105, 255})

	// Faction stripe
	drawLineAA(img, cx-20, cy-8, cx+20, cy-8, factionColor)

	// Concrete texture
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 10; i++ {
		px := cx - 20 + rng.Intn(40)
		py := cy - 14 + rng.Intn(20)
		setPixelBlend(img, px, py, color.RGBA{130, 130, 135, 40})
	}
}

// drawConstructionStage draws a building under construction
func drawConstructionStage(img *image.RGBA, w, h, stage int, factionColor color.RGBA) {
	cx, cy := w/2, h/2+10

	switch stage {
	case 0: // Foundation + scaffolding only
		// Foundation
		drawIsoBox(img, cx, cy+5, 40, 24, 4,
			color.RGBA{100, 100, 105, 255},
			color.RGBA{80, 80, 85, 255},
			color.RGBA{70, 70, 75, 255})
		// Scaffolding
		drawScaffolding(img, cx, cy, 35, 30)

	case 1: // Partial building
		drawIsoBox(img, cx, cy+5, 40, 24, 4,
			color.RGBA{100, 100, 105, 255},
			color.RGBA{80, 80, 85, 255},
			color.RGBA{70, 70, 75, 255})
		// Partial walls
		darker := color.RGBA{clampU8(float64(factionColor.R) * 0.6), clampU8(float64(factionColor.G) * 0.6), clampU8(float64(factionColor.B) * 0.6), 255}
		drawIsoBox(img, cx, cy, 38, 22, 15,
			factionColor, darker,
			color.RGBA{clampU8(float64(factionColor.R) * 0.4), clampU8(float64(factionColor.G) * 0.4), clampU8(float64(factionColor.B) * 0.4), 255})
		drawScaffolding(img, cx, cy-10, 30, 20)

	case 2: // Nearly complete (some scaffolding remains)
		darker := color.RGBA{clampU8(float64(factionColor.R) * 0.6), clampU8(float64(factionColor.G) * 0.6), clampU8(float64(factionColor.B) * 0.6), 255}
		drawIsoBoxDetailed(img, cx, cy, 40, 24, 25,
			factionColor, darker,
			color.RGBA{clampU8(float64(factionColor.R) * 0.4), clampU8(float64(factionColor.G) * 0.4), clampU8(float64(factionColor.B) * 0.4), 255}, true)
		// Remaining scaffolding on one side
		drawLineAA(img, cx+30, cy-25, cx+30, cy+5, color.RGBA{160, 160, 160, 150})
		drawLineAA(img, cx+35, cy-25, cx+35, cy+5, color.RGBA{160, 160, 160, 150})
		for i := 0; i < 3; i++ {
			y := cy - 20 + i*10
			drawLineAA(img, cx+30, y, cx+35, y, color.RGBA{160, 160, 160, 150})
		}
	}
}

func drawScaffolding(img *image.RGBA, cx, cy, halfW, height int) {
	scaffColor := color.RGBA{170, 170, 140, 180}
	// Vertical poles
	for _, x := range []int{cx - halfW, cx - halfW/2, cx, cx + halfW/2, cx + halfW} {
		drawLineAA(img, x, cy-height, x, cy, scaffColor)
	}
	// Horizontal bars
	for i := 0; i < 4; i++ {
		y := cy - height + i*height/3
		drawLineAA(img, cx-halfW, y, cx+halfW, y, scaffColor)
	}
	// Cross braces
	drawLineAA(img, cx-halfW, cy-height, cx, cy, color.RGBA{170, 170, 140, 100})
	drawLineAA(img, cx, cy-height, cx+halfW, cy, color.RGBA{170, 170, 140, 100})
}

func applyDamageOverlay(img *image.RGBA, w, h int) {
	rng := rand.New(rand.NewSource(666))
	// Burn marks
	for i := 0; i < 5; i++ {
		px := w/4 + rng.Intn(w/2)
		py := h/4 + rng.Intn(h/2)
		size := 3 + rng.Intn(6)
		fillCircle(img, px, py, size, color.RGBA{30, 25, 20, uint8(80 + rng.Intn(60))})
	}
	// Cracks
	for i := 0; i < 3; i++ {
		x := w/4 + rng.Intn(w/2)
		y := h/4 + rng.Intn(h/2)
		for j := 0; j < 12; j++ {
			setPixelBlend(img, x, y, color.RGBA{20, 20, 20, uint8(100 + rng.Intn(100))})
			x += rng.Intn(3) - 1
			y += rng.Intn(3) - 1
		}
	}
	// Fire glow spots
	for i := 0; i < 2; i++ {
		px := w/3 + rng.Intn(w/3)
		py := h/3 + rng.Intn(h/3)
		fillCircleGradient(img, px, py, 5,
			color.RGBA{255, 150, 30, 100},
			color.RGBA{255, 80, 0, 0})
	}
}

// ===================== UNIT SPRITES =====================

// drawUnitShadow draws a shadow ellipse beneath a unit
func drawUnitShadow(img *image.RGBA, cx, cy, rx, ry int) {
	fillEllipse(img, cx+2, cy, rx, ry, color.RGBA{0, 0, 0, 40})
}

func infantrySprite(img *image.RGBA, w, h, dir, frame int) {
	cx, cy := w/2, h/2
	angle := directionAngle(dir)

	drawUnitShadow(img, cx, cy+12, 7, 4)

	// Animation: slight bob for walking
	bob := 0
	if frame == 1 {
		bob = -1
	} else if frame == 2 {
		bob = 1
	}

	// Body direction offset
	dx := math.Cos(angle) * 2
	dy := math.Sin(angle) * 1

	bodyCx := cx + int(dx)
	bodyCy := cy + int(dy) + bob

	// Legs (animated stride)
	legSpread := 0
	if frame == 1 {
		legSpread = 2
	} else if frame == 2 {
		legSpread = -2
	}
	// Left leg
	drawThickLineAA(img, bodyCx-2, bodyCy+5, bodyCx-3-legSpread, bodyCy+11, 1.5, color.RGBA{45, 85, 45, 255})
	// Right leg
	drawThickLineAA(img, bodyCx+2, bodyCy+5, bodyCx+3+legSpread, bodyCy+11, 1.5, color.RGBA{45, 85, 45, 255})
	// Boots
	fillRect(img, bodyCx-5-legSpread, bodyCy+10, 4, 2, color.RGBA{45, 40, 35, 255})
	fillRect(img, bodyCx+2+legSpread, bodyCy+10, 4, 2, color.RGBA{45, 40, 35, 255})

	// Body (torso)
	fillRectGradientV(img, bodyCx-4, bodyCy-3, 8, 9,
		color.RGBA{55, 105, 55, 255},
		color.RGBA{45, 85, 45, 255})

	// Arms
	weaponDx := math.Cos(angle) * 5
	weaponDy := math.Sin(angle) * 3
	drawThickLineAA(img, bodyCx+3, bodyCy-1, bodyCx+int(weaponDx)+4, bodyCy+int(weaponDy)-2, 1.5, color.RGBA{50, 95, 50, 255})
	drawThickLineAA(img, bodyCx-3, bodyCy, bodyCx-2, bodyCy+3, 1.5, color.RGBA{50, 95, 50, 255})

	// Head
	fillCircle(img, bodyCx, bodyCy-6, 4, color.RGBA{200, 170, 140, 255})
	// Helmet
	fillCircle(img, bodyCx, bodyCy-8, 4, color.RGBA{60, 95, 60, 250})
	fillCircle(img, bodyCx, bodyCy-7, 4, color.RGBA{65, 100, 65, 200})

	// Weapon (rifle pointing in facing direction)
	rifleEndX := bodyCx + int(weaponDx) + 5
	rifleEndY := bodyCy + int(weaponDy) - 5
	drawThickLineAA(img, bodyCx+3, bodyCy-2, rifleEndX, rifleEndY, 1.5, color.RGBA{75, 75, 70, 240})
}

func tankSprite(img *image.RGBA, w, h, dir, frame int) {
	cx, cy := w/2, h/2
	angle := directionAngle(dir)

	drawUnitShadow(img, cx, cy+16, 20, 10)

	// Track bob animation
	trackOffset := 0
	if frame == 1 {
		trackOffset = 1
	} else if frame == 2 {
		trackOffset = -1
	}

	// Hull rotation (simplified as visual offset)
	hullDx := math.Cos(angle) * 3
	hullDy := math.Sin(angle) * 1.5
	hcx := cx + int(hullDx)
	hcy := cy + int(hullDy)

	// Tracks (dark base)
	fillEllipse(img, hcx, hcy+6+trackOffset, 22, 9, color.RGBA{50, 50, 45, 255})
	// Track detail
	fillEllipse(img, hcx, hcy+6+trackOffset, 20, 7, color.RGBA{65, 65, 58, 255})
	// Track segments
	for i := -4; i <= 4; i++ {
		x := hcx + i*5
		drawLineAA(img, x, hcy+1+trackOffset, x, hcy+11+trackOffset, color.RGBA{55, 55, 48, 180})
	}
	// Track wheels
	for i := -3; i <= 3; i++ {
		fillCircle(img, hcx+i*6, hcy+6+trackOffset, 2, color.RGBA{75, 75, 70, 200})
	}

	// Hull (isometric box, rotated look)
	drawIsoBox(img, hcx, hcy+1, 16, 9, 7,
		color.RGBA{95, 115, 85, 255},
		color.RGBA{75, 90, 65, 255},
		color.RGBA{60, 78, 52, 255})

	// Reactive armor plates
	for i := 0; i < 3; i++ {
		fillRect(img, hcx-12+i*9, hcy-4, 7, 3, color.RGBA{85, 105, 75, 230})
	}

	// Turret (always shows barrel pointing in direction)
	fillEllipse(img, hcx, hcy-5, 9, 6, color.RGBA{85, 105, 75, 255})
	fillEllipse(img, hcx, hcy-6, 8, 5, color.RGBA{100, 120, 85, 255})

	// Gun barrel (points in direction)
	barrelLen := 14.0
	barrelEndX := hcx + int(math.Cos(angle)*barrelLen)
	barrelEndY := hcy - 6 + int(math.Sin(angle)*barrelLen*0.5)
	drawThickLineAA(img, hcx, hcy-6, barrelEndX, barrelEndY, 2.5, color.RGBA{65, 65, 58, 255})
	// Muzzle brake
	mx := hcx + int(math.Cos(angle)*(barrelLen+2))
	my := hcy - 6 + int(math.Sin(angle)*(barrelLen+2)*0.5)
	fillCircle(img, mx, my, 2, color.RGBA{75, 75, 68, 255})

	// Commander hatch
	fillCircle(img, hcx+3, hcy-5, 2, color.RGBA{78, 95, 68, 255})
}

func harvesterSprite(img *image.RGBA, w, h, dir, frame int) {
	cx, cy := w/2, h/2
	angle := directionAngle(dir)

	drawUnitShadow(img, cx, cy+16, 20, 10)

	bob := 0
	if frame == 1 {
		bob = 1
	}

	hullDx := math.Cos(angle) * 2
	hullDy := math.Sin(angle) * 1
	hcx := cx + int(hullDx)
	hcy := cy + int(hullDy) + bob

	// Treads
	fillEllipse(img, hcx, hcy+8, 20, 7, color.RGBA{50, 50, 45, 255})
	fillEllipse(img, hcx, hcy+8, 18, 5, color.RGBA{65, 65, 58, 255})

	// Cargo container (main body)
	drawIsoBox(img, hcx, hcy+2, 18, 11, 12,
		color.RGBA{210, 170, 45, 255},
		color.RGBA{180, 140, 35, 255},
		color.RGBA{150, 115, 25, 255})

	// Container fill indicator
	fillRect(img, hcx-10, hcy-6, 20, 3, color.RGBA{230, 190, 60, 255})

	// Scoop arm (points in direction)
	scoopAngle := angle + math.Pi // scoop at front
	scoopX := hcx + int(math.Cos(scoopAngle)*16)
	scoopY := hcy + int(math.Sin(scoopAngle)*8) + 4
	drawThickLineAA(img, hcx-int(math.Cos(angle)*8), hcy+2, scoopX, scoopY, 2, color.RGBA{150, 150, 145, 240})
	// Scoop bucket
	scoopBobble := 0
	if frame == 2 {
		scoopBobble = 2
	}
	fillRect(img, scoopX-4, scoopY-2+scoopBobble, 8, 5, color.RGBA{170, 170, 165, 230})

	// Cab
	cabX := hcx + int(math.Cos(angle)*8)
	cabY := hcy - 10 + int(math.Sin(angle)*3)
	fillRect(img, cabX-4, cabY, 8, 7, color.RGBA{190, 150, 35, 255})
	// Window
	fillRect(img, cabX-3, cabY+1, 6, 3, color.RGBA{140, 200, 225, 200})
}

func mcvSprite(img *image.RGBA, w, h, dir, frame int) {
	cx, cy := w/2, h/2
	angle := directionAngle(dir)

	drawUnitShadow(img, cx, cy+20, 24, 12)

	bob := 0
	if frame == 1 {
		bob = -1
	} else if frame == 2 {
		bob = 1
	}

	hullDx := math.Cos(angle) * 3
	hullDy := math.Sin(angle) * 1.5
	hcx := cx + int(hullDx)
	hcy := cy + int(hullDy) + bob

	// Large wheels
	fillEllipse(img, hcx, hcy+12, 24, 9, color.RGBA{45, 45, 40, 255})
	fillEllipse(img, hcx, hcy+12, 22, 7, color.RGBA{60, 60, 55, 255})
	// Wheel details
	for i := -3; i <= 3; i++ {
		fillCircle(img, hcx+i*7, hcy+12, 3, color.RGBA{55, 55, 50, 220})
		fillCircle(img, hcx+i*7, hcy+12, 1, color.RGBA{70, 70, 65, 200})
	}

	// Large body (purple/blue)
	drawIsoBox(img, hcx, hcy+2, 22, 14, 16,
		color.RGBA{85, 65, 185, 255},
		color.RGBA{65, 48, 155, 255},
		color.RGBA{50, 38, 125, 255})

	// Deploy mechanism on top
	fillRect(img, hcx-10, hcy-18, 20, 7, color.RGBA{105, 85, 205, 255})
	fillRect(img, hcx-8, hcy-20, 16, 5, color.RGBA{125, 105, 225, 255})
	// Deploy hinges
	fillCircle(img, hcx-8, hcy-15, 2, color.RGBA{150, 150, 155, 200})
	fillCircle(img, hcx+8, hcy-15, 2, color.RGBA{150, 150, 155, 200})

	// Satellite dish on top
	fillEllipse(img, hcx, hcy-24, 6, 4, color.RGBA{185, 185, 195, 255})
	fillCircle(img, hcx, hcy-24, 2, color.RGBA{150, 150, 165, 255})
	drawLineAA(img, hcx, hcy-20, hcx, hcy-24, color.RGBA{165, 165, 175, 255})

	// Cab (front based on direction)
	cabX := hcx + int(math.Cos(angle)*12)
	cabY := hcy - 8 + int(math.Sin(angle)*5)
	fillRect(img, cabX-5, cabY, 10, 9, color.RGBA{75, 58, 165, 255})
	fillRect(img, cabX-4, cabY+1, 8, 4, color.RGBA{140, 185, 215, 200})

	// Star emblem
	fillCircle(img, hcx-5, hcy-8, 3, color.RGBA{255, 225, 55, 210})
}

func engineerSprite(img *image.RGBA, w, h, dir, frame int) {
	cx, cy := w/2, h/2
	angle := directionAngle(dir)

	drawUnitShadow(img, cx, cy+12, 7, 4)

	bob := 0
	if frame == 1 {
		bob = -1
	} else if frame == 2 {
		bob = 1
	}

	dx := math.Cos(angle) * 2
	dy := math.Sin(angle) * 1
	bcx := cx + int(dx)
	bcy := cy + int(dy) + bob

	// Legs
	legSpread := frame - 1
	drawThickLineAA(img, bcx-2, bcy+5, bcx-3-legSpread, bcy+11, 1.5, color.RGBA{50, 50, 120, 255})
	drawThickLineAA(img, bcx+2, bcy+5, bcx+3+legSpread, bcy+11, 1.5, color.RGBA{50, 50, 120, 255})
	fillRect(img, bcx-5-legSpread, bcy+10, 4, 2, color.RGBA{40, 40, 40, 255})
	fillRect(img, bcx+2+legSpread, bcy+10, 4, 2, color.RGBA{40, 40, 40, 255})

	// Body (blue overalls)
	fillRectGradientV(img, bcx-4, bcy-3, 8, 9,
		color.RGBA{50, 50, 140, 255},
		color.RGBA{40, 40, 110, 255})

	// Hard hat (yellow)
	fillCircle(img, bcx, bcy-6, 4, color.RGBA{200, 170, 140, 255})
	fillCircle(img, bcx, bcy-8, 5, color.RGBA{240, 220, 50, 250})

	// Wrench (tool)
	toolDx := math.Cos(angle) * 6
	toolDy := math.Sin(angle) * 3
	drawThickLineAA(img, bcx+3, bcy, bcx+int(toolDx)+5, bcy+int(toolDy)+2, 1.5, color.RGBA{160, 160, 165, 240})
}

func attackDogSprite(img *image.RGBA, w, h, dir, frame int) {
	cx, cy := w/2, h/2
	angle := directionAngle(dir)

	drawUnitShadow(img, cx, cy+10, 8, 4)

	dx := math.Cos(angle) * 3
	dy := math.Sin(angle) * 1.5
	bcx := cx + int(dx)
	bcy := cy + int(dy)

	// Animation: running stride
	stride := 0
	if frame == 1 {
		stride = 2
	} else if frame == 2 {
		stride = -2
	}

	// Legs (4)
	legColor := color.RGBA{100, 70, 40, 255}
	drawThickLineAA(img, bcx-3, bcy+3, bcx-5-stride, bcy+9, 1.2, legColor)
	drawThickLineAA(img, bcx-1, bcy+3, bcx-2+stride, bcy+9, 1.2, legColor)
	drawThickLineAA(img, bcx+1, bcy+3, bcx+2-stride, bcy+9, 1.2, legColor)
	drawThickLineAA(img, bcx+3, bcy+3, bcx+5+stride, bcy+9, 1.2, legColor)

	// Body
	fillEllipse(img, bcx, bcy, 7, 4, color.RGBA{110, 80, 45, 255})
	fillEllipse(img, bcx, bcy-1, 6, 3, color.RGBA{120, 85, 50, 240})

	// Head
	headX := bcx + int(math.Cos(angle)*6)
	headY := bcy - 2 + int(math.Sin(angle)*2)
	fillCircle(img, headX, headY, 3, color.RGBA{115, 80, 45, 255})
	// Snout
	snoutX := headX + int(math.Cos(angle)*3)
	snoutY := headY + int(math.Sin(angle)*1)
	fillCircle(img, snoutX, snoutY, 2, color.RGBA{105, 75, 40, 255})
	// Eye
	setPixelBlend(img, headX-1, headY-1, color.RGBA{30, 30, 30, 255})

	// Tail
	tailX := bcx - int(math.Cos(angle)*7)
	tailY := bcy - 3 - int(math.Sin(angle)*2)
	drawLineAA(img, bcx-int(math.Cos(angle)*5), bcy-1, tailX, tailY, color.RGBA{100, 70, 40, 220})
}

func apocalypseTankSprite(img *image.RGBA, w, h, dir, frame int) {
	cx, cy := w/2, h/2
	angle := directionAngle(dir)

	drawUnitShadow(img, cx, cy+20, 26, 12)

	trackBob := 0
	if frame == 1 {
		trackBob = 1
	}

	hullDx := math.Cos(angle) * 3
	hullDy := math.Sin(angle) * 1.5
	hcx := cx + int(hullDx)
	hcy := cy + int(hullDy) + trackBob

	// Heavy tracks
	fillEllipse(img, hcx, hcy+8, 26, 10, color.RGBA{45, 42, 38, 255})
	fillEllipse(img, hcx, hcy+8, 24, 8, color.RGBA{60, 58, 52, 255})
	for i := -5; i <= 5; i++ {
		x := hcx + i*5
		drawLineAA(img, x, hcy+2, x, hcy+14, color.RGBA{50, 48, 42, 180})
	}

	// Heavy hull
	drawIsoBox(img, hcx, hcy+2, 22, 12, 10,
		color.RGBA{140, 50, 50, 255},
		color.RGBA{110, 35, 35, 255},
		color.RGBA{85, 25, 25, 255})

	// Extra armor plates
	fillRect(img, hcx-18, hcy-5, 36, 3, color.RGBA{120, 40, 40, 230})
	fillRect(img, hcx-16, hcy-3, 32, 2, color.RGBA{100, 30, 30, 220})

	// Twin turret
	fillEllipse(img, hcx, hcy-8, 12, 8, color.RGBA{130, 45, 45, 255})
	fillEllipse(img, hcx, hcy-9, 11, 7, color.RGBA{145, 55, 55, 255})

	// Twin gun barrels
	barrelLen := 16.0
	for offset := -2.0; offset <= 2.0; offset += 4.0 {
		perpX := -math.Sin(angle) * offset
		perpY := math.Cos(angle) * offset * 0.5
		bx := hcx + int(perpX)
		by := hcy - 9 + int(perpY)
		endX := bx + int(math.Cos(angle)*barrelLen)
		endY := by + int(math.Sin(angle)*barrelLen*0.5)
		drawThickLineAA(img, bx, by, endX, endY, 2.0, color.RGBA{60, 25, 25, 255})
	}

	// Soviet star
	fillCircle(img, hcx, hcy-8, 3, color.RGBA{255, 220, 50, 200})
}

func v3RocketSprite(img *image.RGBA, w, h, dir, frame int) {
	cx, cy := w/2, h/2
	angle := directionAngle(dir)

	drawUnitShadow(img, cx, cy+16, 20, 10)

	bob := 0
	if frame == 1 {
		bob = 1
	}

	hullDx := math.Cos(angle) * 2
	hullDy := math.Sin(angle) * 1
	hcx := cx + int(hullDx)
	hcy := cy + int(hullDy) + bob

	// Wheels
	fillEllipse(img, hcx, hcy+8, 18, 6, color.RGBA{50, 50, 45, 255})

	// Truck body
	drawIsoBox(img, hcx, hcy+2, 16, 10, 8,
		color.RGBA{80, 95, 75, 255},
		color.RGBA{60, 75, 55, 255},
		color.RGBA{48, 60, 42, 255})

	// Rocket launcher rail (angled upward)
	railBaseX := hcx
	railBaseY := hcy - 6
	railTopX := hcx + int(math.Cos(angle)*12)
	railTopY := hcy - 22 + int(math.Sin(angle)*5)
	drawThickLineAA(img, railBaseX-2, railBaseY, railTopX-2, railTopY, 2, color.RGBA{100, 100, 95, 240})
	drawThickLineAA(img, railBaseX+2, railBaseY, railTopX+2, railTopY, 2, color.RGBA{100, 100, 95, 240})

	// Rocket (on the rail)
	rocketX := (railBaseX + railTopX) / 2
	rocketY := (railBaseY + railTopY) / 2
	fillEllipse(img, rocketX, rocketY, 3, 8, color.RGBA{180, 180, 180, 255})
	// Nose cone (red)
	fillCircle(img, rocketX, rocketY-6, 2, color.RGBA{200, 50, 50, 255})
	// Fins
	drawLineAA(img, rocketX-3, rocketY+5, rocketX-5, rocketY+8, color.RGBA{150, 150, 145, 230})
	drawLineAA(img, rocketX+3, rocketY+5, rocketX+5, rocketY+8, color.RGBA{150, 150, 145, 230})

	// Cab
	cabX := hcx - int(math.Cos(angle)*10)
	cabY := hcy - 5 - int(math.Sin(angle)*4)
	fillRect(img, cabX-4, cabY, 8, 7, color.RGBA{70, 85, 65, 255})
	fillRect(img, cabX-3, cabY+1, 6, 3, color.RGBA{140, 190, 210, 200})
}

// ===================== VISUAL EFFECTS =====================

func explosionFrame(img *image.RGBA, w, h, frame int) {
	cx, cy := w/2, h/2
	t := float64(frame) / 7.0 // 0.0 to 1.0

	// Expanding fireball
	maxR := float64(w) / 2 * 0.9
	r := int(maxR * (0.2 + t*0.8))

	if t < 0.5 {
		// Hot core (white → yellow → orange)
		coreR := int(float64(r) * (1.0 - t))
		fillCircleGradient(img, cx, cy, coreR,
			color.RGBA{255, 255, 220, 255},
			color.RGBA{255, 200, 50, 200})
	}

	// Fireball
	alpha := clampU8(255 * (1.0 - t*0.8))
	fillCircleGradient(img, cx, cy, r,
		color.RGBA{255, 180, 30, alpha},
		color.RGBA{200, 60, 10, alpha / 2})

	// Outer smoke ring (later frames)
	if t > 0.3 {
		smokeAlpha := clampU8(150 * (t - 0.3) / 0.7 * (1.0 - t))
		fillCircleGradient(img, cx, cy, r+4,
			color.RGBA{80, 80, 80, 0},
			color.RGBA{60, 60, 60, smokeAlpha})
	}

	// Sparks
	rng := rand.New(rand.NewSource(int64(frame * 77)))
	if t < 0.7 {
		numSparks := 8 - frame
		if numSparks < 0 {
			numSparks = 0
		}
		for i := 0; i < numSparks; i++ {
			angle := rng.Float64() * math.Pi * 2
			dist := float64(r) * (0.5 + rng.Float64()*0.8)
			sx := cx + int(math.Cos(angle)*dist)
			sy := cy + int(math.Sin(angle)*dist)
			sparkAlpha := clampU8(255 * (1.0 - t))
			setPixelBlend(img, sx, sy, color.RGBA{255, 255, 150, sparkAlpha})
			setPixelBlend(img, sx+1, sy, color.RGBA{255, 200, 50, sparkAlpha / 2})
		}
	}
}

func muzzleFlashFrame(img *image.RGBA, w, h, frame int) {
	cx, cy := w/2, h/2
	t := float64(frame) / 2.0

	// Central flash
	r := int(float64(w)/4 * (1.0 - t*0.5))
	alpha := clampU8(255 * (1.0 - t*0.6))

	fillCircleGradient(img, cx, cy, r,
		color.RGBA{255, 255, 220, alpha},
		color.RGBA{255, 200, 50, alpha / 3})

	// Flash spikes
	rng := rand.New(rand.NewSource(int64(frame * 33)))
	numSpikes := 4 + rng.Intn(3)
	for i := 0; i < numSpikes; i++ {
		angle := rng.Float64() * math.Pi * 2
		length := float64(r) * (1.0 + rng.Float64()*0.8)
		endX := cx + int(math.Cos(angle)*length)
		endY := cy + int(math.Sin(angle)*length)
		drawLineAA(img, cx, cy, endX, endY, color.RGBA{255, 240, 150, alpha})
	}
}

func smokeFrame(img *image.RGBA, w, h, frame int) {
	cx, cy := w/2, h/2
	t := float64(frame) / 3.0

	// Rising smoke puff
	riseY := int(t * 6)
	r := int(float64(w)/4*(0.5+t*0.5)) + 1
	alpha := clampU8(160 * (1.0 - t*0.7))

	fillCircleGradient(img, cx, cy-riseY, r,
		color.RGBA{160, 160, 165, alpha},
		color.RGBA{120, 120, 125, alpha / 3})

	// Secondary smaller puff
	if frame > 0 {
		fillCircle(img, cx+3, cy-riseY+2, r/2,
			color.RGBA{140, 140, 145, alpha / 2})
	}
}

func oreSparkleFrame(img *image.RGBA, w, h, frame int) {
	cx, cy := w/2, h/2
	t := float64(frame) / 3.0

	// Twinkling star shape
	size := 2 + int(math.Sin(t*math.Pi)*3)
	alpha := clampU8(200 + math.Sin(t*math.Pi*2)*55)

	// Cross pattern
	for i := -size; i <= size; i++ {
		a := clampU8(float64(alpha) * (1.0 - math.Abs(float64(i))/float64(size+1)))
		setPixelBlend(img, cx+i, cy, color.RGBA{255, 230, 100, a})
		setPixelBlend(img, cx, cy+i, color.RGBA{255, 230, 100, a})
	}
	// Center bright dot
	setPixelBlend(img, cx, cy, color.RGBA{255, 255, 220, 255})
}

func selectionCircle(img *image.RGBA, w, h int) {
	cx, cy := w/2, h/2
	rx, ry := w/2-2, h/2-2

	// Green glow ellipse ring
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			dx := float64(px-cx) / float64(rx)
			dy := float64(py-cy) / float64(ry)
			d := math.Sqrt(dx*dx + dy*dy)
			if d > 0.7 && d < 1.0 {
				t := (d - 0.7) / 0.3
				alpha := clampU8(180 * math.Sin(t*math.Pi))
				setPixelBlend(img, px, py, color.RGBA{50, 255, 50, alpha})
			}
			// Inner glow
			if d < 0.75 && d > 0.6 {
				alpha := clampU8(60 * (1.0 - (0.75-d)/0.15))
				setPixelBlend(img, px, py, color.RGBA{100, 255, 100, alpha})
			}
		}
	}
}

func rallyFlag(img *image.RGBA, w, h int) {
	cx := w / 2

	// Pole
	drawThickLineAA(img, cx, h-2, cx, 4, 1.5, color.RGBA{180, 180, 185, 255})

	// Flag (green triangle)
	fillTriangle(img, cx, 4, cx+8, 8, cx, 12, color.RGBA{50, 220, 50, 230})
	// Flag highlight
	fillTriangle(img, cx, 4, cx+5, 6, cx, 8, color.RGBA{80, 255, 80, 150})

	// Base
	fillCircle(img, cx, h-2, 2, color.RGBA{120, 120, 125, 255})
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Tool to process downloaded Kenney assets into the format our RTS engine expects.
// Resizes/copies isometric tiles, buildings, units, effects into the right directories.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

var baseDir string

func main() {
	// Find project root
	baseDir = findProjectRoot()
	fmt.Printf("Project root: %s\n", baseDir)

	dlDir := filepath.Join(baseDir, "assets", "downloads")

	// Process terrain tiles from Kenney Isometric Landscape
	processTerrainTiles(dlDir)

	// Process building sprites from Kenney Isometric City
	processBuildingSprites(dlDir)

	// Process unit sprites from Kenney Tanks
	processUnitSprites(dlDir)

	// Process effects from Kenney Smoke Particles + Particle Pack
	processEffects(dlDir)

	fmt.Println("\n✅ All assets processed successfully!")
}

func findProjectRoot() string {
	// Try relative paths
	candidates := []string{
		".",
		"../..",
		filepath.Join(os.Getenv("HOME"), ".openclaw/workspace/sa/rts-engine"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "go.mod")); err == nil {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	fmt.Println("Could not find project root!")
	os.Exit(1)
	return ""
}

// loadPNG loads a PNG image
func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

// savePNG saves an image as PNG
func savePNG(path string, img image.Image) error {
	os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// resizeNearestNeighbor resizes an image using nearest-neighbor (crisp for pixel art)
func resizeNearestNeighbor(src image.Image, dstW, dstH int) *image.NRGBA {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))

	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			srcX := srcBounds.Min.X + x*srcW/dstW
			srcY := srcBounds.Min.Y + y*srcH/dstH
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

// resizeBilinear resizes using bilinear interpolation (smoother)
func resizeBilinear(src image.Image, dstW, dstH int) *image.NRGBA {
	srcBounds := src.Bounds()
	srcW := float64(srcBounds.Dx())
	srcH := float64(srcBounds.Dy())
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))

	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			// Map destination pixel to source coordinates
			srcX := (float64(x) + 0.5) * srcW / float64(dstW) - 0.5
			srcY := (float64(y) + 0.5) * srcH / float64(dstH) - 0.5

			x0 := int(math.Floor(srcX))
			y0 := int(math.Floor(srcY))
			x1 := x0 + 1
			y1 := y0 + 1

			// Clamp
			if x0 < 0 { x0 = 0 }
			if y0 < 0 { y0 = 0 }
			if x1 >= int(srcW) { x1 = int(srcW) - 1 }
			if y1 >= int(srcH) { y1 = int(srcH) - 1 }

			fx := srcX - float64(x0)
			fy := srcY - float64(y0)

			r00, g00, b00, a00 := src.At(srcBounds.Min.X+x0, srcBounds.Min.Y+y0).RGBA()
			r10, g10, b10, a10 := src.At(srcBounds.Min.X+x1, srcBounds.Min.Y+y0).RGBA()
			r01, g01, b01, a01 := src.At(srcBounds.Min.X+x0, srcBounds.Min.Y+y1).RGBA()
			r11, g11, b11, a11 := src.At(srcBounds.Min.X+x1, srcBounds.Min.Y+y1).RGBA()

			lerp := func(a, b, c, d uint32, fx, fy float64) uint8 {
				top := float64(a)*(1-fx) + float64(b)*fx
				bot := float64(c)*(1-fx) + float64(d)*fx
				v := top*(1-fy) + bot*fy
				return uint8(v / 256)
			}

			r := lerp(r00, r10, r01, r11, fx, fy)
			g := lerp(g00, g10, g01, g11, fx, fy)
			b := lerp(b00, b10, b01, b11, fx, fy)
			a := lerp(a00, a10, a01, a11, fx, fy)

			dst.SetNRGBA(x, y, color.NRGBA{r, g, b, a})
		}
	}
	return dst
}

// copyAndResize copies a source PNG to dest, resizing to target dimensions
func copyAndResize(src, dst string, w, h int) error {
	img, err := loadPNG(src)
	if err != nil {
		return err
	}
	resized := resizeBilinear(img, w, h)
	return savePNG(dst, resized)
}

// copyFile copies a PNG file directly (no resize)
func copyFile(src, dst string) error {
	img, err := loadPNG(src)
	if err != nil {
		return err
	}
	return savePNG(dst, img)
}

func processTerrainTiles(dlDir string) {
	fmt.Println("\n🌍 Processing terrain tiles...")
	tilesDir := filepath.Join(baseDir, "assets", "tiles")
	landscapeDir := filepath.Join(dlDir, "kenney_isometric_landscape", "PNG")
	roadsDir := filepath.Join(dlDir, "kenney_isometric_roads", "png")

	tw, th := 128, 64

	// Mapping: terrain name -> [source files for variants]
	// Based on visual inspection of Kenney landscape tiles:
	// 010 = flat grass, 014 = flat grass variant, 020 = sand
	// 035 = water, 037 = water variant, 060 = water variant
	// 070 = grass+water, 090 = road/stone, 100 = grass+feature
	terrainMap := map[string][]string{
		"grass": {
			filepath.Join(landscapeDir, "landscapeTiles_010.png"), // flat grass
			filepath.Join(landscapeDir, "landscapeTiles_014.png"), // grass variant
			filepath.Join(landscapeDir, "landscapeTiles_002.png"), // grass variant
		},
		"dirt": {
			filepath.Join(landscapeDir, "landscapeTiles_003.png"),
			filepath.Join(landscapeDir, "landscapeTiles_007.png"),
			filepath.Join(landscapeDir, "landscapeTiles_016.png"),
		},
		"sand": {
			filepath.Join(landscapeDir, "landscapeTiles_020.png"),
			filepath.Join(landscapeDir, "landscapeTiles_003.png"),
			filepath.Join(landscapeDir, "landscapeTiles_007.png"),
		},
		"water": {
			filepath.Join(landscapeDir, "landscapeTiles_035.png"),
			filepath.Join(landscapeDir, "landscapeTiles_037.png"),
			filepath.Join(landscapeDir, "landscapeTiles_060.png"),
		},
		"deep_water": {
			filepath.Join(landscapeDir, "landscapeTiles_035.png"),
			filepath.Join(landscapeDir, "landscapeTiles_060.png"),
			filepath.Join(landscapeDir, "landscapeTiles_037.png"),
		},
		"rock": {
			filepath.Join(landscapeDir, "landscapeTiles_010.png"),
			filepath.Join(landscapeDir, "landscapeTiles_014.png"),
			filepath.Join(landscapeDir, "landscapeTiles_002.png"),
		},
		"snow": {
			filepath.Join(landscapeDir, "landscapeTiles_010.png"),
			filepath.Join(landscapeDir, "landscapeTiles_014.png"),
			filepath.Join(landscapeDir, "landscapeTiles_002.png"),
		},
		"forest": {
			filepath.Join(landscapeDir, "landscapeTiles_010.png"),
			filepath.Join(landscapeDir, "landscapeTiles_014.png"),
			filepath.Join(landscapeDir, "landscapeTiles_002.png"),
		},
	}

	// Road tiles from Kenney Isometric Roads
	roadFiles := []string{
		filepath.Join(roadsDir, "roadNS.png"),
		filepath.Join(roadsDir, "roadEW.png"),
		filepath.Join(roadsDir, "crossroad.png"),
	}
	// Fallback if road files don't exist - use landscape
	for _, rf := range roadFiles {
		if _, err := os.Stat(rf); err != nil {
			roadFiles = []string{
				filepath.Join(landscapeDir, "landscapeTiles_090.png"),
				filepath.Join(landscapeDir, "landscapeTiles_010.png"),
				filepath.Join(landscapeDir, "landscapeTiles_014.png"),
			}
			break
		}
	}
	terrainMap["road"] = roadFiles

	// Bridge/Concrete from roads pack
	terrainMap["bridge"] = []string{
		filepath.Join(roadsDir, "bridgeEW.png"),
		filepath.Join(roadsDir, "bridgeNS.png"),
		filepath.Join(roadsDir, "bridgeEW.png"),
	}

	// Concrete - use road tiles
	terrainMap["concrete"] = roadFiles

	// Urban from city pack - use road tiles
	terrainMap["urban"] = roadFiles

	// Ore/Gem - tint grass tiles
	terrainMap["ore"] = terrainMap["grass"]
	terrainMap["gem"] = terrainMap["grass"]

	// Cliff
	terrainMap["cliff"] = terrainMap["rock"]

	for name, sources := range terrainMap {
		for i, src := range sources {
			dst := filepath.Join(tilesDir, fmt.Sprintf("%s_%d.png", name, i))
			if _, err := os.Stat(src); err != nil {
				fmt.Printf("  ⚠️  Missing: %s\n", src)
				continue
			}
			err := copyAndResize(src, dst, tw, th)
			if err != nil {
				fmt.Printf("  ❌ Error processing %s: %v\n", name, err)
			} else {
				fmt.Printf("  ✅ %s_%d.png\n", name, i)
			}
		}
		// Also create default (copy of variant 0)
		defaultDst := filepath.Join(tilesDir, name+".png")
		if len(sources) > 0 {
			if _, err := os.Stat(sources[0]); err == nil {
				copyAndResize(sources[0], defaultDst, tw, th)
			}
		}
	}

	// Post-process: tint ore tiles gold, gem tiles cyan
	tintTiles(tilesDir, "ore", color.NRGBA{255, 200, 50, 255})
	tintTiles(tilesDir, "gem", color.NRGBA{50, 200, 255, 255})
	// Tint snow white
	tintTiles(tilesDir, "snow", color.NRGBA{220, 230, 255, 255})
	// Tint rock gray
	tintTiles(tilesDir, "rock", color.NRGBA{160, 160, 170, 255})
	// Tint cliff darker
	tintTiles(tilesDir, "cliff", color.NRGBA{120, 120, 130, 255})
	// Tint deep_water darker blue
	tintTiles(tilesDir, "deep_water", color.NRGBA{30, 60, 180, 255})
	// Tint forest darker green
	tintTiles(tilesDir, "forest", color.NRGBA{20, 120, 40, 255})
}

// tintTiles applies a color tint to all variants of a terrain type
func tintTiles(tilesDir, name string, tint color.NRGBA) {
	for i := 0; i < 4; i++ {
		suffix := fmt.Sprintf("_%d.png", i)
		if i == 3 {
			suffix = ".png" // default
		}
		path := filepath.Join(tilesDir, name+suffix)
		img, err := loadPNG(path)
		if err != nil {
			continue
		}
		bounds := img.Bounds()
		tinted := image.NewNRGBA(bounds)

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				if a == 0 {
					continue
				}
				// Blend with tint (multiply)
				nr := uint8((float64(r>>8) * float64(tint.R) / 255))
				ng := uint8((float64(g>>8) * float64(tint.G) / 255))
				nb := uint8((float64(b>>8) * float64(tint.B) / 255))
				na := uint8(a >> 8)
				tinted.SetNRGBA(x, y, color.NRGBA{nr, ng, nb, na})
			}
		}
		savePNG(path, tinted)
	}
}

func processBuildingSprites(dlDir string) {
	fmt.Println("\n🏗️  Processing building sprites...")
	spritesDir := filepath.Join(baseDir, "assets", "sprites")
	cityDir := filepath.Join(dlDir, "kenney_isometric_city", "PNG")

	// Kenney City Tiles mapping to RTS buildings:
	// Building tiles are the taller ones (height > 100)
	// 000 = road with small building, 004 = building, 006 = road+building
	// 011 = factory-like, 012 = bed/small, 013 = foundation (construction yard)
	// 023 = building, 024 = foundation, 026 = tunnel (turret)
	// 031 = building (wall), 033 = bunker (turret), 034 = terrain, 036 = road+trees

	buildingMap := map[string]struct {
		src string
		w, h int
	}{
		"construction_yard": {filepath.Join(cityDir, "cityTiles_013.png"), 160, 160},
		"power_plant":       {filepath.Join(cityDir, "cityTiles_011.png"), 140, 140},
		"barracks":          {filepath.Join(cityDir, "cityTiles_023.png"), 140, 140},
		"war_factory":       {filepath.Join(cityDir, "cityTiles_018.png"), 160, 160},
		"refinery":          {filepath.Join(cityDir, "cityTiles_024.png"), 150, 150},
		"radar":             {filepath.Join(cityDir, "cityTiles_031.png"), 140, 140},
		"turret":            {filepath.Join(cityDir, "cityTiles_033.png"), 100, 100},
		"wall":              {filepath.Join(cityDir, "cityTiles_026.png"), 80, 80},
	}

	factions := []string{"allied", "soviet"}
	tints := map[string]color.NRGBA{
		"allied": {100, 150, 255, 255}, // Blue tint
		"soviet": {255, 100, 100, 255}, // Red tint
	}

	for name, info := range buildingMap {
		if _, err := os.Stat(info.src); err != nil {
			fmt.Printf("  ⚠️  Missing: %s\n", info.src)
			continue
		}

		// Default (no tint)
		dst := filepath.Join(spritesDir, name+".png")
		copyAndResize(info.src, dst, info.w, info.h)
		fmt.Printf("  ✅ %s.png\n", name)

		// Faction variants
		for _, faction := range factions {
			fDst := filepath.Join(spritesDir, fmt.Sprintf("%s_%s.png", name, faction))
			copyAndResize(info.src, fDst, info.w, info.h)

			// Apply faction tint
			img, _ := loadPNG(fDst)
			if img != nil {
				tint := tints[faction]
				tinted := applyTint(img, tint, 0.3) // 30% tint blend
				savePNG(fDst, tinted)
			}
			fmt.Printf("  ✅ %s_%s.png\n", name, faction)

			// Construction stages (progressively more transparent)
			for stage := 0; stage < 3; stage++ {
				sDst := filepath.Join(spritesDir, fmt.Sprintf("%s_%s_build_%d.png", name, faction, stage))
				copyAndResize(info.src, sDst, info.w, info.h)
				img, _ := loadPNG(sDst)
				if img != nil {
					opacity := 0.3 + float64(stage)*0.25 // 30%, 55%, 80%
					tinted := applyTint(img, tints[faction], 0.3)
					faded := applyOpacity(tinted, opacity)
					savePNG(sDst, faded)
				}
			}

			// Damaged version (darker + red tint)
			dDst := filepath.Join(spritesDir, fmt.Sprintf("%s_%s_damaged.png", name, faction))
			copyAndResize(info.src, dDst, info.w, info.h)
			img, _ = loadPNG(dDst)
			if img != nil {
				tinted := applyTint(img, color.NRGBA{200, 80, 80, 255}, 0.4)
				darkened := applyBrightness(tinted, 0.6)
				savePNG(dDst, darkened)
			}
		}
	}
}

func processUnitSprites(dlDir string) {
	fmt.Println("\n🎖️  Processing unit sprites...")
	spritesDir := filepath.Join(baseDir, "assets", "sprites")
	tanksDir := filepath.Join(dlDir, "kenney_tanks", "PNG", "Default size")

	// Kenney tanks are top-down, we'll use them directly (they work well for RTS)
	// tank body + turret compositing for 8 directions

	unitMap := map[string]struct {
		body    string
		turret  string
		w, h    int
	}{
		"tank":            {"tanks_tankGreen_body1.png", "tanks_turret1.png", 48, 48},
		"harvester":       {"tanks_tankDesert_body2.png", "", 48, 48},
		"mcv":             {"tanks_tankGrey_body3.png", "", 64, 64},
		"apocalypse_tank": {"tanks_tankNavy_body4.png", "tanks_turret3.png", 56, 56},
		"v3_rocket":       {"tanks_tankDesert_body5.png", "tanks_turret4.png", 48, 48},
	}

	for name, info := range unitMap {
		bodyPath := filepath.Join(tanksDir, info.body)
		if _, err := os.Stat(bodyPath); err != nil {
			fmt.Printf("  ⚠️  Missing body: %s\n", bodyPath)
			// Generate a fallback colored rectangle
			generateFallbackUnit(spritesDir, name, info.w, info.h)
			continue
		}

		bodyImg, _ := loadPNG(bodyPath)
		var turretImg image.Image
		if info.turret != "" {
			turretPath := filepath.Join(tanksDir, info.turret)
			turretImg, _ = loadPNG(turretPath)
		}

		// Generate 8 directions x 3 frames
		for dir := 0; dir < 8; dir++ {
			angle := float64(dir) * math.Pi / 4.0
			for frame := 0; frame < 3; frame++ {
				// Slight offset for animation frames
				frameAngle := angle + float64(frame)*0.02

				rotated := rotateImage(bodyImg, frameAngle, info.w, info.h)

				// Composite turret if available
				if turretImg != nil {
					turretRot := rotateImage(turretImg, frameAngle, info.w, info.h)
					draw.Draw(rotated, rotated.Bounds(), turretRot, image.Point{}, draw.Over)
				}

				dst := filepath.Join(spritesDir, fmt.Sprintf("%s_d%d_f%d.png", name, dir, frame))
				savePNG(dst, rotated)
			}
		}

		// Default sprite (direction 2 = south)
		defaultDst := filepath.Join(spritesDir, name+".png")
		angle := math.Pi / 2.0 // South
		rotated := rotateImage(bodyImg, angle, info.w, info.h)
		if turretImg != nil {
			turretRot := rotateImage(turretImg, angle, info.w, info.h)
			draw.Draw(rotated, rotated.Bounds(), turretRot, image.Point{}, draw.Over)
		}
		savePNG(defaultDst, rotated)
		fmt.Printf("  ✅ %s (8 dirs × 3 frames)\n", name)
	}

	// Infantry and other non-tank units: generate colored sprites
	infantryUnits := map[string]color.NRGBA{
		"infantry":    {80, 120, 60, 255},  // Green camo
		"engineer":    {200, 180, 50, 255}, // Yellow
		"attack_dog":  {139, 90, 43, 255},  // Brown
	}

	for name, clr := range infantryUnits {
		generateInfantryUnit(spritesDir, name, 24, 32, clr)
		fmt.Printf("  ✅ %s (8 dirs × 3 frames)\n", name)
	}
}

func processEffects(dlDir string) {
	fmt.Println("\n💥 Processing effects...")
	effectsDir := filepath.Join(baseDir, "assets", "effects")
	smokeDir := filepath.Join(dlDir, "kenney_smoke_particles", "PNG")
	particleDir := filepath.Join(dlDir, "kenney_particle_pack", "PNG (Transparent)")

	// Explosions from smoke particles
	explosionDir := filepath.Join(smokeDir, "Explosion")
	for i := 0; i < 8; i++ {
		src := filepath.Join(explosionDir, fmt.Sprintf("explosion%02d.png", i))
		dst := filepath.Join(effectsDir, fmt.Sprintf("explosion_%d.png", i))
		if _, err := os.Stat(src); err == nil {
			copyAndResize(src, dst, 64, 64)
			fmt.Printf("  ✅ explosion_%d.png\n", i)
		}
	}

	// Muzzle flash from particle pack
	flashDir := filepath.Join(smokeDir, "Flash")
	for i := 0; i < 3; i++ {
		src := filepath.Join(flashDir, fmt.Sprintf("flash%02d.png", i))
		dst := filepath.Join(effectsDir, fmt.Sprintf("muzzle_%d.png", i))
		if _, err := os.Stat(src); err == nil {
			copyAndResize(src, dst, 32, 32)
			fmt.Printf("  ✅ muzzle_%d.png\n", i)
		}
	}

	// Smoke from smoke particles
	blackSmokeDir := filepath.Join(smokeDir, "Black smoke")
	for i := 0; i < 4; i++ {
		src := filepath.Join(blackSmokeDir, fmt.Sprintf("blackSmoke%02d.png", i))
		dst := filepath.Join(effectsDir, fmt.Sprintf("smoke_%d.png", i))
		if _, err := os.Stat(src); err == nil {
			copyAndResize(src, dst, 48, 48)
			fmt.Printf("  ✅ smoke_%d.png\n", i)
		}
	}

	// Ore sparkle from particle pack
	for i := 0; i < 4; i++ {
		src := filepath.Join(particleDir, fmt.Sprintf("flare_0%d.png", i+1))
		if _, err := os.Stat(src); err != nil {
			// Try star pattern
			src = filepath.Join(particleDir, fmt.Sprintf("star_0%d.png", i+1))
		}
		dst := filepath.Join(effectsDir, fmt.Sprintf("ore_sparkle_%d.png", i))
		if _, err := os.Stat(src); err == nil {
			copyAndResize(src, dst, 24, 24)
			fmt.Printf("  ✅ ore_sparkle_%d.png\n", i)
		} else {
			// Generate fallback sparkle
			generateSparkle(dst, 24, i)
		}
	}

	// Selection circle and rally flag (generate)
	generateSelectionCircle(filepath.Join(effectsDir, "selection_circle.png"), 64)
	generateRallyFlag(filepath.Join(effectsDir, "rally_flag.png"), 32, 48)
	fmt.Println("  ✅ selection_circle.png")
	fmt.Println("  ✅ rally_flag.png")
}

// rotateImage rotates an image by angle (radians) and fits into w×h
func rotateImage(src image.Image, angle float64, w, h int) *image.NRGBA {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	dst := image.NewNRGBA(image.Rect(0, 0, w, h))

	cx := float64(w) / 2
	cy := float64(h) / 2
	srcCx := float64(srcW) / 2
	srcCy := float64(srcH) / 2

	cosA := math.Cos(-angle)
	sinA := math.Sin(-angle)

	// Scale factor to fit source into destination
	scale := math.Min(float64(w)/float64(srcW), float64(h)/float64(srcH))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Translate to center, rotate, scale, translate back
			dx := float64(x) - cx
			dy := float64(y) - cy

			// Inverse rotation
			rx := dx*cosA - dy*sinA
			ry := dx*sinA + dy*cosA

			// Scale
			rx /= scale
			ry /= scale

			// Translate to source center
			sx := int(rx + srcCx)
			sy := int(ry + srcCy)

			if sx >= 0 && sx < srcW && sy >= 0 && sy < srcH {
				dst.Set(x, y, src.At(srcBounds.Min.X+sx, srcBounds.Min.Y+sy))
			}
		}
	}
	return dst
}

// applyTint blends an image with a tint color at given strength (0-1)
func applyTint(src image.Image, tint color.NRGBA, strength float64) *image.NRGBA {
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)
	inv := 1.0 - strength

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			nr := uint8(float64(r>>8)*inv + float64(tint.R)*strength)
			ng := uint8(float64(g>>8)*inv + float64(tint.G)*strength)
			nb := uint8(float64(b>>8)*inv + float64(tint.B)*strength)
			dst.SetNRGBA(x, y, color.NRGBA{nr, ng, nb, uint8(a >> 8)})
		}
	}
	return dst
}

// applyOpacity multiplies alpha by opacity factor
func applyOpacity(src image.Image, opacity float64) *image.NRGBA {
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			na := uint8(float64(a>>8) * opacity)
			dst.SetNRGBA(x, y, color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), na})
		}
	}
	return dst
}

// applyBrightness multiplies RGB by factor
func applyBrightness(src image.Image, factor float64) *image.NRGBA {
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			nr := uint8(math.Min(float64(r>>8)*factor, 255))
			ng := uint8(math.Min(float64(g>>8)*factor, 255))
			nb := uint8(math.Min(float64(b>>8)*factor, 255))
			dst.SetNRGBA(x, y, color.NRGBA{nr, ng, nb, uint8(a >> 8)})
		}
	}
	return dst
}

func generateFallbackUnit(dir, name string, w, h int) {
	clr := color.NRGBA{100, 100, 100, 255}
	for dir2 := 0; dir2 < 8; dir2++ {
		for frame := 0; frame < 3; frame++ {
			img := image.NewNRGBA(image.Rect(0, 0, w, h))
			// Draw a diamond shape
			cx, cy := w/2, h/2
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					dx := math.Abs(float64(x-cx)) / float64(w/2)
					dy := math.Abs(float64(y-cy)) / float64(h/2)
					if dx+dy < 1.0 {
						img.SetNRGBA(x, y, clr)
					}
				}
			}
			dst := filepath.Join(dir, fmt.Sprintf("%s_d%d_f%d.png", name, dir2, frame))
			savePNG(dst, img)
		}
	}
	// Default
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	cx, cy := w/2, h/2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := math.Abs(float64(x-cx)) / float64(w/2)
			dy := math.Abs(float64(y-cy)) / float64(h/2)
			if dx+dy < 1.0 {
				img.SetNRGBA(x, y, clr)
			}
		}
	}
	savePNG(filepath.Join(dir, name+".png"), img)
}

func generateInfantryUnit(dir, name string, w, h int, clr color.NRGBA) {
	for d := 0; d < 8; d++ {
		for frame := 0; frame < 3; frame++ {
			img := image.NewNRGBA(image.Rect(0, 0, w, h))
			// Draw a simple humanoid shape
			cx := w / 2
			// Head (circle)
			headR := w / 5
			headY := h / 4
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					dx := float64(x - cx)
					dy := float64(y - headY)
					// Head
					if dx*dx+dy*dy < float64(headR*headR) {
						skin := color.NRGBA{200, 170, 140, 255}
						img.SetNRGBA(x, y, skin)
						continue
					}
					// Body (rectangle below head)
					bodyTop := headY + headR
					bodyBot := h * 3 / 4
					bodyW := w / 3
					if y >= bodyTop && y <= bodyBot && x >= cx-bodyW && x <= cx+bodyW {
						// Animate: slight sway
						offset := 0
						if frame == 1 {
							offset = 1
						} else if frame == 2 {
							offset = -1
						}
						if x+offset >= cx-bodyW && x+offset <= cx+bodyW {
							img.SetNRGBA(x, y, clr)
						}
						continue
					}
					// Legs
					legTop := bodyBot
					legBot := h - 1
					legW := w / 6
					// Left leg
					legOffset := 0
					if frame == 1 { legOffset = 2 }
					if frame == 2 { legOffset = -2 }
					if y >= legTop && y <= legBot {
						if (x >= cx-bodyW && x <= cx-bodyW+legW) {
							if y+legOffset >= legTop {
								img.SetNRGBA(x, y, clr)
							}
						}
						if (x >= cx+bodyW-legW && x <= cx+bodyW) {
							if y-legOffset >= legTop {
								img.SetNRGBA(x, y, clr)
							}
						}
					}
				}
			}

			_ = d // direction doesn't change shape much for tiny infantry
			dst := filepath.Join(dir, fmt.Sprintf("%s_d%d_f%d.png", name, d, frame))
			savePNG(dst, img)
		}
	}
	// Default
	defaultPath := filepath.Join(dir, name+".png")
	src := filepath.Join(dir, fmt.Sprintf("%s_d2_f0.png", name))
	if img, err := loadPNG(src); err == nil {
		savePNG(defaultPath, img)
	}
}

func generateSparkle(path string, size, variant int) {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	cx, cy := size/2, size/2

	colors := []color.NRGBA{
		{255, 223, 0, 255},   // Gold
		{255, 200, 50, 255},  // Amber
		{255, 255, 100, 255}, // Light gold
		{200, 180, 0, 255},   // Dark gold
	}
	clr := colors[variant%len(colors)]

	// Draw a cross/star shape
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := math.Abs(float64(x - cx))
			dy := math.Abs(float64(y - cy))
			dist := math.Sqrt(dx*dx + dy*dy)
			maxDist := float64(size) / 2

			// Star pattern
			if (dx < 2 || dy < 2) && dist < maxDist*0.8 {
				alpha := uint8(255 * (1.0 - dist/maxDist))
				img.SetNRGBA(x, y, color.NRGBA{clr.R, clr.G, clr.B, alpha})
			}
		}
	}
	savePNG(path, img)
}

func generateSelectionCircle(path string, size int) {
	img := image.NewNRGBA(image.Rect(0, 0, size, size/2))
	cx := float64(size) / 2
	cy := float64(size) / 4
	rx := float64(size)/2 - 2
	ry := float64(size)/4 - 2

	for y := 0; y < size/2; y++ {
		for x := 0; x < size; x++ {
			dx := (float64(x) - cx) / rx
			dy := (float64(y) - cy) / ry
			dist := dx*dx + dy*dy

			if dist > 0.85 && dist < 1.15 {
				alpha := uint8(200 * (1.0 - math.Abs(dist-1.0)/0.15))
				img.SetNRGBA(x, y, color.NRGBA{0, 255, 0, alpha})
			}
		}
	}
	savePNG(path, img)
}

func generateRallyFlag(path string, w, h int) {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	// Pole
	poleX := w / 3
	for y := 0; y < h; y++ {
		img.SetNRGBA(poleX, y, color.NRGBA{139, 90, 43, 255})
		img.SetNRGBA(poleX+1, y, color.NRGBA{139, 90, 43, 255})
	}

	// Flag
	flagW := w * 2 / 3
	flagH := h / 3
	for y := 2; y < 2+flagH; y++ {
		for x := poleX + 2; x < poleX+2+flagW && x < w; x++ {
			wave := int(2 * math.Sin(float64(x-poleX)*0.3))
			fy := y + wave
			if fy >= 0 && fy < h {
				img.SetNRGBA(x, fy, color.NRGBA{255, 50, 50, 255})
			}
		}
	}
	savePNG(path, img)
}

func init() {
	// Suppress unused import warning
	_ = strings.Contains
}

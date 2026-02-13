package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
	"image/png"
)

// resizePNG loads a PNG, resizes to target dimensions, and saves
func resizePNG(src, dst string, tw, th int) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	srcImg, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	dstImg := image.NewRGBA(image.Rect(0, 0, tw, th))
	xdraw.CatmullRom.Scale(dstImg, dstImg.Bounds(), srcImg, srcImg.Bounds(), xdraw.Over, nil)

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, dstImg)
}

// copyPNG copies a PNG as-is
func copyPNG(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func main() {
	dlBase := "/tmp/assets_dl"
	assetsBase := "assets"
	tilesDir := filepath.Join(assetsBase, "tiles")
	spritesDir := filepath.Join(assetsBase, "sprites")
	effectsDir := filepath.Join(assetsBase, "effects")
	os.MkdirAll(tilesDir, 0755)
	os.MkdirAll(spritesDir, 0755)
	os.MkdirAll(effectsDir, 0755)

	tw, th := 128, 64

	// ===== TERRAIN TILES from Kenney Isometric Roads =====
	roadsDir := filepath.Join(dlBase, "kenney_isometric_roads", "PNG")
	
	// Direct terrain mappings (roads pack has named terrain files)
	terrainMap := map[string]string{
		"grass":   "grass.png",
		"dirt":    "dirt.png", 
		"water":   "water.png",
		"road":    "road.png",
		"bridge":  "bridgeEW.png",
	}
	for name, srcFile := range terrainMap {
		src := filepath.Join(roadsDir, srcFile)
		if _, err := os.Stat(src); err != nil {
			fmt.Printf("  ⚠ Missing: %s\n", src)
			continue
		}
		// Main + 3 variants
		for v := 0; v < 3; v++ {
			dst := filepath.Join(tilesDir, fmt.Sprintf("%s_%d.png", name, v))
			if err := resizePNG(src, dst, tw, th); err != nil {
				fmt.Printf("  ✗ %s: %v\n", dst, err)
			} else {
				fmt.Printf("  → %s\n", dst)
			}
		}
		dst := filepath.Join(tilesDir, name+".png")
		resizePNG(src, dst, tw, th)
		fmt.Printf("  → %s\n", dst)
	}

	// ===== LANDSCAPE TILES for other terrain types =====
	landDir := filepath.Join(dlBase, "kenney_isometric_landscape", "PNG")
	// These mappings are based on Kenney isometric landscape numbering
	landMap := map[string][]string{
		"sand":       {"landscapeTiles_036.png", "landscapeTiles_037.png", "landscapeTiles_038.png"},
		"snow":       {"landscapeTiles_048.png", "landscapeTiles_049.png", "landscapeTiles_050.png"},
		"rock":       {"landscapeTiles_060.png", "landscapeTiles_061.png", "landscapeTiles_062.png"},
		"deep_water": {"landscapeTiles_072.png", "landscapeTiles_073.png", "landscapeTiles_074.png"},
		"forest":     {"landscapeTiles_024.png", "landscapeTiles_025.png", "landscapeTiles_026.png"},
		"cliff":      {"landscapeTiles_084.png", "landscapeTiles_085.png", "landscapeTiles_086.png"},
	}
	for name, srcFiles := range landMap {
		for v, srcFile := range srcFiles {
			src := filepath.Join(landDir, srcFile)
			if _, err := os.Stat(src); err != nil {
				// Try with the first file if others don't exist
				src = filepath.Join(landDir, srcFiles[0])
			}
			dst := filepath.Join(tilesDir, fmt.Sprintf("%s_%d.png", name, v))
			if err := resizePNG(src, dst, tw, th); err != nil {
				fmt.Printf("  ✗ %s: %v\n", dst, err)
			} else {
				fmt.Printf("  → %s\n", dst)
			}
		}
		resizePNG(filepath.Join(landDir, srcFiles[0]), filepath.Join(tilesDir, name+".png"), tw, th)
		fmt.Printf("  → %s\n", filepath.Join(tilesDir, name+".png"))
	}

	// ===== CITY TILES for urban/concrete =====
	cityDir := filepath.Join(dlBase, "kenney_isometric_city", "PNG")
	cityMap := map[string][]string{
		"urban":    {"cityTiles_000.png", "cityTiles_001.png", "cityTiles_002.png"},
		"concrete": {"cityTiles_003.png", "cityTiles_004.png", "cityTiles_005.png"},
		"ore":     {"cityTiles_006.png", "cityTiles_007.png", "cityTiles_008.png"},
		"gem":     {"cityTiles_009.png", "cityTiles_010.png", "cityTiles_011.png"},
	}
	for name, srcFiles := range cityMap {
		for v, srcFile := range srcFiles {
			src := filepath.Join(cityDir, srcFile)
			if _, err := os.Stat(src); err != nil {
				src = filepath.Join(cityDir, srcFiles[0])
			}
			dst := filepath.Join(tilesDir, fmt.Sprintf("%s_%d.png", name, v))
			if err := resizePNG(src, dst, tw, th); err != nil {
				fmt.Printf("  ✗ %s: %v\n", dst, err)
			} else {
				fmt.Printf("  → %s\n", dst)
			}
		}
		resizePNG(filepath.Join(cityDir, srcFiles[0]), filepath.Join(tilesDir, name+".png"), tw, th)
		fmt.Printf("  → %s\n", filepath.Join(tilesDir, name+".png"))
	}

	// ===== BUILDING SPRITES from city =====
	buildingMap := map[string]string{
		"construction_yard": "cityTiles_020.png",
		"power_plant":       "cityTiles_030.png",
		"barracks":          "cityTiles_040.png",
		"war_factory":       "cityTiles_050.png",
		"refinery":          "cityTiles_060.png",
		"radar":             "cityTiles_070.png",
	}
	for name, srcFile := range buildingMap {
		src := filepath.Join(cityDir, srcFile)
		if _, err := os.Stat(src); err != nil {
			fmt.Printf("  ⚠ Missing building: %s\n", src)
			continue
		}
		dst := filepath.Join(spritesDir, name+".png")
		if err := copyPNG(src, dst); err == nil {
			fmt.Printf("  → %s\n", dst)
		}
		// Also save faction variants
		for _, faction := range []string{"allied", "soviet"} {
			fdst := filepath.Join(spritesDir, fmt.Sprintf("%s_%s.png", name, faction))
			copyPNG(src, fdst)
			fmt.Printf("  → %s\n", fdst)
		}
	}

	// ===== TANK SPRITES =====
	tanksDir := filepath.Join(dlBase, "kenney_tanks", "PNG", "Default size")
	tankFiles, _ := filepath.Glob(filepath.Join(tanksDir, "*.png"))
	fmt.Printf("\n=== Tank files found: %d ===\n", len(tankFiles))
	for _, f := range tankFiles {
		fmt.Printf("  %s\n", filepath.Base(f))
	}

	// ===== SMOKE/EXPLOSION EFFECTS =====
	smokeDir := filepath.Join(dlBase, "kenney_smoke", "PNG")
	// Black smoke for smoke effects
	for i := 0; i < 4; i++ {
		src := filepath.Join(smokeDir, "Black smoke", fmt.Sprintf("blackSmoke%02d.png", i))
		dst := filepath.Join(effectsDir, fmt.Sprintf("smoke_%d.png", i))
		if err := resizePNG(src, dst, 32, 32); err == nil {
			fmt.Printf("  → %s\n", dst)
		}
	}
	// Explosion frames
	for i := 0; i < 8; i++ {
		src := filepath.Join(smokeDir, "Explosion", fmt.Sprintf("explosion%02d.png", i))
		dst := filepath.Join(effectsDir, fmt.Sprintf("explosion_%d.png", i))
		if err := resizePNG(src, dst, 64, 64); err == nil {
			fmt.Printf("  → %s\n", dst)
		}
	}
	// Flash for muzzle
	for i := 0; i < 3; i++ {
		src := filepath.Join(smokeDir, "Flash", fmt.Sprintf("flash%02d.png", i))
		dst := filepath.Join(effectsDir, fmt.Sprintf("muzzle_%d.png", i))
		if err := resizePNG(src, dst, 32, 32); err == nil {
			fmt.Printf("  → %s\n", dst)
		}
	}

	// ===== PARTICLES (muzzle, sparkle) =====
	particleDir := filepath.Join(dlBase, "kenney_particles", "PNG (Transparent)")
	// Muzzle flash
	for i := 1; i <= 3; i++ {
		src := filepath.Join(particleDir, fmt.Sprintf("muzzle_%02d.png", i))
		dst := filepath.Join(effectsDir, fmt.Sprintf("muzzle_%d.png", i-1))
		if err := resizePNG(src, dst, 32, 32); err == nil {
			fmt.Printf("  → %s (particle)\n", dst)
		}
	}

	fmt.Println("\n✅ Asset import complete!")
	fmt.Println("Note: Kenney assets are CC0 (public domain). Credit: kenney.nl")
	
	_ = color.RGBA{}
	_ = math.Pi
}

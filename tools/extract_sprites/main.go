// extract_sprites extracts individual sprite frames from RA1/RA2 sprite sheets
// and saves them as individual PNGs for use in the game engine.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	rawDir := "assets/ra2/raw"
	outDir := "assets/ra2"

	// Extract terrain tiles from iso_64x64_outside.png
	extractTerrainTiles(rawDir, outDir)

	// Extract RA1 unit frames
	extractRA1Units(rawDir, outDir)

	// Extract RA1 structure frames
	extractRA1Structures(rawDir, outDir)

	// Process RA2 building sprites (remove blue chroma key)
	processRA2Buildings(rawDir, outDir)

	fmt.Println("Done extracting sprites!")
}

func extractTerrainTiles(rawDir, outDir string) {
	terrainDir := filepath.Join(outDir, "terrain")
	os.MkdirAll(terrainDir, 0755)

	img := loadPNG(filepath.Join(rawDir, "iso_64x64_outside.png"))
	if img == nil {
		fmt.Println("SKIP: iso_64x64_outside.png not found")
		return
	}

	// The iso tileset has diamond-shaped tiles at 64x64
	// First 2 rows: grass variants (10 tiles each row, 64x64 grid)
	tileW, tileH := 64, 64
	names := []struct {
		name string
		x, y int
	}{
		// Row 0: grass tiles
		{"grass_1", 0, 0}, {"grass_2", 64, 0}, {"grass_3", 128, 0},
		{"grass_4", 192, 0}, {"grass_5", 256, 0}, {"grass_6", 320, 0},
		{"grass_7", 384, 0}, {"grass_8", 448, 0}, {"grass_9", 512, 0},
		// Row 1: more grass
		{"grass_dark_1", 0, 64}, {"grass_dark_2", 64, 64}, {"grass_dark_3", 128, 64},
		{"grass_dark_4", 192, 64}, {"grass_dark_5", 256, 64}, {"grass_dark_6", 320, 64},
		// Water tiles (from the water row ~row 10-11, around y=640-704)
		{"water_1", 0, 640}, {"water_2", 64, 640}, {"water_3", 128, 640},
		{"water_4", 192, 640}, {"water_5", 256, 640},
		{"water_6", 0, 704}, {"water_7", 64, 704}, {"water_8", 128, 704},
	}

	for _, t := range names {
		sub := cropImage(img, t.x, t.y, tileW, tileH)
		savePNG(filepath.Join(terrainDir, t.name+".png"), sub)
	}
	fmt.Printf("Extracted %d terrain tiles\n", len(names))
}

func extractRA1Units(rawDir, outDir string) {
	unitsDir := filepath.Join(outDir, "units")
	os.MkdirAll(unitsDir, 0755)

	// ra1_units_1.png: Allied tanks (top) and Soviet tanks (bottom), 8 rotation frames per row
	// 297x593, roughly 8 columns of ~37px each, ~16 rows of ~37px
	img := loadPNG(filepath.Join(rawDir, "ra1_units_1.png"))
	if img == nil {
		fmt.Println("SKIP: ra1_units_1.png not found")
		return
	}

	bounds := img.Bounds()
	frameW := bounds.Dx() / 8 // 8 rotation frames per row
	// Extract first frame of each row as the "default" facing
	rows := bounds.Dy() / frameW
	for r := 0; r < rows && r < 12; r++ {
		// Take the south-facing frame (frame 0)
		sub := cropImage(img, 0, r*frameW, frameW, frameW)
		name := fmt.Sprintf("unit_ra1_%d_f0", r)
		if r < 4 {
			name = fmt.Sprintf("allied_tank_r%d", r)
		} else if r >= 4 && r < 8 {
			name = fmt.Sprintf("allied_heli_r%d", r-4)
		} else {
			name = fmt.Sprintf("soviet_tank_r%d", r-8)
		}
		savePNG(filepath.Join(unitsDir, name+".png"), sub)
	}

	// ra1_units_3.png: Vehicles with rotations (870x732)
	// Allied (blue) top half, Soviet (red) bottom half
	img3 := loadPNG(filepath.Join(rawDir, "ra1_units_3.png"))
	if img3 != nil {
		b := img3.Bounds()
		// Extract whole-sheet vehicle sprites as individual large tiles
		// These are ~58x49 per frame, 15 cols x ~15 rows
		fW, fH := b.Dx()/15, b.Dy()/15
		if fW < 30 {
			fW = 58
		}
		if fH < 30 {
			fH = 49
		}
		// Extract key frames: row 0 = allied medium tank facing south
		for col := 0; col < 8; col++ {
			sub := cropImage(img3, col*fW, 0, fW, fH)
			savePNG(filepath.Join(unitsDir, fmt.Sprintf("allied_vehicle_f%d.png", col)), sub)
		}
		// Soviet vehicles (bottom half)
		midY := b.Dy() / 2
		for col := 0; col < 8; col++ {
			sub := cropImage(img3, col*fW, midY, fW, fH)
			savePNG(filepath.Join(unitsDir, fmt.Sprintf("soviet_vehicle_f%d.png", col)), sub)
		}
		fmt.Println("Extracted RA1 vehicle frames")
	}

	// ra1_units_5.png: More unit sprites
	img5 := loadPNG(filepath.Join(rawDir, "ra1_units_5.png"))
	if img5 != nil {
		// Save as-is for now (infantry/misc units)
		savePNG(filepath.Join(unitsDir, "ra1_misc_units.png"), img5)
		fmt.Println("Copied RA1 misc units")
	}

	fmt.Println("Extracted RA1 unit sprites")
}

func extractRA1Structures(rawDir, outDir string) {
	bldgDir := filepath.Join(outDir, "buildings")
	os.MkdirAll(bldgDir, 0755)

	// ra1_structures_2.png: War Factory build-up animations (1171x881)
	// Allied (blue) top half, Soviet (red) bottom half
	img := loadPNG(filepath.Join(rawDir, "ra1_structures_2.png"))
	if img == nil {
		fmt.Println("SKIP: ra1_structures_2.png not found")
		return
	}

	b := img.Bounds()
	// Each structure frame is approximately 78x73 (15 cols, 12 rows)
	fW := b.Dx() / 15
	fH := b.Dy() / 12

	// Extract first frame of key buildings
	// Row 0: small structures build up (Allied)
	// Row 2-5: War Factory build up (Allied)
	// Bottom half: Soviet versions

	// Allied War Factory - last frame of build animation (fully built)
	// Row 4 or 5, last columns tend to be the completed building
	for row := 0; row < 6; row++ {
		// Get the last (rightmost non-empty) frame in this row
		lastCol := 14
		for lastCol > 0 {
			sub := cropImage(img, lastCol*fW, row*fH, fW, fH)
			if !isEmptyImage(sub) {
				break
			}
			lastCol--
		}
		sub := cropImage(img, lastCol*fW, row*fH, fW, fH)
		savePNG(filepath.Join(bldgDir, fmt.Sprintf("allied_structure_r%d.png", row)), sub)
	}

	// Soviet structures (bottom half)
	midRow := 6
	for row := midRow; row < 12; row++ {
		lastCol := 14
		for lastCol > 0 {
			sub := cropImage(img, lastCol*fW, row*fH, fW, fH)
			if !isEmptyImage(sub) {
				break
			}
			lastCol--
		}
		sub := cropImage(img, lastCol*fW, row*fH, fW, fH)
		savePNG(filepath.Join(bldgDir, fmt.Sprintf("soviet_structure_r%d.png", row-midRow)), sub)
	}

	// ra1_structures_1.png: Smaller structures in a row (877x147)
	img1 := loadPNG(filepath.Join(rawDir, "ra1_structures_1.png"))
	if img1 != nil {
		savePNG(filepath.Join(bldgDir, "ra1_small_structures.png"), img1)
	}

	fmt.Println("Extracted RA1 structure sprites")
}

func processRA2Buildings(rawDir, outDir string) {
	bldgDir := filepath.Join(outDir, "buildings")
	os.MkdirAll(bldgDir, 0755)

	// Soviet barracks has blue chroma key background (RGB 0,0,255 and similar)
	for _, entry := range []struct {
		src, dst string
	}{
		{"soviet_barracks.png", "ra2_soviet_barracks.png"},
		{"soviet_battle_lab.png", "ra2_soviet_battle_lab.png"},
		{"soviet_cloning_vat.png", "ra2_soviet_cloning_vat.png"},
		{"soviet_tesla_coil.png", "ra2_soviet_tesla_coil.png"},
		{"soviet_iron_curtain.png", "ra2_soviet_iron_curtain.png"},
	} {
		img := loadPNG(filepath.Join(rawDir, entry.src))
		if img == nil {
			continue
		}
		// Remove blue/teal chroma key
		cleaned := removeChromaKey(img)
		savePNG(filepath.Join(bldgDir, entry.dst), cleaned)
		fmt.Printf("Processed %s -> %s\n", entry.src, entry.dst)
	}

	// Allied buildings (already have transparency in the RA2 sheets)
	for _, entry := range []struct {
		src, dst string
	}{
		{"allied_chronosphere.png", "ra2_allied_chronosphere.png"},
		{"allied_buildings_2.png", "ra2_allied_buildings.png"},
		{"allied_buildings_3.png", "ra2_allied_misc.png"},
	} {
		img := loadPNG(filepath.Join(rawDir, entry.src))
		if img == nil {
			continue
		}
		savePNG(filepath.Join(bldgDir, entry.dst), img)
		fmt.Printf("Copied %s -> %s\n", entry.src, entry.dst)
	}
}

// removeChromaKey removes blue (0,0,255) and teal (0,128,128) chroma key backgrounds
func removeChromaKey(img image.Image) *image.RGBA {
	b := img.Bounds()
	out := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(bl>>8)

			// Blue chroma key: pure blue or near-blue
			if b8 > 200 && r8 < 30 && g8 < 30 {
				out.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
				continue
			}
			// Teal chroma key (0, ~128, ~128)
			if r8 < 20 && g8 > 100 && g8 < 160 && b8 > 100 && b8 < 160 {
				out.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
				continue
			}

			out.Set(x, y, img.At(x, y))
		}
	}
	return out
}

func loadPNG(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		// Try as generic image
		f.Seek(0, 0)
		img, _, err = image.Decode(f)
		if err != nil {
			fmt.Printf("Failed to decode %s: %v\n", path, err)
			return nil
		}
	}
	return img
}

func cropImage(img image.Image, x, y, w, h int) image.Image {
	b := img.Bounds()
	if x+w > b.Max.X {
		w = b.Max.X - x
	}
	if y+h > b.Max.Y {
		h = b.Max.Y - y
	}
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			out.Set(dx, dy, img.At(x+dx, y+dy))
		}
	}
	return out
}

func isEmptyImage(img image.Image) bool {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				return false
			}
		}
	}
	return true
}

func savePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("Failed to create %s: %v\n", path, err)
		return
	}
	defer f.Close()
	png.Encode(f, img)
}

package render3d

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteAtlas manages loaded sprite images for billboard rendering
type SpriteAtlas struct {
	sprites  map[string]*ebiten.Image
	basePath string
	loaded   bool
}

// NewSpriteAtlas creates a new sprite atlas
func NewSpriteAtlas() *SpriteAtlas {
	return &SpriteAtlas{
		sprites: make(map[string]*ebiten.Image),
	}
}

// LoadFromDirectory loads all PNG files from assets/ra2/ subdirectories
func (sa *SpriteAtlas) LoadFromDirectory(basePath string) {
	sa.basePath = basePath

	dirs := []string{"buildings", "units", "terrain", "effects"}
	total := 0
	for _, dir := range dirs {
		dirPath := filepath.Join(basePath, dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".png" {
				continue
			}
			key := dir + "/" + e.Name()[:len(e.Name())-4] // e.g. "buildings/allied_structure_r4"
			img := loadEbitenImage(filepath.Join(dirPath, e.Name()))
			if img != nil {
				sa.sprites[key] = img
				total++
			}
		}
	}
	if total > 0 {
		sa.loaded = true
		fmt.Printf("SpriteAtlas: loaded %d sprites from %s\n", total, basePath)
	}
}

// Get returns a sprite by key, or nil
func (sa *SpriteAtlas) Get(key string) *ebiten.Image {
	return sa.sprites[key]
}

// Has returns true if the sprite exists
func (sa *SpriteAtlas) Has(key string) bool {
	_, ok := sa.sprites[key]
	return ok
}

// IsLoaded returns true if any sprites were loaded
func (sa *SpriteAtlas) IsLoaded() bool {
	return sa.loaded
}

// GetBuildingSprite returns the sprite for a building key and faction
func (sa *SpriteAtlas) GetBuildingSprite(buildingKey, faction string) *ebiten.Image {
	// Try RA2 sprites first
	if faction == "Soviet" || faction == "soviet" {
		if img := sa.Get("buildings/ra2_soviet_" + buildingKey); img != nil {
			return img
		}
	}

	// Try RA1 structure sprites mapped to building keys
	// Map building keys to extracted sprite names
	switch buildingKey {
	case "construction_yard":
		if faction == "Soviet" || faction == "soviet" {
			return sa.Get("buildings/soviet_structure_r4")
		}
		return sa.Get("buildings/allied_structure_r4")
	case "war_factory":
		if faction == "Soviet" || faction == "soviet" {
			return sa.Get("buildings/soviet_structure_r3")
		}
		return sa.Get("buildings/allied_structure_r3")
	case "barracks":
		if faction == "Soviet" || faction == "soviet" {
			return sa.Get("buildings/soviet_structure_r1")
		}
		return sa.Get("buildings/allied_structure_r1")
	case "power_plant":
		if faction == "Soviet" || faction == "soviet" {
			return sa.Get("buildings/soviet_structure_r0")
		}
		return sa.Get("buildings/allied_structure_r0")
	case "refinery":
		if faction == "Soviet" || faction == "soviet" {
			return sa.Get("buildings/soviet_structure_r2")
		}
		return sa.Get("buildings/allied_structure_r2")
	}
	return nil
}

// GetUnitSprite returns the sprite for a unit type and faction
func (sa *SpriteAtlas) GetUnitSprite(unitType, faction string) *ebiten.Image {
	switch unitType {
	case "tank":
		if faction == "Soviet" || faction == "soviet" {
			return sa.Get("units/soviet_tank_r0")
		}
		return sa.Get("units/allied_tank_r0")
	case "mcv":
		if faction == "Soviet" || faction == "soviet" {
			if img := sa.Get("units/soviet_vehicle_f0"); img != nil {
				return img
			}
		}
		if img := sa.Get("units/allied_vehicle_f0"); img != nil {
			return img
		}
	case "harvester":
		// Use a different rotation frame to distinguish from MCV
		if faction == "Soviet" || faction == "soviet" {
			if img := sa.Get("units/soviet_vehicle_f2"); img != nil {
				return img
			}
			return sa.Get("units/soviet_vehicle_f0")
		}
		if img := sa.Get("units/allied_vehicle_f2"); img != nil {
			return img
		}
		return sa.Get("units/allied_vehicle_f0")
	case "infantry":
		// Use RA2 infantry sprites
		if img := sa.Get("units/GI"); img != nil {
			return img
		}
	case "engineer":
		if img := sa.Get("units/ENGINEER"); img != nil {
			return img
		}
	case "dog":
		if img := sa.Get("units/DOG"); img != nil {
			return img
		}
	}
	// Fallback: try GI for any unmatched infantry
	if unitType == "infantry" || unitType == "" {
		return sa.Get("units/GI")
	}
	return nil
}

// DrawBillboard renders a sprite as a billboard (camera-facing quad) at a 3D position
func (sa *SpriteAtlas) DrawBillboard(screen *ebiten.Image, cam *Camera3D, sprite *ebiten.Image, worldX, worldY, worldZ, scale float64) {
	if sprite == nil {
		return
	}

	// Project world position to screen
	sx, sy, depth := cam.Project3DToScreen(worldX, worldY, worldZ)
	// In orthographic projection, clip.Z ranges [-1, 1]; only skip if outside view volume
	_ = depth

	// Scale based on distance/zoom
	imgW := float64(sprite.Bounds().Dx())
	imgH := float64(sprite.Bounds().Dy())

	// The sprite should cover approximately 'scale' world units on screen
	// Zoom = world units across screen width, so pixelsPerUnit = screenW / Zoom
	pixelsPerUnit := float64(cam.ScreenW) / cam.Zoom
	targetW := scale * pixelsPerUnit
	scaleF := targetW / imgW
	targetH := imgH * scaleF

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scaleF, scaleF)
	op.GeoM.Translate(float64(sx)-targetW/2, float64(sy)-targetH)

	screen.DrawImage(sprite, op)
}

func loadEbitenImage(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		// Try generic decode
		f.Seek(0, 0)
		img, _, err = image.Decode(f)
		if err != nil {
			return nil
		}
	}

	return ebiten.NewImageFromImage(img)
}

// FindAssetsPath tries to locate the assets/ra2 directory
func FindAssetsPath() string {
	// Try relative to executable
	candidates := []string{
		"assets/ra2",
		"../assets/ra2",
		"../../assets/ra2",
	}

	// Also try relative to source file
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		srcDir := filepath.Dir(filename)
		candidates = append(candidates,
			filepath.Join(srcDir, "../../assets/ra2"),
			filepath.Join(srcDir, "../../../assets/ra2"),
		)
	}

	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return ""
}

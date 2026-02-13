package render

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/1siamBot/rts-engine/engine/maplib"
	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteManager holds all loaded sprite images
type SpriteManager struct {
	// Terrain: map[TerrainType][variant]*ebiten.Image
	TerrainSprites  map[maplib.TerrainType][]*ebiten.Image
	// Terrain fallback (single sprite, backward compat)
	TerrainDefault  map[maplib.TerrainType]*ebiten.Image

	BuildingSprites map[string]*ebiten.Image
	UnitSprites     map[string]*ebiten.Image

	// Unit directional sprites: key = "unitname", [direction 0-7][frame 0-2]
	UnitDirSprites  map[string][8][3]*ebiten.Image

	// Effects: key = "effectname", frames
	EffectSprites   map[string][]*ebiten.Image
}

// NewSpriteManager loads all sprites from embedded assets
func NewSpriteManager() *SpriteManager {
	sm := &SpriteManager{
		TerrainSprites:  make(map[maplib.TerrainType][]*ebiten.Image),
		TerrainDefault:  make(map[maplib.TerrainType]*ebiten.Image),
		BuildingSprites: make(map[string]*ebiten.Image),
		UnitSprites:     make(map[string]*ebiten.Image),
		UnitDirSprites:  make(map[string][8][3]*ebiten.Image),
		EffectSprites:   make(map[string][]*ebiten.Image),
	}

	assetsDir := getAssetsDir()

	// Terrain tiles (with variants)
	terrainFiles := map[maplib.TerrainType]string{
		maplib.TerrainGrass:     "grass",
		maplib.TerrainDirt:      "dirt",
		maplib.TerrainSand:      "sand",
		maplib.TerrainWater:     "water",
		maplib.TerrainDeepWater: "deep_water",
		maplib.TerrainRock:      "rock",
		maplib.TerrainCliff:     "cliff",
		maplib.TerrainRoad:      "road",
		maplib.TerrainBridge:    "bridge",
		maplib.TerrainOre:       "ore",
		maplib.TerrainGem:       "gem",
		maplib.TerrainSnow:      "snow",
		maplib.TerrainUrban:     "urban",
		maplib.TerrainForest:    "forest",
	}

	for terrain, name := range terrainFiles {
		// Load variants
		var variants []*ebiten.Image
		for v := 0; v < 3; v++ {
			img := loadFromFile(filepath.Join(assetsDir, "tiles", fmt.Sprintf("%s_%d.png", name, v)))
			if img != nil {
				variants = append(variants, img)
			}
		}
		if len(variants) > 0 {
			sm.TerrainSprites[terrain] = variants
			sm.TerrainDefault[terrain] = variants[0]
		} else {
			// Fallback to non-variant file
			img := loadFromFile(filepath.Join(assetsDir, "tiles", name+".png"))
			if img != nil {
				sm.TerrainSprites[terrain] = []*ebiten.Image{img}
				sm.TerrainDefault[terrain] = img
			}
		}
	}

	// Building sprites
	buildingNames := []string{"construction_yard", "power_plant", "barracks", "war_factory", "refinery", "radar", "turret", "wall"}
	for _, name := range buildingNames {
		img := loadFromFile(filepath.Join(assetsDir, "sprites", name+".png"))
		if img != nil {
			sm.BuildingSprites[name] = img
		}
		// Also load faction variants
		for _, faction := range []string{"allied", "soviet"} {
			fimg := loadFromFile(filepath.Join(assetsDir, "sprites", fmt.Sprintf("%s_%s.png", name, faction)))
			if fimg != nil {
				sm.BuildingSprites[fmt.Sprintf("%s_%s", name, faction)] = fimg
			}
			// Construction stages
			for stage := 0; stage < 3; stage++ {
				simg := loadFromFile(filepath.Join(assetsDir, "sprites", fmt.Sprintf("%s_%s_build_%d.png", name, faction, stage)))
				if simg != nil {
					sm.BuildingSprites[fmt.Sprintf("%s_%s_build_%d", name, faction, stage)] = simg
				}
			}
			// Damaged
			dimg := loadFromFile(filepath.Join(assetsDir, "sprites", fmt.Sprintf("%s_%s_damaged.png", name, faction)))
			if dimg != nil {
				sm.BuildingSprites[fmt.Sprintf("%s_%s_damaged", name, faction)] = dimg
			}
		}
	}

	// Unit sprites (directional + default)
	unitNames := []string{"infantry", "tank", "harvester", "mcv", "engineer", "attack_dog", "apocalypse_tank", "v3_rocket"}
	for _, name := range unitNames {
		// Default sprite
		img := loadFromFile(filepath.Join(assetsDir, "sprites", name+".png"))
		if img != nil {
			sm.UnitSprites[name] = img
		}
		// Directional sprites
		var dirFrames [8][3]*ebiten.Image
		loaded := false
		for dir := 0; dir < 8; dir++ {
			for frame := 0; frame < 3; frame++ {
				dimg := loadFromFile(filepath.Join(assetsDir, "sprites", fmt.Sprintf("%s_d%d_f%d.png", name, dir, frame)))
				if dimg != nil {
					dirFrames[dir][frame] = dimg
					loaded = true
				}
			}
		}
		if loaded {
			sm.UnitDirSprites[name] = dirFrames
		}
	}

	// Effects
	effectDefs := map[string]int{
		"explosion":    8,
		"muzzle":       3,
		"smoke":        4,
		"ore_sparkle":  4,
	}
	for name, count := range effectDefs {
		var frames []*ebiten.Image
		for f := 0; f < count; f++ {
			img := loadFromFile(filepath.Join(assetsDir, "effects", fmt.Sprintf("%s_%d.png", name, f)))
			if img != nil {
				frames = append(frames, img)
			}
		}
		if len(frames) > 0 {
			sm.EffectSprites[name] = frames
		}
	}
	// Single-frame effects
	for _, name := range []string{"selection_circle", "rally_flag"} {
		img := loadFromFile(filepath.Join(assetsDir, "effects", name+".png"))
		if img != nil {
			sm.EffectSprites[name] = []*ebiten.Image{img}
		}
	}

	totalTerrain := 0
	for _, v := range sm.TerrainSprites {
		totalTerrain += len(v)
	}
	log.Printf("SpriteManager: loaded %d terrain (%d types), %d building, %d unit (%d directional), %d effects",
		totalTerrain, len(sm.TerrainSprites), len(sm.BuildingSprites), len(sm.UnitSprites), len(sm.UnitDirSprites), len(sm.EffectSprites))

	return sm
}

// GetTerrainVariant returns a deterministic variant sprite based on tile position
func (sm *SpriteManager) GetTerrainVariant(terrain maplib.TerrainType, tileX, tileY int) *ebiten.Image {
	variants, ok := sm.TerrainSprites[terrain]
	if !ok || len(variants) == 0 {
		return nil
	}
	// Deterministic hash based on position
	hash := uint(tileX*7919 + tileY*7927 + tileX*tileY*31)
	return variants[hash%uint(len(variants))]
}

// GetUnitDirectionalSprite returns a sprite for unit facing direction and animation frame
func (sm *SpriteManager) GetUnitDirectionalSprite(unitType string, direction int, frame int) *ebiten.Image {
	dirFrames, ok := sm.UnitDirSprites[unitType]
	if !ok {
		// Fallback to default
		return sm.UnitSprites[unitType]
	}
	dir := direction % 8
	if dir < 0 {
		dir += 8
	}
	f := frame % 3
	if f < 0 {
		f = 0
	}
	img := dirFrames[dir][f]
	if img != nil {
		return img
	}
	// Fallback
	return sm.UnitSprites[unitType]
}

func getAssetsDir() string {
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Join(filepath.Dir(exe), "assets")
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "..", "..", "assets")
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	if _, err := os.Stat("assets"); err == nil {
		return "assets"
	}
	return "assets"
}

func loadFromFile(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Printf("Warning: could not decode sprite %s: %v", path, err)
		return nil
	}

	return ebiten.NewImageFromImage(img)
}

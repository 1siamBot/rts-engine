package render

import (
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
	TerrainSprites  map[maplib.TerrainType]*ebiten.Image
	BuildingSprites map[string]*ebiten.Image
	UnitSprites     map[string]*ebiten.Image
}

// NewSpriteManager loads all sprites from embedded assets
func NewSpriteManager() *SpriteManager {
	sm := &SpriteManager{
		TerrainSprites:  make(map[maplib.TerrainType]*ebiten.Image),
		BuildingSprites: make(map[string]*ebiten.Image),
		UnitSprites:     make(map[string]*ebiten.Image),
	}

	assetsDir := getAssetsDir()

	// Terrain tiles
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
		img := loadFromFile(filepath.Join(assetsDir, "tiles", name+".png"))
		if img != nil {
			sm.TerrainSprites[terrain] = img
		}
	}

	// Building sprites
	for _, name := range []string{"construction_yard", "power_plant", "barracks", "war_factory", "refinery"} {
		img := loadFromFile(filepath.Join(assetsDir, "sprites", name+".png"))
		if img != nil {
			sm.BuildingSprites[name] = img
		}
	}

	// Unit sprites
	for _, name := range []string{"infantry", "tank", "harvester", "mcv"} {
		img := loadFromFile(filepath.Join(assetsDir, "sprites", name+".png"))
		if img != nil {
			sm.UnitSprites[name] = img
		}
	}

	log.Printf("SpriteManager: loaded %d terrain, %d building, %d unit sprites",
		len(sm.TerrainSprites), len(sm.BuildingSprites), len(sm.UnitSprites))

	return sm
}

func getAssetsDir() string {
	// Try relative to executable
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Join(filepath.Dir(exe), "assets")
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	// Try relative to source file
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "..", "..", "assets")
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	// Try current working directory
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

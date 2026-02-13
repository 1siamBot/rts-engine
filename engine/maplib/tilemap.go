package maplib

import (
	"encoding/json"
	"os"
)

// TerrainType defines the terrain of a tile
type TerrainType uint8

const (
	TerrainGrass TerrainType = iota
	TerrainDirt
	TerrainSand
	TerrainWater
	TerrainDeepWater
	TerrainRock
	TerrainCliff
	TerrainRoad
	TerrainBridge
	TerrainOre
	TerrainGem
	TerrainSnow
	TerrainUrban
	TerrainForest
)

// Passability flags
type PassFlag uint8

const (
	PassInfantry PassFlag = 1 << iota
	PassVehicle
	PassNaval
	PassAir
	PassAll PassFlag = PassInfantry | PassVehicle | PassNaval | PassAir
)

// Tile represents a single map tile
type Tile struct {
	Terrain    TerrainType `json:"terrain"`
	Height     int8        `json:"height"`     // elevation level (0-7)
	Passable   PassFlag    `json:"passable"`
	TileVariant uint8      `json:"variant"`    // visual variant index
	OreAmount  int         `json:"ore"`        // resource amount (0 = none)
	Occupied   bool        `json:"-"`          // runtime: building placed here
}

// TileMap represents the game map
type TileMap struct {
	Name    string `json:"name"`
	Author  string `json:"author"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Tiles   []Tile `json:"tiles"`

	// Map metadata
	StartPositions []StartPos `json:"start_positions"`
	Description    string     `json:"description"`
	MaxPlayers     int        `json:"max_players"`

	// Isometric rendering constants
	TileWidth  int `json:"tile_width"`  // pixel width of a tile (default 64)
	TileHeight int `json:"tile_height"` // pixel height of a tile (default 32)
}

// StartPos defines a player start position
type StartPos struct {
	PlayerSlot int `json:"player_slot"`
	X          int `json:"x"`
	Y          int `json:"y"`
}

// NewTileMap creates a new empty map
func NewTileMap(name string, width, height int) *TileMap {
	tm := &TileMap{
		Name:       name,
		Width:      width,
		Height:     height,
		Tiles:      make([]Tile, width*height),
		TileWidth:  128,
		TileHeight: 64,
		MaxPlayers: 2,
	}

	// Default all tiles to grass, passable by ground
	for i := range tm.Tiles {
		tm.Tiles[i] = Tile{
			Terrain:  TerrainGrass,
			Passable: PassAll,
		}
	}

	return tm
}

// At returns a pointer to the tile at (x, y)
func (tm *TileMap) At(x, y int) *Tile {
	if x < 0 || y < 0 || x >= tm.Width || y >= tm.Height {
		return nil
	}
	return &tm.Tiles[y*tm.Width+x]
}

// InBounds checks if coordinates are within map bounds
func (tm *TileMap) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < tm.Width && y < tm.Height
}

// IsPassable checks if a tile can be traversed by a given movement type
func (tm *TileMap) IsPassable(x, y int, flag PassFlag) bool {
	t := tm.At(x, y)
	if t == nil {
		return false
	}
	return t.Passable&flag != 0 && !t.Occupied
}

// WorldToIso converts world tile coords to isometric screen coords
func (tm *TileMap) WorldToIso(wx, wy float64) (sx, sy float64) {
	tw := float64(tm.TileWidth)
	th := float64(tm.TileHeight)
	sx = (wx - wy) * (tw / 2)
	sy = (wx + wy) * (th / 2)
	return
}

// IsoToWorld converts isometric screen coords to world tile coords
func (tm *TileMap) IsoToWorld(sx, sy float64) (wx, wy float64) {
	tw := float64(tm.TileWidth)
	th := float64(tm.TileHeight)
	wx = (sx/tw + sy/th)
	wy = (sy/th - sx/tw)
	return
}

// SaveJSON saves the map to a JSON file
func (tm *TileMap) SaveJSON(path string) error {
	data, err := json.MarshalIndent(tm, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadJSON loads a map from a JSON file
func LoadJSON(path string) (*TileMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tm TileMap
	if err := json.Unmarshal(data, &tm); err != nil {
		return nil, err
	}
	return &tm, nil
}

// SetTerrain sets terrain for a rectangular region
func (tm *TileMap) SetTerrain(x1, y1, x2, y2 int, terrain TerrainType) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if t := tm.At(x, y); t != nil {
				t.Terrain = terrain
				// Update passability based on terrain
				switch terrain {
				case TerrainWater, TerrainDeepWater:
					t.Passable = PassNaval | PassAir
				case TerrainCliff:
					t.Passable = PassAir
				case TerrainRock:
					t.Passable = PassInfantry | PassAir
				default:
					t.Passable = PassAll
				}
			}
		}
	}
}

// PlaceOre places ore resources at a position
// SetOccupied marks a tile as occupied/unoccupied by a building
func (tm *TileMap) SetOccupied(x, y int, occupied bool) {
	if t := tm.At(x, y); t != nil {
		t.Occupied = occupied
	}
}

func (tm *TileMap) PlaceOre(x, y, amount int) {
	if t := tm.At(x, y); t != nil {
		t.Terrain = TerrainOre
		t.OreAmount = amount
	}
}

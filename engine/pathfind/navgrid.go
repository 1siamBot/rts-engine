package pathfind

import "github.com/1siamBot/rts-engine/engine/maplib"

// NavGrid provides a navigation grid derived from the tile map
type NavGrid struct {
	Width, Height int
	Costs         []float64 // movement cost per cell (0 = impassable)
	passFlags     []maplib.PassFlag
}

// NewNavGrid builds a navigation grid from a tile map
func NewNavGrid(tm *maplib.TileMap) *NavGrid {
	ng := &NavGrid{
		Width:     tm.Width,
		Height:    tm.Height,
		Costs:     make([]float64, tm.Width*tm.Height),
		passFlags: make([]maplib.PassFlag, tm.Width*tm.Height),
	}
	for i, t := range tm.Tiles {
		ng.passFlags[i] = t.Passable
		if t.Passable == 0 || t.Occupied {
			ng.Costs[i] = 0
		} else {
			switch t.Terrain {
			case maplib.TerrainRoad, maplib.TerrainBridge:
				ng.Costs[i] = 0.7
			case maplib.TerrainForest:
				ng.Costs[i] = 1.5
			case maplib.TerrainSand:
				ng.Costs[i] = 1.3
			case maplib.TerrainRock:
				ng.Costs[i] = 2.0
			default:
				ng.Costs[i] = 1.0
			}
		}
	}
	return ng
}

// Passable checks if a cell is passable for a given movement flag
func (ng *NavGrid) Passable(x, y int, flag maplib.PassFlag) bool {
	if x < 0 || y < 0 || x >= ng.Width || y >= ng.Height {
		return false
	}
	return ng.passFlags[y*ng.Width+x]&flag != 0 && ng.Costs[y*ng.Width+x] > 0
}

// Cost returns the movement cost at (x,y)
func (ng *NavGrid) Cost(x, y int) float64 {
	if x < 0 || y < 0 || x >= ng.Width || y >= ng.Height {
		return 0
	}
	return ng.Costs[y*ng.Width+x]
}

// SetBlocked marks a cell as blocked (for runtime building placement)
func (ng *NavGrid) SetBlocked(x, y int) {
	if x >= 0 && y >= 0 && x < ng.Width && y < ng.Height {
		ng.Costs[y*ng.Width+x] = 0
	}
}

// SetCost sets a custom cost for a cell
func (ng *NavGrid) SetCost(x, y int, cost float64) {
	if x >= 0 && y >= 0 && x < ng.Width && y < ng.Height {
		ng.Costs[y*ng.Width+x] = cost
	}
}

// Refresh rebuilds the nav grid from a tile map
func (ng *NavGrid) Refresh(tm *maplib.TileMap) {
	*ng = *NewNavGrid(tm)
}

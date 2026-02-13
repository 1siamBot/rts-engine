package systems

import (
	"github.com/1siamBot/rts-engine/engine/core"
)

// FogState represents visibility of a tile
type FogState uint8

const (
	FogShroud   FogState = iota // never seen
	FogExplored                  // seen before but not now
	FogVisible                   // currently visible
)

// FogOfWar manages visibility per player
type FogOfWar struct {
	Width, Height int
	Grid          []FogState // per-tile fog state
	PlayerID      int
}

func NewFogOfWar(w, h, playerID int) *FogOfWar {
	return &FogOfWar{
		Width:    w,
		Height:   h,
		Grid:     make([]FogState, w*h),
		PlayerID: playerID,
	}
}

// At returns the fog state at (x, y)
func (f *FogOfWar) At(x, y int) FogState {
	if x < 0 || y < 0 || x >= f.Width || y >= f.Height {
		return FogShroud
	}
	return f.Grid[y*f.Width+x]
}

// IsVisible returns true if tile is currently visible
func (f *FogOfWar) IsVisible(x, y int) bool {
	return f.At(x, y) == FogVisible
}

// FogSystem updates fog of war each tick
type FogSystem struct {
	Fogs    map[int]*FogOfWar // playerID -> fog
	Players *core.PlayerManager
}

func NewFogSystem(w, h int, pm *core.PlayerManager) *FogSystem {
	fs := &FogSystem{
		Fogs:    make(map[int]*FogOfWar),
		Players: pm,
	}
	for _, p := range pm.Players {
		fs.Fogs[p.ID] = NewFogOfWar(w, h, p.ID)
	}
	return fs
}

func (s *FogSystem) Priority() int { return 2 }

func (s *FogSystem) Update(w *core.World, _ float64) {
	// Demote all visible to explored
	for _, fog := range s.Fogs {
		for i := range fog.Grid {
			if fog.Grid[i] == FogVisible {
				fog.Grid[i] = FogExplored
			}
		}
	}

	// Reveal tiles around units with FogVision
	units := w.Query(core.CompPosition, core.CompFogVision, core.CompOwner)
	for _, id := range units {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		vis := w.Get(id, core.CompFogVision).(*core.FogVision)
		own := w.Get(id, core.CompOwner).(*core.Owner)

		fog := s.Fogs[own.PlayerID]
		if fog == nil {
			continue
		}

		cx, cy := int(pos.X), int(pos.Y)
		r := vis.Range
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if dx*dx+dy*dy <= r*r {
					tx, ty := cx+dx, cy+dy
					if tx >= 0 && ty >= 0 && tx < fog.Width && ty < fog.Height {
						fog.Grid[ty*fog.Width+tx] = FogVisible
					}
				}
			}
		}

		// Also reveal for allies
		for _, p := range s.Players.Players {
			if p.ID != own.PlayerID && s.Players.AreAllies(own.PlayerID, p.ID) {
				afog := s.Fogs[p.ID]
				if afog == nil {
					continue
				}
				for dy := -r; dy <= r; dy++ {
					for dx := -r; dx <= r; dx++ {
						if dx*dx+dy*dy <= r*r {
							tx, ty := cx+dx, cy+dy
							if tx >= 0 && ty >= 0 && tx < afog.Width && ty < afog.Height {
								afog.Grid[ty*afog.Width+tx] = FogVisible
							}
						}
					}
				}
			}
		}
	}
}

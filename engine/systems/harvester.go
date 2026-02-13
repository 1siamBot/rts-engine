package systems

import (
	"math"

	"github.com/1siamBot/rts-engine/engine/core"
	"github.com/1siamBot/rts-engine/engine/maplib"
	"github.com/1siamBot/rts-engine/engine/pathfind"
)

// HarvesterSystem manages resource gathering
type HarvesterSystem struct {
	NavGrid  *pathfind.NavGrid
	TileMap  *maplib.TileMap
	Players  *core.PlayerManager
	EventBus *core.EventBus
}

func (s *HarvesterSystem) Priority() int { return 30 }

func (s *HarvesterSystem) Update(w *core.World, dt float64) {
	ids := w.Query(core.CompPosition, core.CompHarvester, core.CompMovable, core.CompOwner)
	for _, id := range ids {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		harv := w.Get(id, core.CompHarvester).(*core.Harvester)
		mov := w.Get(id, core.CompMovable).(*core.Movable)
		own := w.Get(id, core.CompOwner).(*core.Owner)

		switch harv.State {
		case core.HarvIdle:
			// Find nearest ore
			ox, oy := s.findNearestOre(int(pos.X), int(pos.Y))
			if ox >= 0 {
				harv.State = core.HarvMovingToOre
				OrderMove(w, s.NavGrid, id, ox, oy)
			}

		case core.HarvMovingToOre:
			if mov.PathIdx >= len(mov.Path) {
				// Arrived at ore
				tx, ty := int(pos.X), int(pos.Y)
				tile := s.TileMap.At(tx, ty)
				if tile != nil && tile.OreAmount > 0 {
					harv.State = core.HarvHarvesting
				} else {
					harv.State = core.HarvIdle
				}
			}

		case core.HarvHarvesting:
			tx, ty := int(pos.X), int(pos.Y)
			tile := s.TileMap.At(tx, ty)
			if tile == nil || tile.OreAmount <= 0 {
				if harv.Current > 0 {
					harv.State = core.HarvReturning
					s.returnToRefinery(w, id, pos, mov)
				} else {
					harv.State = core.HarvIdle
				}
				continue
			}
			amount := int(harv.Rate * dt * 20)
			if amount < 1 {
				amount = 1
			}
			if amount > tile.OreAmount {
				amount = tile.OreAmount
			}
			remaining := harv.Capacity - harv.Current
			if amount > remaining {
				amount = remaining
			}
			harv.Current += amount
			tile.OreAmount -= amount
			if tile.OreAmount <= 0 {
				tile.Terrain = maplib.TerrainDirt
			}
			if harv.Current >= harv.Capacity {
				harv.State = core.HarvReturning
				s.returnToRefinery(w, id, pos, mov)
			}

		case core.HarvReturning:
			if mov.PathIdx >= len(mov.Path) {
				harv.State = core.HarvUnloading
			}

		case core.HarvUnloading:
			player := s.Players.GetPlayer(own.PlayerID)
			if player != nil {
				value := harv.Current * 25 // each unit of ore = $25
				if harv.Resource == "gem" {
					value = harv.Current * 50
				}
				player.Credits += value
				if s.EventBus != nil {
					s.EventBus.Emit(core.Event{Type: core.EvtResourceHarvested, Tick: w.TickCount})
				}
			}
			harv.Current = 0
			harv.State = core.HarvIdle
		}
	}
}

func (s *HarvesterSystem) findNearestOre(fx, fy int) (int, int) {
	bestDist := math.MaxFloat64
	bx, by := -1, -1
	for y := 0; y < s.TileMap.Height; y++ {
		for x := 0; x < s.TileMap.Width; x++ {
			t := s.TileMap.At(x, y)
			if t != nil && t.OreAmount > 0 {
				dx := float64(x - fx)
				dy := float64(y - fy)
				d := dx*dx + dy*dy
				if d < bestDist {
					bestDist = d
					bx, by = x, y
				}
			}
		}
	}
	return bx, by
}

func (s *HarvesterSystem) returnToRefinery(w *core.World, id core.EntityID, pos *core.Position, mov *core.Movable) {
	// Find nearest own refinery/construction yard
	own := w.Get(id, core.CompOwner).(*core.Owner)
	buildings := w.Query(core.CompPosition, core.CompBuilding, core.CompOwner)
	bestDist := math.MaxFloat64
	bx, by := int(pos.X), int(pos.Y)
	for _, bid := range buildings {
		bown := w.Get(bid, core.CompOwner).(*core.Owner)
		if bown.PlayerID != own.PlayerID {
			continue
		}
		bpos := w.Get(bid, core.CompPosition).(*core.Position)
		d := pos.DistanceTo(bpos)
		if d < bestDist {
			bestDist = d
			bx, by = int(bpos.X), int(bpos.Y)
		}
	}
	OrderMove(w, s.NavGrid, id, bx, by)
}

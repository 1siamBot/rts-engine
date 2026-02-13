package systems

import (
	"math"

	"github.com/1siamBot/rts-engine/engine/core"
	"github.com/1siamBot/rts-engine/engine/maplib"
	"github.com/1siamBot/rts-engine/engine/pathfind"
)

// MovementSystem moves units along their paths
type MovementSystem struct {
	NavGrid *pathfind.NavGrid
}

func (s *MovementSystem) Priority() int { return 10 }

func (s *MovementSystem) Update(w *core.World, dt float64) {
	ids := w.Query(core.CompPosition, core.CompMovable)
	// Collect positions for steering
	positions := make(map[core.EntityID][3]float64)
	for _, id := range ids {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		positions[id] = [3]float64{pos.X, pos.Y, 0.5}
	}

	for _, id := range ids {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		mov := w.Get(id, core.CompMovable).(*core.Movable)

		if mov.PathIdx >= len(mov.Path) {
			continue
		}

		// Collect nearby units for avoidance
		var others [][3]float64
		for oid, op := range positions {
			if oid != id {
				dx := pos.X - op[0]
				dy := pos.Y - op[1]
				if dx*dx+dy*dy < 9 { // within 3 tiles
					others = append(others, [3]float64{op[0], op[1], op[2]})
				}
			}
		}

		// Convert path to Point slice for steering
		pts := make([]pathfind.Point, len(mov.Path))
		for i, tp := range mov.Path {
			pts[i] = pathfind.Point{X: tp.X, Y: tp.Y}
		}
		steer := pathfind.Steer(pos.X, pos.Y, mov.Speed, pts, mov.PathIdx, others)
		pos.X += steer.VX * dt
		pos.Y += steer.VY * dt

		// Update facing
		if steer.VX != 0 || steer.VY != 0 {
			pos.Facing = math.Atan2(steer.VY, steer.VX)
		}

		// Check if reached current waypoint
		target := mov.Path[mov.PathIdx]
		tx, ty := float64(target.X)+0.5, float64(target.Y)+0.5
		dx, dy := tx-pos.X, ty-pos.Y
		if dx*dx+dy*dy < 0.15 {
			mov.PathIdx++
		}
	}
}

// MovePassFlag converts core.MoveType to maplib.PassFlag
func MovePassFlag(mt core.MoveType) maplib.PassFlag {
	switch mt {
	case core.MoveInfantry:
		return maplib.PassInfantry
	case core.MoveVehicle:
		return maplib.PassVehicle
	case core.MoveNaval:
		return maplib.PassNaval
	case core.MoveAir:
		return maplib.PassAir
	case core.MoveAmphibious:
		return maplib.PassVehicle | maplib.PassNaval
	default:
		return maplib.PassAll
	}
}

// OrderMove sets a path for an entity to a destination
func OrderMove(w *core.World, ng *pathfind.NavGrid, id core.EntityID, gx, gy int) {
	pos := w.Get(id, core.CompPosition)
	mov := w.Get(id, core.CompMovable)
	if pos == nil || mov == nil {
		return
	}
	p := pos.(*core.Position)
	m := mov.(*core.Movable)
	sx, sy := int(p.X), int(p.Y)
	flag := MovePassFlag(m.MoveType)
	path := pathfind.FindPath(ng, sx, sy, gx, gy, flag)
	if path != nil {
		path = pathfind.SmoothPath(ng, path, flag)
		m.Path = make([]core.TilePos, len(path))
		for i, pt := range path {
			m.Path[i] = core.TilePos{X: pt.X, Y: pt.Y}
		}
		m.PathIdx = 0
	}
}

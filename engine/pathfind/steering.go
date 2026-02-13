package pathfind

import "math"

// SteerResult contains the computed steering velocity
type SteerResult struct {
	VX, VY float64
}

// Steer computes a velocity for a unit moving along a path while avoiding others
// ux, uy: unit position; speed: max speed; path: waypoints; pathIdx: current waypoint
// others: list of (x, y, radius) of nearby units to avoid
func Steer(ux, uy, speed float64, path []Point, pathIdx int, others [][3]float64) SteerResult {
	if pathIdx >= len(path) {
		return SteerResult{}
	}

	// Seek toward current waypoint
	target := path[pathIdx]
	tx, ty := float64(target.X)+0.5, float64(target.Y)+0.5
	dx, dy := tx-ux, ty-uy
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.01 {
		return SteerResult{}
	}

	seekX, seekY := dx/dist*speed, dy/dist*speed

	// Separation from other units
	sepX, sepY := 0.0, 0.0
	for _, o := range others {
		ox, oy, or := o[0], o[1], o[2]
		sx, sy := ux-ox, uy-oy
		d := math.Sqrt(sx*sx + sy*sy)
		minDist := or + 0.5
		if d < minDist && d > 0.001 {
			force := (minDist - d) / minDist
			sepX += sx / d * force * speed * 0.5
			sepY += sy / d * force * speed * 0.5
		}
	}

	vx := seekX + sepX
	vy := seekY + sepY

	// Clamp to max speed
	v := math.Sqrt(vx*vx + vy*vy)
	if v > speed {
		vx = vx / v * speed
		vy = vy / v * speed
	}

	return SteerResult{VX: vx, VY: vy}
}

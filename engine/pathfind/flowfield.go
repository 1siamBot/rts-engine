package pathfind

import (
	"math"

	"github.com/1siamBot/rts-engine/engine/maplib"
)

// FlowField stores a direction vector for each cell pointing toward the goal
type FlowField struct {
	Width, Height int
	DirX, DirY    []float64
	Cost          []float64 // integration field cost
}

// NewFlowField generates a flow field toward (gx, gy) for the given movement flag
func NewFlowField(ng *NavGrid, gx, gy int, flag maplib.PassFlag) *FlowField {
	w, h := ng.Width, ng.Height
	ff := &FlowField{
		Width:  w,
		Height: h,
		DirX:   make([]float64, w*h),
		DirY:   make([]float64, w*h),
		Cost:   make([]float64, w*h),
	}

	inf := math.MaxFloat64
	for i := range ff.Cost {
		ff.Cost[i] = inf
	}
	if gx < 0 || gy < 0 || gx >= w || gy >= h {
		return ff
	}
	ff.Cost[gy*w+gx] = 0

	// BFS integration pass
	type pt struct{ x, y int }
	queue := []pt{{gx, gy}}
	dirs := [8][2]int{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		curCost := ff.Cost[cur.y*w+cur.x]
		for _, d := range dirs {
			nx, ny := cur.x+d[0], cur.y+d[1]
			if !ng.Passable(nx, ny, flag) {
				continue
			}
			moveCost := ng.Cost(nx, ny)
			if d[0] != 0 && d[1] != 0 {
				moveCost *= math.Sqrt2
			}
			newCost := curCost + moveCost
			idx := ny*w + nx
			if newCost < ff.Cost[idx] {
				ff.Cost[idx] = newCost
				queue = append(queue, pt{nx, ny})
			}
		}
	}

	// Direction pass: each cell points toward lowest-cost neighbor
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			if ff.Cost[idx] >= inf {
				continue
			}
			bestCost := ff.Cost[idx]
			var bx, by float64
			for _, d := range dirs {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || ny < 0 || nx >= w || ny >= h {
					continue
				}
				c := ff.Cost[ny*w+nx]
				if c < bestCost {
					bestCost = c
					bx, by = float64(d[0]), float64(d[1])
				}
			}
			// Normalize
			length := math.Sqrt(bx*bx + by*by)
			if length > 0 {
				ff.DirX[idx] = bx / length
				ff.DirY[idx] = by / length
			}
		}
	}

	return ff
}

// Direction returns the flow direction at (x,y)
func (ff *FlowField) Direction(x, y int) (float64, float64) {
	if x < 0 || y < 0 || x >= ff.Width || y >= ff.Height {
		return 0, 0
	}
	idx := y*ff.Width + x
	return ff.DirX[idx], ff.DirY[idx]
}

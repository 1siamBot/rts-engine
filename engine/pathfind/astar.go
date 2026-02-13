package pathfind

import (
	"container/heap"
	"math"

	"github.com/1siamBot/rts-engine/engine/maplib"
)

// Point represents a 2D integer coordinate
type Point struct{ X, Y int }

// FindPath finds a path from start to goal using A*
func FindPath(ng *NavGrid, sx, sy, gx, gy int, flag maplib.PassFlag) []Point {
	if !ng.Passable(gx, gy, flag) {
		return nil
	}

	start := Point{sx, sy}
	goal := Point{gx, gy}

	open := &nodeHeap{}
	heap.Init(open)
	heap.Push(open, &node{p: start, g: 0, f: heuristic(start, goal)})

	came := make(map[Point]Point)
	gScore := make(map[Point]float64)
	gScore[start] = 0

	dirs := [8][2]int{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
	}

	for open.Len() > 0 {
		cur := heap.Pop(open).(*node)
		if cur.p == goal {
			return reconstructPath(came, goal)
		}

		for _, d := range dirs {
			nx, ny := cur.p.X+d[0], cur.p.Y+d[1]
			if !ng.Passable(nx, ny, flag) {
				continue
			}
			// Prevent diagonal cutting through walls
			if d[0] != 0 && d[1] != 0 {
				if !ng.Passable(cur.p.X+d[0], cur.p.Y, flag) || !ng.Passable(cur.p.X, cur.p.Y+d[1], flag) {
					continue
				}
			}
			np := Point{nx, ny}
			moveCost := ng.Cost(nx, ny)
			if d[0] != 0 && d[1] != 0 {
				moveCost *= math.Sqrt2
			}
			tentG := gScore[cur.p] + moveCost
			if old, ok := gScore[np]; ok && tentG >= old {
				continue
			}
			gScore[np] = tentG
			came[np] = cur.p
			heap.Push(open, &node{p: np, g: tentG, f: tentG + heuristic(np, goal)})
		}
	}
	return nil // no path
}

// SmoothPath removes unnecessary waypoints using line-of-sight checks
func SmoothPath(ng *NavGrid, path []Point, flag maplib.PassFlag) []Point {
	if len(path) <= 2 {
		return path
	}
	smooth := []Point{path[0]}
	cur := 0
	for cur < len(path)-1 {
		farthest := cur + 1
		for i := len(path) - 1; i > cur+1; i-- {
			if lineOfSight(ng, path[cur], path[i], flag) {
				farthest = i
				break
			}
		}
		smooth = append(smooth, path[farthest])
		cur = farthest
	}
	return smooth
}

func lineOfSight(ng *NavGrid, a, b Point, flag maplib.PassFlag) bool {
	dx := abs(b.X - a.X)
	dy := abs(b.Y - a.Y)
	sx, sy := 1, 1
	if a.X > b.X {
		sx = -1
	}
	if a.Y > b.Y {
		sy = -1
	}
	err := dx - dy
	x, y := a.X, a.Y
	for {
		if !ng.Passable(x, y, flag) {
			return false
		}
		if x == b.X && y == b.Y {
			return true
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func heuristic(a, b Point) float64 {
	dx := math.Abs(float64(a.X - b.X))
	dy := math.Abs(float64(a.Y - b.Y))
	return dx + dy + (math.Sqrt2-2)*math.Min(dx, dy)
}

func reconstructPath(came map[Point]Point, goal Point) []Point {
	path := []Point{goal}
	cur := goal
	for {
		prev, ok := came[cur]
		if !ok {
			break
		}
		path = append(path, prev)
		cur = prev
	}
	// Reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// --- Priority queue ---

type node struct {
	p    Point
	g, f float64
}

type nodeHeap []*node

func (h nodeHeap) Len() int            { return len(h) }
func (h nodeHeap) Less(i, j int) bool   { return h[i].f < h[j].f }
func (h nodeHeap) Swap(i, j int)        { h[i], h[j] = h[j], h[i] }
func (h *nodeHeap) Push(x interface{})  { *h = append(*h, x.(*node)) }
func (h *nodeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

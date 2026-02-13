package render3d

import (
	"math"

	"github.com/1siamBot/rts-engine/engine/maplib"
)

// TerrainColors maps terrain to natural, moderate base colors
var TerrainBaseColors = map[maplib.TerrainType]Color3{
	maplib.TerrainGrass:     {0.30, 0.55, 0.22},
	maplib.TerrainDirt:      {0.55, 0.42, 0.28},
	maplib.TerrainSand:      {0.78, 0.72, 0.50},
	maplib.TerrainWater:     {0.18, 0.42, 0.72},
	maplib.TerrainDeepWater: {0.10, 0.22, 0.58},
	maplib.TerrainRock:      {0.50, 0.48, 0.45},
	maplib.TerrainCliff:     {0.45, 0.42, 0.38},
	maplib.TerrainRoad:      {0.55, 0.53, 0.50},
	maplib.TerrainBridge:    {0.52, 0.38, 0.22},
	maplib.TerrainOre:       {0.72, 0.62, 0.15},
	maplib.TerrainGem:       {0.15, 0.68, 0.68},
	maplib.TerrainSnow:      {0.82, 0.82, 0.85},
	maplib.TerrainUrban:     {0.58, 0.56, 0.54},
	maplib.TerrainForest:    {0.16, 0.48, 0.12},
}

// perlinNoise simple hash-based noise for height variation
func perlinNoise(x, y int) float64 {
	h := uint32(x*73856093 ^ y*19349663)
	h = (h >> 13) ^ h
	h = h * (h*h*15731 + 789221) + 1376312589
	return float64(h&0x7fffffff) / float64(0x7fffffff)
}

func smoothNoise(x, y int) float64 {
	corners := (perlinNoise(x-1, y-1) + perlinNoise(x+1, y-1) + perlinNoise(x-1, y+1) + perlinNoise(x+1, y+1)) / 16.0
	sides := (perlinNoise(x-1, y) + perlinNoise(x+1, y) + perlinNoise(x, y-1) + perlinNoise(x, y+1)) / 8.0
	center := perlinNoise(x, y) / 4.0
	return corners + sides + center
}

// GenerateTerrainMesh creates a 3D mesh for a tile range
func GenerateTerrainMesh(tm *maplib.TileMap, minX, minY, maxX, maxY int, time float64) *Mesh3D {
	mesh := NewMesh()
	up := V3(0, 1, 0)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			tile := tm.At(x, y)
			if tile == nil {
				continue
			}

			baseColor, ok := TerrainBaseColors[tile.Terrain]
			if !ok {
				baseColor = Color3{0.5, 0.5, 0.5}
			}

			// Per-tile color variation for natural look
			hash := uint32(x*7919+y*7927+x*y*31) % 1000
			variation := float64(hash)/1000.0*0.08 - 0.04
			baseColor.R = math.Max(0, math.Min(1, baseColor.R+variation))
			baseColor.G = math.Max(0, math.Min(1, baseColor.G+variation*1.2))
			baseColor.B = math.Max(0, math.Min(1, baseColor.B+variation*0.5))

			// Height: tile height + perlin noise variation
			h := float64(tile.Height) * 0.15
			noiseH := smoothNoise(x, y) * 0.06
			h += noiseH

			// Water: animate with wave
			if tile.Terrain == maplib.TerrainWater || tile.Terrain == maplib.TerrainDeepWater {
				h = -0.05 + 0.04*math.Sin(time*2.0+float64(x)*0.5+float64(y)*0.7)
				// Specular shimmer
				spec := 0.15 * math.Abs(math.Sin(time*1.8+float64(x)*0.4+float64(y)*0.3))
				baseColor.R = math.Min(1, baseColor.R+spec*0.3)
				baseColor.G = math.Min(1, baseColor.G+spec*0.5)
				baseColor.B = math.Min(1, baseColor.B+spec)
			}

			// Ore golden sparkle
			if tile.OreAmount > 0 {
				sparkle := 0.25 * math.Abs(math.Sin(time*3+float64(x*7+y*13)))
				baseColor.R = math.Min(1, baseColor.R+sparkle)
				baseColor.G = math.Min(1, baseColor.G+sparkle*0.7)
				baseColor.B = math.Min(1, baseColor.B+sparkle*0.1)
			}

			fx, fz := float64(x), float64(y)

			// Corner heights for smoother terrain
			h00 := h
			h10 := h + (smoothNoise(x+1, y)-smoothNoise(x, y))*0.02
			h11 := h + (smoothNoise(x+1, y+1)-smoothNoise(x, y))*0.02
			h01 := h + (smoothNoise(x, y+1)-smoothNoise(x, y))*0.02

			v0 := Vertex3D{Pos: V3(fx, h00, fz), Normal: up, Color: baseColor}
			v1 := Vertex3D{Pos: V3(fx+1, h10, fz), Normal: up, Color: baseColor}
			v2 := Vertex3D{Pos: V3(fx+1, h11, fz+1), Normal: up, Color: baseColor}
			v3 := Vertex3D{Pos: V3(fx, h01, fz+1), Normal: up, Color: baseColor}
			mesh.AddQuad(v0, v1, v2, v3)

			// Subtle grid lines (slightly darker edges)
			gridColor := Color3{baseColor.R * 0.85, baseColor.G * 0.85, baseColor.B * 0.85}
			gridW := 0.03
			// Bottom edge
			ge0 := Vertex3D{Pos: V3(fx, h00+0.005, fz), Normal: up, Color: gridColor}
			ge1 := Vertex3D{Pos: V3(fx+1, h10+0.005, fz), Normal: up, Color: gridColor}
			ge2 := Vertex3D{Pos: V3(fx+1, h10+0.005, fz+gridW), Normal: up, Color: gridColor}
			ge3 := Vertex3D{Pos: V3(fx, h00+0.005, fz+gridW), Normal: up, Color: gridColor}
			mesh.AddQuad(ge0, ge1, ge2, ge3)
			// Left edge
			gl0 := Vertex3D{Pos: V3(fx, h00+0.005, fz), Normal: up, Color: gridColor}
			gl1 := Vertex3D{Pos: V3(fx+gridW, h00+0.005, fz), Normal: up, Color: gridColor}
			gl2 := Vertex3D{Pos: V3(fx+gridW, h01+0.005, fz+1), Normal: up, Color: gridColor}
			gl3 := Vertex3D{Pos: V3(fx, h01+0.005, fz+1), Normal: up, Color: gridColor}
			mesh.AddQuad(gl0, gl1, gl2, gl3)

			// Cliff/elevation side faces
			if tile.Height > 0 {
				sideColor := Color3{baseColor.R * 0.65, baseColor.G * 0.65, baseColor.B * 0.65}
				addTerrainSides(mesh, tm, x, y, h, sideColor)
			}

			// Forest: add 3D tree geometry (cone + cylinder)
			if tile.Terrain == maplib.TerrainForest {
				addTreeGeometry(mesh, fx+0.5, h, fz+0.5, hash)
			}
		}
	}
	return mesh
}

func addTerrainSides(mesh *Mesh3D, tm *maplib.TileMap, x, y int, h float64, c Color3) {
	fx, fz := float64(x), float64(y)
	dirs := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	normals := []Vec3{{-1, 0, 0}, {1, 0, 0}, {0, 0, -1}, {0, 0, 1}}

	for di, d := range dirs {
		nx, ny := x+d[0], y+d[1]
		adjH := 0.0
		if adj := tm.At(nx, ny); adj != nil {
			adjH = float64(adj.Height) * 0.15
		}
		if h > adjH+0.01 {
			n := normals[di]
			var v0, v1, v2, v3 Vertex3D
			switch di {
			case 0: // -X
				v0 = Vertex3D{Pos: V3(fx, adjH, fz), Normal: n, Color: c}
				v1 = Vertex3D{Pos: V3(fx, adjH, fz+1), Normal: n, Color: c}
				v2 = Vertex3D{Pos: V3(fx, h, fz+1), Normal: n, Color: c}
				v3 = Vertex3D{Pos: V3(fx, h, fz), Normal: n, Color: c}
			case 1: // +X
				v0 = Vertex3D{Pos: V3(fx+1, adjH, fz+1), Normal: n, Color: c}
				v1 = Vertex3D{Pos: V3(fx+1, adjH, fz), Normal: n, Color: c}
				v2 = Vertex3D{Pos: V3(fx+1, h, fz), Normal: n, Color: c}
				v3 = Vertex3D{Pos: V3(fx+1, h, fz+1), Normal: n, Color: c}
			case 2: // -Z
				v0 = Vertex3D{Pos: V3(fx+1, adjH, fz), Normal: n, Color: c}
				v1 = Vertex3D{Pos: V3(fx, adjH, fz), Normal: n, Color: c}
				v2 = Vertex3D{Pos: V3(fx, h, fz), Normal: n, Color: c}
				v3 = Vertex3D{Pos: V3(fx+1, h, fz), Normal: n, Color: c}
			case 3: // +Z
				v0 = Vertex3D{Pos: V3(fx, adjH, fz+1), Normal: n, Color: c}
				v1 = Vertex3D{Pos: V3(fx+1, adjH, fz+1), Normal: n, Color: c}
				v2 = Vertex3D{Pos: V3(fx+1, h, fz+1), Normal: n, Color: c}
				v3 = Vertex3D{Pos: V3(fx, h, fz+1), Normal: n, Color: c}
			}
			mesh.AddQuad(v0, v1, v2, v3)
		}
	}
}

func addTreeGeometry(mesh *Mesh3D, cx, baseH, cz float64, hash uint32) {
	trunkH := 0.25 + float64(hash%100)/600.0
	trunkR := 0.04

	// Trunk (4-sided prism)
	tc := Color3{0.45, 0.30, 0.12}
	trunk := MakeBox(trunkR*2, trunkH, trunkR*2, tc)
	trunkMat := Mat4Translate(cx, baseH+trunkH/2, cz)
	mesh.Append(trunk.Transform(trunkMat))

	// Canopy as cone (using cylinder with tapered top approximation)
	canopyR := 0.18 + float64(hash%150)/600.0
	canopyH := 0.30 + float64(hash%80)/350.0
	cc := Color3{0.08, 0.45 + float64(hash%100)/400.0, 0.06}
	canopy := MakeCone(canopyR, canopyH, 6, cc)
	canopyMat := Mat4Translate(cx, baseH+trunkH+canopyH/2, cz)
	mesh.Append(canopy.Transform(canopyMat))

	// Second smaller cone on top for fuller look
	cc2 := Color3{cc.R + 0.03, cc.G + 0.08, cc.B + 0.02}
	canopy2 := MakeCone(canopyR*0.65, canopyH*0.7, 6, cc2)
	canopy2Mat := Mat4Translate(cx, baseH+trunkH+canopyH*0.8+canopyH*0.35, cz)
	mesh.Append(canopy2.Transform(canopy2Mat))
}

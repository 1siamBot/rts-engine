package render3d

import (
	"math"

	"github.com/1siamBot/rts-engine/engine/maplib"
)

// TerrainColors maps terrain to base colors
var TerrainBaseColors = map[maplib.TerrainType]Color3{
	maplib.TerrainGrass:     {0.18, 0.55, 0.18},
	maplib.TerrainDirt:      {0.55, 0.45, 0.35},
	maplib.TerrainSand:      {0.85, 0.78, 0.58},
	maplib.TerrainWater:     {0.15, 0.45, 0.85},
	maplib.TerrainDeepWater: {0.05, 0.15, 0.55},
	maplib.TerrainRock:      {0.50, 0.50, 0.50},
	maplib.TerrainCliff:     {0.40, 0.38, 0.35},
	maplib.TerrainRoad:      {0.55, 0.55, 0.55},
	maplib.TerrainBridge:    {0.55, 0.35, 0.17},
	maplib.TerrainOre:       {0.75, 0.65, 0.10},
	maplib.TerrainGem:       {0.10, 0.80, 0.80},
	maplib.TerrainSnow:      {0.92, 0.92, 0.96},
	maplib.TerrainUrban:     {0.65, 0.65, 0.65},
	maplib.TerrainForest:    {0.08, 0.40, 0.08},
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

			// Add slight color variation per tile for natural look
			hash := uint32(x*7919+y*7927+x*y*31) % 1000
			variation := float64(hash)/1000.0*0.1 - 0.05
			baseColor.R = math.Max(0, math.Min(1, baseColor.R+variation))
			baseColor.G = math.Max(0, math.Min(1, baseColor.G+variation))
			baseColor.B = math.Max(0, math.Min(1, baseColor.B+variation))

			// Height: tile height + slight noise
			h := float64(tile.Height) * 0.15
			noiseH := float64(hash%100) / 100.0 * 0.03
			h += noiseH

			// Water: animate with wave
			if tile.Terrain == maplib.TerrainWater || tile.Terrain == maplib.TerrainDeepWater {
				h = -0.05 + 0.03*math.Sin(time*2+float64(x)*0.5+float64(y)*0.7)
				// Add specular highlight simulation
				spec := 0.1 * math.Abs(math.Sin(time*1.5+float64(x)*0.3))
				baseColor.R = math.Min(1, baseColor.R+spec)
				baseColor.G = math.Min(1, baseColor.G+spec)
				baseColor.B = math.Min(1, baseColor.B+spec)
			}

			// Ore sparkle
			if tile.OreAmount > 0 {
				sparkle := 0.2 * math.Abs(math.Sin(time*3+float64(x*7+y*13)))
				baseColor.R = math.Min(1, baseColor.R+sparkle)
				baseColor.G = math.Min(1, baseColor.G+sparkle*0.8)
			}

			fx, fz := float64(x), float64(y)

			// Tile quad on XZ plane at height h
			v0 := Vertex3D{Pos: V3(fx, h, fz), Normal: up, Color: baseColor}
			v1 := Vertex3D{Pos: V3(fx+1, h, fz), Normal: up, Color: baseColor}
			v2 := Vertex3D{Pos: V3(fx+1, h, fz+1), Normal: up, Color: baseColor}
			v3 := Vertex3D{Pos: V3(fx, h, fz+1), Normal: up, Color: baseColor}
			mesh.AddQuad(v0, v1, v2, v3)

			// Cliff/elevation side faces
			if tile.Height > 0 {
				sideColor := Color3{baseColor.R * 0.6, baseColor.G * 0.6, baseColor.B * 0.6}
				// Check adjacent tiles for height differences and add sides
				addTerrainSides(mesh, tm, x, y, h, sideColor)
			}

			// Forest: add tree stump geometry
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
	// Simple tree: cylinder trunk + cone/cylinder canopy
	trunkH := 0.3 + float64(hash%100)/500.0
	trunkR := 0.05

	// Trunk (simplified as 4-sided prism)
	tc := Color3{0.4, 0.25, 0.1}
	trunk := MakeBox(trunkR*2, trunkH, trunkR*2, tc)
	trunkMat := Mat4Translate(cx, baseH+trunkH/2, cz)
	mesh.Append(trunk.Transform(trunkMat))

	// Canopy (box approximation of foliage)
	canopyR := 0.2 + float64(hash%150)/500.0
	canopyH := 0.25 + float64(hash%80)/400.0
	cc := Color3{0.05, 0.35 + float64(hash%100)/500.0, 0.05}
	canopy := MakeBox(canopyR*2, canopyH, canopyR*2, cc)
	canopyMat := Mat4Translate(cx, baseH+trunkH+canopyH/2, cz)
	mesh.Append(canopy.Transform(canopyMat))
}

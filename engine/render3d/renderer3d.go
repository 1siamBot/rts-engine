package render3d

import (
	"fmt"
	"image/color"
	"math"
	"sort"

	"github.com/1siamBot/rts-engine/engine/core"
	"github.com/1siamBot/rts-engine/engine/maplib"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Renderer3D handles 3D rendering of the RTS scene
type Renderer3D struct {
	Camera    *Camera3D
	Lighting  LightingSetup
	Particles *ParticleSystem

	// Internal
	whiteImg *ebiten.Image
	time     float64

	// Building model cache: key -> mesh
	buildingModels map[string]*Mesh3D
	unitModels     map[string]*Mesh3D
	minimapImg     *ebiten.Image

	// Terrain cache
	terrainCache      *Mesh3D
	terrainCacheKey   string // "minX,minY,maxX,maxY"
	waterCache        *Mesh3D
	waterCacheKey     string
	waterCacheTime    float64
}

// NewRenderer3D creates the 3D renderer
func NewRenderer3D(screenW, screenH int) *Renderer3D {
	r := &Renderer3D{
		Camera:         NewCamera3D(screenW, screenH),
		Lighting:       DefaultLighting(),
		Particles:      NewParticleSystem(),
		buildingModels: make(map[string]*Mesh3D),
		unitModels:     make(map[string]*Mesh3D),
	}

	// 1x1 white image for colored triangle rendering
	r.whiteImg = ebiten.NewImage(4, 4)
	r.whiteImg.Fill(color.White)

	return r
}

// Update advances time-based effects
func (r *Renderer3D) Update(dt float64) {
	r.time += dt
	r.Particles.Update(dt)
}

// DrawSkyGradient fills the screen with a dark-blue-to-lighter-blue sky gradient
func (r *Renderer3D) DrawSkyGradient(screen *ebiten.Image) {
	h := r.Camera.ScreenH
	w := r.Camera.ScreenW
	// Draw in bands for efficiency
	bands := 32
	bandH := h / bands
	if bandH < 1 {
		bandH = 1
	}
	for i := 0; i < bands; i++ {
		t := float64(i) / float64(bands)
		// Top: darker blue, Bottom: lighter blue-gray
		cr := uint8(8 + t*35)
		cg := uint8(12 + t*45)
		cb := uint8(45 + t*50)
		by := i * bandH
		bh := bandH
		if i == bands-1 {
			bh = h - by
		}
		vector.DrawFilledRect(screen, 0, float32(by), float32(w), float32(bh), color.RGBA{cr, cg, cb, 255}, false)
	}
}

// DrawScene renders the complete 3D scene
func (r *Renderer3D) DrawScene(screen *ebiten.Image, tm *maplib.TileMap, world *core.World, localPlayerID int) {
	// 0. Sky gradient background
	r.DrawSkyGradient(screen)

	// 1. Terrain (static cached, water animated separately)
	minX, minY, maxX, maxY := r.Camera.VisibleTileRange(tm.Width, tm.Height)
	cacheKey := fmt.Sprintf("%d,%d,%d,%d", minX, minY, maxX, maxY)
	if r.terrainCache == nil || r.terrainCacheKey != cacheKey {
		r.terrainCache = GenerateTerrainMeshStatic(tm, minX, minY, maxX, maxY)
		r.terrainCacheKey = cacheKey
	}
	r.renderMesh(screen, r.terrainCache)

	// Water tiles (animated separately, lightweight)
	waterKey := cacheKey
	if r.waterCache == nil || r.waterCacheKey != waterKey || r.time-r.waterCacheTime > 0.05 {
		r.waterCache = GenerateWaterMesh(tm, minX, minY, maxX, maxY, r.time)
		r.waterCacheKey = waterKey
		r.waterCacheTime = r.time
	}
	r.renderMesh(screen, r.waterCache)

	// 2. Collect all entities with depth sorting
	type entityDraw struct {
		mesh  *Mesh3D
		depth float64
	}
	var entities []entityDraw

	// Buildings
	for _, id := range world.Query(core.CompBuilding, core.CompPosition, core.CompOwner) {
		pos := world.Get(id, core.CompPosition).(*core.Position)
		own := world.Get(id, core.CompOwner).(*core.Owner)
		bldg := world.Get(id, core.CompBuilding).(*core.Building)

		buildingKey := "generic"
		if bn := world.Get(id, core.CompBuildingName); bn != nil {
			buildingKey = bn.(*core.BuildingName).Key
		}

		mesh := r.getBuildingMesh(buildingKey, own.Faction)
		if mesh == nil {
			// Fallback: generic box
			fc := FactionColor(own.Faction)
			mesh = MakeBox(float64(bldg.SizeX)*0.8, 0.8, float64(bldg.SizeY)*0.8, fc)
		}

		// Position the building
		cx := pos.X + float64(bldg.SizeX)/2.0
		cz := pos.Y + float64(bldg.SizeY)/2.0
		placed := mesh.Transform(Mat4Translate(cx, 0, cz))

		// Damage tint
		if h := world.Get(id, core.CompHealth); h != nil {
			hp := h.(*core.Health)
			if hp.Ratio() < 0.5 {
				// Darken damaged buildings
				for i := range placed.Triangles {
					for j := 0; j < 3; j++ {
						c := &placed.Triangles[i].V[j].Color
						c.R = c.R*0.6 + 0.15
						c.G = c.G * 0.5
						c.B = c.B * 0.5
					}
				}
			}
		}

		_, _, depth := r.Camera.Project3DToScreen(cx, 0, cz)
		entities = append(entities, entityDraw{mesh: placed, depth: depth})
	}

	// Units
	for _, id := range world.Query(core.CompPosition, core.CompSelectable, core.CompOwner) {
		if world.Has(id, core.CompBuilding) {
			continue // skip buildings
		}
		pos := world.Get(id, core.CompPosition).(*core.Position)
		own := world.Get(id, core.CompOwner).(*core.Owner)

		mesh := r.getUnitMesh(world, id, own.Faction)
		if mesh == nil {
			continue
		}

		// Rotate to facing direction
		rotated := RotateModelY(mesh, -pos.Facing)
		placed := rotated.Transform(Mat4Translate(pos.X, pos.Z, pos.Y))

		_, _, depth := r.Camera.Project3DToScreen(pos.X, pos.Z, pos.Y)
		entities = append(entities, entityDraw{mesh: placed, depth: depth})
	}

	// Sort back-to-front
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].depth > entities[j].depth
	})

	for _, e := range entities {
		r.renderMesh(screen, e.mesh)
	}

	// 3. Projectiles
	r.drawProjectiles3D(screen, world)

	// 4. Particles
	particleMesh := r.Particles.GenerateParticleMeshes()
	r.renderMesh(screen, particleMesh)

	// 5. Selection circles
	r.drawSelectionCircles(screen, world, localPlayerID)
}

func (r *Renderer3D) getBuildingMesh(key, faction string) *Mesh3D {
	cacheKey := key + "_" + faction
	if m, ok := r.buildingModels[cacheKey]; ok {
		return m
	}
	var m *Mesh3D
	switch key {
	case "construction_yard":
		m = MakeConstructionYard(faction)
	case "power_plant":
		m = MakePowerPlant(faction)
	case "barracks":
		m = MakeBarracks(faction)
	case "war_factory":
		m = MakeWarFactory(faction)
	case "refinery":
		m = MakeRefinery(faction)
	default:
		// Generic building
		fc := FactionColor(faction)
		m = MakeBox(1.5, 0.8, 1.5, fc)
	}
	r.buildingModels[cacheKey] = m
	return m
}

func (r *Renderer3D) getUnitMesh(world *core.World, id core.EntityID, faction string) *Mesh3D {
	var key string
	if world.Has(id, core.CompMCV) {
		key = "mcv"
	} else if world.Has(id, core.CompHarvester) {
		key = "harvester"
	} else if world.Has(id, core.CompWeapon) {
		spr := world.Get(id, core.CompSprite)
		if spr != nil && spr.(*core.Sprite).Width > 26 {
			key = "tank"
		} else {
			key = "infantry"
		}
	} else {
		key = "infantry"
	}

	cacheKey := key + "_" + faction
	if m, ok := r.unitModels[cacheKey]; ok {
		return m
	}
	var m *Mesh3D
	switch key {
	case "tank":
		m = MakeTankModel(faction)
	case "infantry":
		m = MakeInfantryModel(faction)
	case "harvester":
		m = MakeHarvesterModel(faction)
	case "mcv":
		m = MakeMCVModel(faction)
	default:
		m = MakeInfantryModel(faction)
	}
	r.unitModels[cacheKey] = m
	return m
}

func (r *Renderer3D) drawProjectiles3D(screen *ebiten.Image, world *core.World) {
	for _, id := range world.Query(core.CompPosition, core.CompProjectile) {
		pos := world.Get(id, core.CompPosition).(*core.Position)
		sx, sy, _ := r.Camera.Project3DToScreen(pos.X, 0.3, pos.Y)

		// Glow
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), 6, color.RGBA{255, 200, 50, 80}, false)
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), 3, color.RGBA{255, 255, 100, 255}, false)
	}
}

func (r *Renderer3D) drawSelectionCircles(screen *ebiten.Image, world *core.World, localPlayerID int) {
	// Selection circles are drawn as projected ellipses
	for _, id := range world.Query(core.CompPosition, core.CompSelectable, core.CompOwner) {
		own := world.Get(id, core.CompOwner).(*core.Owner)
		if own.PlayerID != localPlayerID {
			continue
		}
		pos := world.Get(id, core.CompPosition).(*core.Position)
		sel := world.Get(id, core.CompSelectable).(*core.Selectable)

		// Project circle center and a point on the circumference
		cx, cy, _ := r.Camera.Project3DToScreen(pos.X, 0.01, pos.Y)
		rx, ry, _ := r.Camera.Project3DToScreen(pos.X+sel.Radius, 0.01, pos.Y)
		radius := math.Sqrt(float64((rx-cx)*(rx-cx) + (ry-cy)*(ry-cy)))

		if radius < 2 {
			radius = 2
		}

		// Draw selection ellipse
		segments := 24
		for i := 0; i < segments; i++ {
			a0 := float64(i) / float64(segments) * 2 * math.Pi
			a1 := float64(i+1) / float64(segments) * 2 * math.Pi

			wx0 := pos.X + sel.Radius*math.Cos(a0)
			wz0 := pos.Y + sel.Radius*math.Sin(a0)
			wx1 := pos.X + sel.Radius*math.Cos(a1)
			wz1 := pos.Y + sel.Radius*math.Sin(a1)

			sx0, sy0, _ := r.Camera.Project3DToScreen(wx0, 0.01, wz0)
			sx1, sy1, _ := r.Camera.Project3DToScreen(wx1, 0.01, wz1)
			_ = radius
			vector.StrokeLine(screen, float32(sx0), float32(sy0), float32(sx1), float32(sy1), 2, color.RGBA{0, 255, 0, 180}, false)
		}
	}
}

// renderMesh projects and draws a 3D mesh to the screen (batched)
func (r *Renderer3D) renderMesh(screen *ebiten.Image, mesh *Mesh3D) {
	if len(mesh.Triangles) == 0 {
		return
	}

	vp := r.Camera.ViewProj()
	sw := float64(r.Camera.ScreenW)
	sh := float64(r.Camera.ScreenH)

	// Pre-allocate batch buffers
	vertices := make([]ebiten.Vertex, 0, len(mesh.Triangles)*3)
	indices := make([]uint16, 0, len(mesh.Triangles)*3)

	for _, tri := range mesh.Triangles {
		var vs [3]ebiten.Vertex
		allOffScreen := true

		for i := 0; i < 3; i++ {
			v := tri.V[i]
			litColor := r.Lighting.ComputeLighting(v.Normal, v.Color)

			clip := vp.TransformPoint(v.Pos)
			sx := (clip.X*0.5 + 0.5) * sw
			sy := (1 - (clip.Y*0.5 + 0.5)) * sh

			if sx >= -100 && sx <= sw+100 && sy >= -100 && sy <= sh+100 {
				allOffScreen = false
			}

			vs[i] = ebiten.Vertex{
				DstX:   float32(sx),
				DstY:   float32(sy),
				SrcX:   1,
				SrcY:   1,
				ColorR: float32(litColor.R),
				ColorG: float32(litColor.G),
				ColorB: float32(litColor.B),
				ColorA: 1,
			}
		}

		if allOffScreen {
			continue
		}

		// Back-face culling (screen-space winding order, Y-down)
		ax := vs[1].DstX - vs[0].DstX
		ay := vs[1].DstY - vs[0].DstY
		bx := vs[2].DstX - vs[0].DstX
		by := vs[2].DstY - vs[0].DstY
		cross := ax*by - ay*bx
		if cross < 0.5 {
			continue // back-face or degenerate
		}

		// Add to batch
		base := uint16(len(vertices))
		vertices = append(vertices, vs[0], vs[1], vs[2])
		indices = append(indices, base, base+1, base+2)

		// Flush if approaching uint16 limit
		if len(vertices) >= 65000 {
			screen.DrawTriangles(vertices, indices, r.whiteImg, nil)
			vertices = vertices[:0]
			indices = indices[:0]
		}
	}

	if len(vertices) > 0 {
		screen.DrawTriangles(vertices, indices, r.whiteImg, nil)
	}
}

// DrawGrid draws a grid overlay in 3D space
func (r *Renderer3D) DrawGrid(screen *ebiten.Image, tm *maplib.TileMap) {
	minX, minY, maxX, maxY := r.Camera.VisibleTileRange(tm.Width, tm.Height)
	gridColor := color.RGBA{255, 255, 255, 30}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			sx0, sy0, _ := r.Camera.Project3DToScreen(float64(x), 0.01, float64(y))
			sx1, sy1, _ := r.Camera.Project3DToScreen(float64(x+1), 0.01, float64(y))
			sx2, sy2, _ := r.Camera.Project3DToScreen(float64(x), 0.01, float64(y+1))
			vector.StrokeLine(screen, float32(sx0), float32(sy0), float32(sx1), float32(sy1), 1, gridColor, false)
			vector.StrokeLine(screen, float32(sx0), float32(sy0), float32(sx2), float32(sy2), 1, gridColor, false)
		}
	}
}

// DrawSelectionBox draws a selection rectangle on screen
func (r *Renderer3D) DrawSelectionBox(screen *ebiten.Image, x1, y1, x2, y2 int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	selColor := color.RGBA{0, 255, 0, 128}
	// Fill using semi-transparent rect
	vector.DrawFilledRect(screen, float32(x1), float32(y1), float32(x2-x1), float32(y2-y1), color.RGBA{0, 255, 0, 30}, false)
	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y1), 1, selColor, false)
	vector.StrokeLine(screen, float32(x2), float32(y1), float32(x2), float32(y2), 1, selColor, false)
	vector.StrokeLine(screen, float32(x2), float32(y2), float32(x1), float32(y2), 1, selColor, false)
	vector.StrokeLine(screen, float32(x1), float32(y2), float32(x1), float32(y1), 1, selColor, false)
}

// DrawMinimap draws a top-down minimap
func (r *Renderer3D) DrawMinimap(screen *ebiten.Image, tm *maplib.TileMap, posX, posY, size int) {
	if r.minimapImg == nil || r.minimapImg.Bounds().Dx() != size {
		r.minimapImg = ebiten.NewImage(size, size)
	}
	minimap := r.minimapImg
	minimap.Fill(color.RGBA{0, 0, 0, 180})

	scaleX := float64(size) / float64(tm.Width)
	scaleY := float64(size) / float64(tm.Height)

	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			tile := tm.At(x, y)
			if tile == nil {
				continue
			}
			bc, ok := TerrainBaseColors[tile.Terrain]
			if !ok {
				bc = Color3{0.5, 0.5, 0.5}
			}
			clr := color.RGBA{uint8(bc.R * 255), uint8(bc.G * 255), uint8(bc.B * 255), 255}
			px := float32(float64(x) * scaleX)
			py := float32(float64(y) * scaleY)
			pw := float32(scaleX) + 1
			ph := float32(scaleY) + 1
			vector.DrawFilledRect(minimap, px, py, pw, ph, clr, false)
		}
	}

	// Camera viewport
	wx0, wy0 := r.Camera.ScreenToWorld(0, 0)
	wx1, wy1 := r.Camera.ScreenToWorld(r.Camera.ScreenW, r.Camera.ScreenH)
	vx0, vy0 := float32(wx0*scaleX), float32(wy0*scaleY)
	vx1, vy1 := float32(wx1*scaleX), float32(wy1*scaleY)
	viewColor := color.RGBA{255, 255, 255, 200}
	vector.StrokeLine(minimap, vx0, vy0, vx1, vy0, 1, viewColor, false)
	vector.StrokeLine(minimap, vx1, vy0, vx1, vy1, 1, viewColor, false)
	vector.StrokeLine(minimap, vx1, vy1, vx0, vy1, 1, viewColor, false)
	vector.StrokeLine(minimap, vx0, vy1, vx0, vy0, 1, viewColor, false)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(posX), float64(posY))
	screen.DrawImage(minimap, op)
}

// DrawHealthBar draws a health bar at screen position
func (r *Renderer3D) DrawHealthBar(screen *ebiten.Image, sx, sy int, ratio float64, width int) {
	barH := float32(4)
	barW := float32(width)
	bx := float32(sx) - barW/2
	by := float32(sy) - 5

	// Background
	vector.DrawFilledRect(screen, bx, by, barW, barH, color.RGBA{40, 40, 40, 200}, false)

	// Health fill
	var hc color.RGBA
	if ratio > 0.6 {
		hc = color.RGBA{0, 200, 0, 255}
	} else if ratio > 0.3 {
		hc = color.RGBA{255, 200, 0, 255}
	} else {
		hc = color.RGBA{255, 0, 0, 255}
	}
	vector.DrawFilledRect(screen, bx, by, barW*float32(ratio), barH, hc, false)
}

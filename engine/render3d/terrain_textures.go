package render3d

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/1siamBot/rts-engine/engine/maplib"
	"github.com/hajimehoshi/ebiten/v2"
)

// TerrainTextureAtlas holds loaded terrain tile images grouped by type
type TerrainTextureAtlas struct {
	tiles   map[string][]*ebiten.Image // e.g. "grass" -> [grass_1, grass_2, ...]
	loaded  bool
}

// NewTerrainTextureAtlas creates a new atlas
func NewTerrainTextureAtlas() *TerrainTextureAtlas {
	return &TerrainTextureAtlas{
		tiles: make(map[string][]*ebiten.Image),
	}
}

// LoadFromDirectory loads all terrain PNGs from the terrain directory
func (ta *TerrainTextureAtlas) LoadFromDirectory(terrainDir string) {
	entries, err := os.ReadDir(terrainDir)
	if err != nil {
		return
	}

	// Group by prefix (e.g. grass_1.png -> "grass")
	grouped := make(map[string][]string)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".png" {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".png")
		// Find prefix: everything before the last _N
		lastUnderscore := strings.LastIndex(name, "_")
		if lastUnderscore < 0 {
			continue
		}
		prefix := name[:lastUnderscore]
		grouped[prefix] = append(grouped[prefix], filepath.Join(terrainDir, e.Name()))
	}

	total := 0
	for prefix, paths := range grouped {
		sort.Strings(paths)
		for _, p := range paths {
			img := loadTerrainImage(p)
			if img != nil {
				ta.tiles[prefix] = append(ta.tiles[prefix], img)
				total++
			}
		}
	}

	if total > 0 {
		ta.loaded = true
		fmt.Printf("TerrainTextureAtlas: loaded %d tiles (%d types)\n", total, len(ta.tiles))
	}
}

// GetTile returns a terrain tile image for the given terrain type and position
// Uses position hash for deterministic variant selection
func (ta *TerrainTextureAtlas) GetTile(terrain maplib.TerrainType, x, y int) *ebiten.Image {
	prefix := terrainTypeToPrefix(terrain)
	tiles := ta.tiles[prefix]
	if len(tiles) == 0 {
		// Fallback to grass
		tiles = ta.tiles["grass"]
		if len(tiles) == 0 {
			return nil
		}
	}
	// Deterministic hash for variant selection
	hash := uint32(x*73856093 ^ y*19349663)
	idx := int(hash) % len(tiles)
	if idx < 0 {
		idx = -idx
	}
	return tiles[idx]
}

// GetWaterTile returns a water tile with animation frame based on time
func (ta *TerrainTextureAtlas) GetWaterTile(x, y int, time float64) *ebiten.Image {
	tiles := ta.tiles["water"]
	if len(tiles) == 0 {
		return nil
	}
	// Cycle through water variants for animation
	// Each tile has a phase offset based on position
	phase := time*0.5 + float64(x*7+y*13)*0.1
	frameIdx := int(math.Floor(phase)) % len(tiles)
	if frameIdx < 0 {
		frameIdx += len(tiles)
	}
	return tiles[frameIdx]
}

func terrainTypeToPrefix(t maplib.TerrainType) string {
	switch t {
	case maplib.TerrainGrass:
		return "grass"
	case maplib.TerrainDirt:
		return "dirt"
	case maplib.TerrainSand:
		return "sand"
	case maplib.TerrainWater, maplib.TerrainDeepWater:
		return "water"
	case maplib.TerrainRock:
		return "rock"
	case maplib.TerrainCliff:
		return "cliff"
	case maplib.TerrainRoad:
		return "road"
	case maplib.TerrainBridge:
		return "bridge"
	case maplib.TerrainOre:
		return "ore"
	case maplib.TerrainGem:
		return "gem"
	case maplib.TerrainSnow:
		return "snow"
	case maplib.TerrainUrban:
		return "urban"
	case maplib.TerrainForest:
		return "grass_dark"
	default:
		return "grass"
	}
}

// RenderTexturedTerrain draws textured terrain tiles using DrawTriangles
func (ta *TerrainTextureAtlas) RenderTexturedTerrain(
	screen *ebiten.Image,
	cam *Camera3D,
	tm *maplib.TileMap,
	minX, minY, maxX, maxY int,
	time float64,
) {
	if !ta.loaded {
		return
	}

	sw := float64(cam.ScreenW)
	sh := float64(cam.ScreenH)
	vp := cam.ViewProj()

	// First pass: render a base color fill to eliminate gaps between textured tiles
	ta.renderBaseColorFill(screen, cam, tm, minX, minY, maxX, maxY, vp, sw, sh)

	// Group tiles by texture for batched rendering
	type tileBatch struct {
		vertices []ebiten.Vertex
		indices  []uint16
	}
	batches := make(map[*ebiten.Image]*tileBatch)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			tile := tm.At(x, y)
			if tile == nil {
				continue
			}

			var tex *ebiten.Image
			if tile.Terrain == maplib.TerrainWater || tile.Terrain == maplib.TerrainDeepWater {
				tex = ta.GetWaterTile(x, y, time)
			} else {
				tex = ta.GetTile(tile.Terrain, x, y)
			}
			if tex == nil {
				continue
			}

			// Height for this tile
			h := float64(tile.Height) * 0.15

			// Project 4 corners of the tile with slight overlap to avoid gaps
			fx, fz := float64(x), float64(y)
			const pad = 0.015 // slight overlap to fully cover base fill

			// For water, add slight wave
			h0, h1, h2, h3 := h, h, h, h
			if tile.Terrain == maplib.TerrainWater || tile.Terrain == maplib.TerrainDeepWater {
				h0 = -0.05 + 0.03*math.Sin(time*2.0+fx*0.5+fz*0.7)
				h1 = -0.05 + 0.03*math.Sin(time*2.0+(fx+1)*0.5+fz*0.7)
				h2 = -0.05 + 0.03*math.Sin(time*2.0+(fx+1)*0.5+(fz+1)*0.7)
				h3 = -0.05 + 0.03*math.Sin(time*2.0+fx*0.5+(fz+1)*0.7)
			}

			p0 := vp.TransformPoint(V3(fx-pad, h0, fz-pad))
			p1 := vp.TransformPoint(V3(fx+1+pad, h1, fz-pad))
			p2 := vp.TransformPoint(V3(fx+1+pad, h2, fz+1+pad))
			p3 := vp.TransformPoint(V3(fx-pad, h3, fz+1+pad))

			s0x := float32((p0.X*0.5 + 0.5) * sw)
			s0y := float32((1 - (p0.Y*0.5 + 0.5)) * sh)
			s1x := float32((p1.X*0.5 + 0.5) * sw)
			s1y := float32((1 - (p1.Y*0.5 + 0.5)) * sh)
			s2x := float32((p2.X*0.5 + 0.5) * sw)
			s2y := float32((1 - (p2.Y*0.5 + 0.5)) * sh)
			s3x := float32((p3.X*0.5 + 0.5) * sw)
			s3y := float32((1 - (p3.Y*0.5 + 0.5)) * sh)

			// Frustum cull
			if s0x < -200 && s1x < -200 && s2x < -200 && s3x < -200 {
				continue
			}
			if s0x > float32(sw)+200 && s1x > float32(sw)+200 && s2x > float32(sw)+200 && s3x > float32(sw)+200 {
				continue
			}
			if s0y < -200 && s1y < -200 && s2y < -200 && s3y < -200 {
				continue
			}
			if s0y > float32(sh)+200 && s1y > float32(sh)+200 && s2y > float32(sh)+200 && s3y > float32(sh)+200 {
				continue
			}

			texW := float32(tex.Bounds().Dx())
			texH := float32(tex.Bounds().Dy())

			// Color tint (white = no tint, can add lighting)
			cr, cg, cb := float32(1.0), float32(1.0), float32(1.0)

			// Slight per-tile color variation for natural look
			hash := uint32(x*7919 + y*7927)
			variation := float32(hash%100) / 100.0 * 0.06 - 0.03
			cr += variation
			cg += variation
			cb += variation

			// Water tint effect
			if tile.Terrain == maplib.TerrainWater || tile.Terrain == maplib.TerrainDeepWater {
				spec := float32(0.1 * math.Abs(math.Sin(time*1.8+float64(x)*0.4)))
				cb += spec * 0.3
				cg += spec * 0.1
			}

			// Ore sparkle overlay
			if tile.OreAmount > 0 && tile.Terrain != maplib.TerrainOre {
				sparkle := float32(0.1 * math.Abs(math.Sin(time*3+float64(x*7+y*13))))
				cr += sparkle * 0.3
				cg += sparkle * 0.2
			}

			batch, ok := batches[tex]
			if !ok {
				batch = &tileBatch{}
				batches[tex] = batch
			}

			base := uint16(len(batch.vertices))
			batch.vertices = append(batch.vertices,
				ebiten.Vertex{DstX: s0x, DstY: s0y, SrcX: 0, SrcY: 0, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: 1},
				ebiten.Vertex{DstX: s1x, DstY: s1y, SrcX: texW, SrcY: 0, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: 1},
				ebiten.Vertex{DstX: s2x, DstY: s2y, SrcX: texW, SrcY: texH, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: 1},
				ebiten.Vertex{DstX: s3x, DstY: s3y, SrcX: 0, SrcY: texH, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: 1},
			)
			batch.indices = append(batch.indices, base, base+1, base+2, base, base+2, base+3)

			// Flush if approaching uint16 limit
			if len(batch.vertices) >= 64000 {
				op := &ebiten.DrawTrianglesOptions{}
				op.AntiAlias = false
				screen.DrawTriangles(batch.vertices, batch.indices, tex, op)
				batch.vertices = batch.vertices[:0]
				batch.indices = batch.indices[:0]
			}
		}
	}

	// Flush all batches
	for tex, batch := range batches {
		if len(batch.vertices) > 0 {
			op := &ebiten.DrawTrianglesOptions{}
			op.AntiAlias = false
			screen.DrawTriangles(batch.vertices, batch.indices, tex, op)
		}
	}
}

// RenderTreeBillboards draws tree sprites on forest tiles
func (ta *TerrainTextureAtlas) RenderTreeBillboards(
	screen *ebiten.Image,
	cam *Camera3D,
	tm *maplib.TileMap,
	sprites *SpriteAtlas,
	minX, minY, maxX, maxY int,
) {
	// Use grass_dark tiles as tree canopy sprites if no dedicated tree sprite
	treeTiles := ta.tiles["grass_dark"]
	if len(treeTiles) == 0 {
		return
	}

	type treeDraw struct {
		x, y  float64
		depth float64
		hash  uint32
	}
	var trees []treeDraw

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			tile := tm.At(x, y)
			if tile == nil || tile.Terrain != maplib.TerrainForest {
				continue
			}
			h := float64(tile.Height) * 0.15
			_, _, depth := cam.Project3DToScreen(float64(x)+0.5, h+0.4, float64(y)+0.5)
			hash := uint32(x*73856093 ^ y*19349663)
			trees = append(trees, treeDraw{float64(x) + 0.5, float64(y) + 0.5, depth, hash})
		}
	}

	// Sort back to front
	sort.Slice(trees, func(i, j int) bool {
		return trees[i].depth > trees[j].depth
	})

	for _, t := range trees {
		// Draw a tree billboard: trunk color block + canopy
		h := 0.0 // base height
		treeH := 0.3 + float64(t.hash%100)/400.0
		canopyScale := 0.8 + float64(t.hash%80)/200.0

		sx, sy, _ := cam.Project3DToScreen(t.x, h+treeH, t.y)

		// Dark trunk (small rect)
		pixelsPerUnit := float64(cam.ScreenW) / cam.Zoom
		trunkW := int(0.08 * pixelsPerUnit)
		trunkH := int(treeH * pixelsPerUnit * 0.4)
		if trunkW < 2 {
			trunkW = 2
		}
		if trunkH < 3 {
			trunkH = 3
		}

		// Canopy (use a green-tinted circle/sprite)
		canopyR := int(canopyScale * pixelsPerUnit * 0.25)
		if canopyR < 4 {
			canopyR = 4
		}

		// Draw canopy as a simple colored oval using the tree tile
		idx := int(t.hash) % len(treeTiles)
		if idx < 0 {
			idx = -idx
		}
		treeTex := treeTiles[idx]

		// Scale tree texture to canopy size
		op := &ebiten.DrawImageOptions{}
		tw := float64(treeTex.Bounds().Dx())
		th := float64(treeTex.Bounds().Dy())
		targetW := float64(canopyR) * 2
		targetH := float64(canopyR) * 1.5 // slightly squished for iso perspective
		op.GeoM.Scale(targetW/tw, targetH/th)
		op.GeoM.Translate(float64(sx)-targetW/2, float64(sy)-targetH-float64(trunkH)/2)
		// Green tint
		op.ColorScale.Scale(0.7, 1.1, 0.6, 1.0)
		screen.DrawImage(treeTex, op)
		_ = trunkW
	}
}

// renderBaseColorFill draws a solid color underneath each tile to prevent gap artifacts
func (ta *TerrainTextureAtlas) renderBaseColorFill(
	screen *ebiten.Image,
	cam *Camera3D,
	tm *maplib.TileMap,
	minX, minY, maxX, maxY int,
	vp Mat4,
	sw, sh float64,
) {
	// Use a small white image for vertex-colored triangles
	whiteImg := ebiten.NewImage(2, 2)
	whiteImg.Fill(color.White)

	var vertices []ebiten.Vertex
	var indices []uint16

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			tile := tm.At(x, y)
			if tile == nil {
				continue
			}

			h := float64(tile.Height) * 0.15
			fx, fz := float64(x), float64(y)

			p0 := vp.TransformPoint(V3(fx, h, fz))
			p1 := vp.TransformPoint(V3(fx+1, h, fz))
			p2 := vp.TransformPoint(V3(fx+1, h, fz+1))
			p3 := vp.TransformPoint(V3(fx, h, fz+1))

			s0x := float32((p0.X*0.5 + 0.5) * sw)
			s0y := float32((1 - (p0.Y*0.5 + 0.5)) * sh)
			s1x := float32((p1.X*0.5 + 0.5) * sw)
			s1y := float32((1 - (p1.Y*0.5 + 0.5)) * sh)
			s2x := float32((p2.X*0.5 + 0.5) * sw)
			s2y := float32((1 - (p2.Y*0.5 + 0.5)) * sh)
			s3x := float32((p3.X*0.5 + 0.5) * sw)
			s3y := float32((1 - (p3.Y*0.5 + 0.5)) * sh)

			// Base color from terrain type (darkened to serve as gap fill)
			bc, ok := TerrainBaseColors[tile.Terrain]
			if !ok {
				bc = Color3{0.3, 0.5, 0.2}
			}
			cr, cg, cb := float32(bc.R*0.7), float32(bc.G*0.7), float32(bc.B*0.7)

			base := uint16(len(vertices))
			vertices = append(vertices,
				ebiten.Vertex{DstX: s0x, DstY: s0y, SrcX: 0, SrcY: 0, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: 1},
				ebiten.Vertex{DstX: s1x, DstY: s1y, SrcX: 1, SrcY: 0, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: 1},
				ebiten.Vertex{DstX: s2x, DstY: s2y, SrcX: 1, SrcY: 1, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: 1},
				ebiten.Vertex{DstX: s3x, DstY: s3y, SrcX: 0, SrcY: 1, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: 1},
			)
			indices = append(indices, base, base+1, base+2, base, base+2, base+3)

			if len(vertices) >= 64000 {
				screen.DrawTriangles(vertices, indices, whiteImg, nil)
				vertices = vertices[:0]
				indices = indices[:0]
			}
		}
	}

	if len(vertices) > 0 {
		screen.DrawTriangles(vertices, indices, whiteImg, nil)
	}
}

func loadTerrainImage(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		f.Seek(0, 0)
		img, _, err = image.Decode(f)
		if err != nil {
			return nil
		}
	}
	return ebiten.NewImageFromImage(img)
}

package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/1siamBot/rts-engine/engine/core"
	"github.com/1siamBot/rts-engine/engine/input"
	"github.com/1siamBot/rts-engine/engine/maplib"
	"github.com/1siamBot/rts-engine/engine/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720
	TickRate     = 20.0 // 20 ticks per second (RTS standard)
	MapSize      = 64
)

// Game implements ebiten.Game interface
type Game struct {
	renderer  *render.IsoRenderer
	tileMap   *maplib.TileMap
	gameLoop  *core.GameLoop
	input     *input.InputState
	players   *core.PlayerManager
	eventBus  *core.EventBus

	// Demo entities
	units     []demoUnit

	// UI state
	showGrid   bool
	showMinimap bool
	hoverTileX int
	hoverTileY int
}

type demoUnit struct {
	id   core.EntityID
	x, y float64
	selected bool
}

func NewGame() *Game {
	g := &Game{
		renderer:  render.NewIsoRenderer(ScreenWidth, ScreenHeight),
		tileMap:   generateDemoMap(),
		gameLoop:  core.NewGameLoop(TickRate),
		input:     input.NewInputState(),
		players:   core.NewPlayerManager(),
		eventBus:  core.NewEventBus(),
		showGrid:  false,
		showMinimap: true,
	}

	// Set up players
	g.players.AddPlayer(&core.Player{
		ID: 0, Name: "Player 1", TeamID: 0, Faction: "Allied",
		Color: 0x0066FFFF, Credits: 10000,
	})
	g.players.AddPlayer(&core.Player{
		ID: 1, Name: "AI Enemy", TeamID: 1, Faction: "Soviet",
		Color: 0xFF0000FF, Credits: 10000, IsAI: true,
	})

	// Center camera on map
	g.renderer.Camera.CenterOn(float64(MapSize)/2, float64(MapSize)/2)

	// Spawn demo units
	g.spawnDemoUnits()

	// Start the game
	g.gameLoop.Play()

	return g
}

func (g *Game) spawnDemoUnits() {
	positions := [][2]float64{
		{10, 10}, {11, 10}, {12, 10},
		{10, 11}, {11, 11},
	}
	for _, pos := range positions {
		id := g.gameLoop.World.Spawn()
		g.gameLoop.World.Attach(id, &core.Position{X: pos[0], Y: pos[1]})
		g.gameLoop.World.Attach(id, &core.Sprite{
			Width: 24, Height: 24, Visible: true, ScaleX: 1, ScaleY: 1,
		})
		g.gameLoop.World.Attach(id, &core.Health{Current: 100, Max: 100})
		g.gameLoop.World.Attach(id, &core.Movable{Speed: 3.0, MoveType: core.MoveVehicle})
		g.gameLoop.World.Attach(id, &core.Selectable{Radius: 0.5})
		g.gameLoop.World.Attach(id, &core.Owner{PlayerID: 0})
		g.gameLoop.World.Attach(id, &core.FogVision{Range: 5})
		g.units = append(g.units, demoUnit{id: id, x: pos[0], y: pos[1]})
	}
}

func (g *Game) Update() error {
	g.input.Update()

	// Camera controls
	g.handleCamera()

	// Toggle grid
	if g.input.IsKeyJustPressed(ebiten.KeyG) {
		g.showGrid = !g.showGrid
	}
	// Toggle minimap
	if g.input.IsKeyJustPressed(ebiten.KeyM) {
		g.showMinimap = !g.showMinimap
	}

	// Track hover tile
	wx, wy := g.renderer.Camera.ScreenToWorld(g.input.MouseX, g.input.MouseY)
	g.hoverTileX = int(math.Floor(wx))
	g.hoverTileY = int(math.Floor(wy))

	// Handle selection
	if g.input.RightJustPressed {
		// Move selected units to clicked position
		for i := range g.units {
			if g.units[i].selected {
				g.units[i].x = math.Floor(wx) + 0.5
				g.units[i].y = math.Floor(wy) + 0.5
				// Update ECS position
				if pos := g.gameLoop.World.Get(g.units[i].id, core.CompPosition); pos != nil {
					p := pos.(*core.Position)
					p.X = g.units[i].x
					p.Y = g.units[i].y
				}
			}
		}
	}

	if g.input.LeftJustReleased && !g.input.Dragging {
		// Click select
		shift := ebiten.IsKeyPressed(ebiten.KeyShift)
		if !shift {
			for i := range g.units {
				g.units[i].selected = false
			}
		}
		for i := range g.units {
			sx, sy := g.renderer.Camera.WorldToScreen(g.units[i].x, g.units[i].y)
			dx := float64(g.input.MouseX - sx)
			dy := float64(g.input.MouseY - sy)
			if math.Sqrt(dx*dx+dy*dy) < 20 {
				g.units[i].selected = !g.units[i].selected
				break
			}
		}
	}

	// Game simulation tick
	g.gameLoop.Update()
	g.eventBus.Dispatch()

	return nil
}

func (g *Game) handleCamera() {
	speed := g.renderer.Camera.Speed / 60.0 // per frame at 60fps

	// WASD / Arrow keys
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.renderer.Camera.Pan(0, -speed)
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.renderer.Camera.Pan(0, speed)
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.renderer.Camera.Pan(-speed, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.renderer.Camera.Pan(speed, 0)
	}

	// Edge scrolling
	if g.renderer.Camera.EdgeScroll {
		edge := g.renderer.Camera.EdgeSize
		if g.input.MouseX < edge {
			g.renderer.Camera.Pan(-speed, 0)
		}
		if g.input.MouseX > ScreenWidth-edge {
			g.renderer.Camera.Pan(speed, 0)
		}
		if g.input.MouseY < edge {
			g.renderer.Camera.Pan(0, -speed)
		}
		if g.input.MouseY > ScreenHeight-edge {
			g.renderer.Camera.Pan(0, speed)
		}
	}

	// Zoom with scroll wheel
	if g.input.ScrollY != 0 {
		g.renderer.Camera.ZoomAt(g.input.ScrollY*0.1, g.input.MouseX, g.input.MouseY)
	}

	// Middle mouse drag to pan
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		g.renderer.Camera.Pan(float64(-g.input.MouseDX), float64(-g.input.MouseDY))
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen with dark background
	screen.Fill(color.RGBA{20, 20, 30, 255})

	// Draw map tiles
	g.renderer.DrawMap(screen, g.tileMap)

	// Draw grid overlay
	if g.showGrid {
		g.renderer.DrawGrid(screen, g.tileMap)
	}

	// Draw hover tile highlight
	if g.tileMap.InBounds(g.hoverTileX, g.hoverTileY) {
		sx, sy := g.renderer.Camera.WorldToScreen(float64(g.hoverTileX), float64(g.hoverTileY))
		tw := float32(g.tileMap.TileWidth)
		th := float32(g.tileMap.TileHeight)
		hw := tw / 2
		hh := th / 2
		cx := float32(sx)
		cy := float32(sy) + hh
		hoverColor := color.RGBA{255, 255, 0, 100}
		vector.StrokeLine(screen, cx, cy-hh, cx+hw, cy, 2, hoverColor, false)
		vector.StrokeLine(screen, cx+hw, cy, cx, cy+hh, 2, hoverColor, false)
		vector.StrokeLine(screen, cx, cy+hh, cx-hw, cy, 2, hoverColor, false)
		vector.StrokeLine(screen, cx-hw, cy, cx, cy-hh, 2, hoverColor, false)
	}

	// Draw units
	for _, u := range g.units {
		sx, sy := g.renderer.Camera.WorldToScreen(u.x, u.y)

		// Unit body (circle)
		if u.selected {
			// Selection ring
			vector.DrawFilledCircle(screen, float32(sx), float32(sy), 16, color.RGBA{0, 255, 0, 60}, false)
			vector.StrokeCircle(screen, float32(sx), float32(sy), 16, 2, color.RGBA{0, 255, 0, 200}, false)
		}

		// Unit color (player 0 = blue)
		unitColor := color.RGBA{60, 120, 255, 255}
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), 10, unitColor, false)
		vector.StrokeCircle(screen, float32(sx), float32(sy), 10, 1, color.RGBA{255, 255, 255, 180}, false)

		// Health bar
		if u.selected {
			barW := float32(24)
			barH := float32(3)
			barX := float32(sx) - barW/2
			barY := float32(sy) - 22
			// Background
			vector.DrawFilledRect(screen, barX, barY, barW, barH, color.RGBA{40, 40, 40, 200}, false)
			// Health fill
			vector.DrawFilledRect(screen, barX, barY, barW, barH, color.RGBA{0, 200, 0, 255}, false)
		}
	}

	// Draw selection box
	if x1, y1, x2, y2, active := g.input.DragRect(); active {
		g.renderer.DrawSelectionBox(screen, x1, y1, x2, y2)
	}

	// Draw minimap
	if g.showMinimap {
		g.renderer.DrawMinimap(screen, g.tileMap, ScreenWidth-170, ScreenHeight-170, 160)
	}

	// HUD overlay
	g.drawHUD(screen)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Top-left info
	tile := g.tileMap.At(g.hoverTileX, g.hoverTileY)
	terrainName := "Out of Bounds"
	if tile != nil {
		terrainName = terrainTypeName(tile.Terrain)
	}

	info := fmt.Sprintf(
		"RTS Engine v0.1.0 | FPS: %.0f | Tick: %d\n"+
		"Tile: (%d, %d) %s | Entities: %d\n"+
		"Zoom: %.1fx | [WASD] Pan [Scroll] Zoom [G] Grid [M] Minimap\n"+
		"[LClick] Select [RClick] Move | Credits: $%d",
		ebiten.ActualFPS(),
		g.gameLoop.CurrentTick(),
		g.hoverTileX, g.hoverTileY, terrainName,
		g.gameLoop.World.EntityCount(),
		g.renderer.Camera.Zoom,
		g.players.GetPlayer(0).Credits,
	)

	ebitenutil.DebugPrint(screen, info)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// generateDemoMap creates a demo map with varied terrain
func generateDemoMap() *maplib.TileMap {
	tm := maplib.NewTileMap("Demo Battlefield", MapSize, MapSize)

	// Base: grass
	tm.SetTerrain(0, 0, MapSize-1, MapSize-1, maplib.TerrainGrass)

	// River through the middle
	for x := 0; x < MapSize; x++ {
		y := MapSize/2 + int(3*math.Sin(float64(x)*0.15))
		tm.SetTerrain(x, y-1, x, y+1, maplib.TerrainWater)
	}

	// Bridge
	tm.SetTerrain(MapSize/2-1, MapSize/2-2, MapSize/2+1, MapSize/2+2, maplib.TerrainBridge)
	for x := MapSize/2 - 1; x <= MapSize/2+1; x++ {
		for y := MapSize/2 - 2; y <= MapSize/2+2; y++ {
			if t := tm.At(x, y); t != nil {
				t.Passable = maplib.PassAll
			}
		}
	}

	// Forest patches
	forests := [][4]int{
		{5, 5, 12, 10}, {45, 8, 55, 15}, {20, 45, 30, 52},
	}
	for _, f := range forests {
		tm.SetTerrain(f[0], f[1], f[2], f[3], maplib.TerrainForest)
	}

	// Ore fields
	orePositions := [][2]int{
		{15, 15}, {16, 15}, {15, 16}, {16, 16}, {17, 15},
		{45, 45}, {46, 45}, {45, 46}, {46, 46}, {47, 45},
	}
	for _, pos := range orePositions {
		tm.PlaceOre(pos[0], pos[1], 1000)
	}

	// Cliffs
	tm.SetTerrain(30, 10, 35, 12, maplib.TerrainCliff)
	tm.SetTerrain(25, 50, 28, 55, maplib.TerrainRock)

	// Roads
	for x := 0; x < MapSize; x++ {
		tm.SetTerrain(x, MapSize/4, x, MapSize/4, maplib.TerrainRoad)
	}
	for y := 0; y < MapSize; y++ {
		tm.SetTerrain(MapSize/4, y, MapSize/4, y, maplib.TerrainRoad)
	}

	// Sand areas
	tm.SetTerrain(50, 50, 60, 60, maplib.TerrainSand)

	// Start positions
	tm.StartPositions = []maplib.StartPos{
		{PlayerSlot: 0, X: 5, Y: 5},
		{PlayerSlot: 1, X: MapSize - 10, Y: MapSize - 10},
	}

	return tm
}

func terrainTypeName(t maplib.TerrainType) string {
	names := map[maplib.TerrainType]string{
		maplib.TerrainGrass:     "Grass",
		maplib.TerrainDirt:      "Dirt",
		maplib.TerrainSand:      "Sand",
		maplib.TerrainWater:     "Water",
		maplib.TerrainDeepWater: "Deep Water",
		maplib.TerrainRock:      "Rock",
		maplib.TerrainCliff:     "Cliff",
		maplib.TerrainRoad:      "Road",
		maplib.TerrainBridge:    "Bridge",
		maplib.TerrainOre:       "Ore Field",
		maplib.TerrainGem:       "Gem Field",
		maplib.TerrainSnow:      "Snow",
		maplib.TerrainUrban:     "Urban",
		maplib.TerrainForest:    "Forest",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return "Unknown"
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("⚔️ RTS Engine v0.1.0 — Phase 1: Foundation")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

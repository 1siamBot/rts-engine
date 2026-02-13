package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/1siamBot/rts-engine/engine/ai"
	"github.com/1siamBot/rts-engine/engine/audio"
	"github.com/1siamBot/rts-engine/engine/core"
	"github.com/1siamBot/rts-engine/engine/input"
	"github.com/1siamBot/rts-engine/engine/maplib"
	"github.com/1siamBot/rts-engine/engine/pathfind"
	"github.com/1siamBot/rts-engine/engine/render"
	"github.com/1siamBot/rts-engine/engine/systems"
	"github.com/1siamBot/rts-engine/engine/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720
	TickRate     = 20.0
	MapSize      = 64
)

// Game implements ebiten.Game
type Game struct {
	renderer *render.IsoRenderer
	tileMap  *maplib.TileMap
	gameLoop *core.GameLoop
	input    *input.InputState
	players  *core.PlayerManager
	eventBus *core.EventBus
	navGrid  *pathfind.NavGrid
	techTree *systems.TechTree
	hud      *ui.HUD
	audioMgr *audio.AudioManager
	fogSys   *systems.FogSystem

	// State
	showGrid    bool
	showMinimap bool
	hoverTileX  int
	hoverTileY  int
	gameState   string // "menu", "playing", "paused", "gameover"

	// Settings
	scrollSpeed float64
}

func NewGame() *Game {
	g := &Game{
		renderer:    render.NewIsoRenderer(ScreenWidth, ScreenHeight),
		tileMap:     generateDemoMap(),
		gameLoop:    core.NewGameLoop(TickRate),
		input:       input.NewInputState(),
		players:     core.NewPlayerManager(),
		eventBus:    core.NewEventBus(),
		techTree:    systems.NewTechTree(),
		audioMgr:    audio.NewAudioManager(),
		showMinimap: true,
		gameState:   "playing",
		scrollSpeed: 500,
	}

	// Players
	g.players.AddPlayer(&core.Player{
		ID: 0, Name: "Player 1", TeamID: 0, Faction: "Allied",
		Color: 0x0066FFFF, Credits: 10000,
	})
	g.players.AddPlayer(&core.Player{
		ID: 1, Name: "AI Enemy", TeamID: 1, Faction: "Soviet",
		Color: 0xFF0000FF, Credits: 10000, IsAI: true,
	})

	// Nav grid
	g.navGrid = pathfind.NewNavGrid(g.tileMap)

	// HUD
	g.hud = ui.NewHUD(ScreenWidth, ScreenHeight, g.techTree, g.players, 0)

	// Fog of war
	g.fogSys = systems.NewFogSystem(g.tileMap.Width, g.tileMap.Height, g.players)

	// Register systems
	w := g.gameLoop.World
	w.AddSystem(&systems.PowerSystem{Players: g.players})
	w.AddSystem(g.fogSys)
	w.AddSystem(&systems.MovementSystem{NavGrid: g.navGrid})
	w.AddSystem(&systems.CombatSystem{EventBus: g.eventBus, Players: g.players})
	w.AddSystem(&systems.ProjectileSystem{EventBus: g.eventBus})
	w.AddSystem(&systems.HarvesterSystem{NavGrid: g.navGrid, TileMap: g.tileMap, Players: g.players, EventBus: g.eventBus})
	w.AddSystem(&systems.ProductionSystem{TechTree: g.techTree, Players: g.players, EventBus: g.eventBus})
	w.AddSystem(&systems.AnimationSystem{})
	w.AddSystem(&systems.GameOverSystem{Players: g.players})
	w.AddSystem(&ai.AISystem{
		Controllers: []*ai.AIController{
			ai.NewAIController(1, ai.DiffMedium, g.techTree, g.navGrid),
		},
		Players: g.players,
	})

	// Center camera
	g.renderer.Camera.CenterOn(float64(MapSize)/2, float64(MapSize)/2)

	// Spawn initial entities
	g.spawnInitialEntities()

	g.gameLoop.Play()
	return g
}

func (g *Game) spawnInitialEntities() {
	w := g.gameLoop.World

	// Player 0: Construction Yard + units
	cyID := w.Spawn()
	w.Attach(cyID, &core.Position{X: 10, Y: 10})
	w.Attach(cyID, &core.Health{Current: 1000, Max: 1000})
	w.Attach(cyID, &core.Building{SizeX: 3, SizeY: 3, PowerGen: 0})
	w.Attach(cyID, &core.Production{Rate: 1.0, Rally: core.TilePos{X: 13, Y: 13}})
	w.Attach(cyID, &core.Owner{PlayerID: 0, Faction: "Allied"})
	w.Attach(cyID, &core.FogVision{Range: 8})
	w.Attach(cyID, &core.Selectable{Radius: 1.5})

	// Player 0: Power Plant
	ppID := w.Spawn()
	w.Attach(ppID, &core.Position{X: 14, Y: 10})
	w.Attach(ppID, &core.Health{Current: 750, Max: 750})
	w.Attach(ppID, &core.Building{SizeX: 2, SizeY: 2, PowerGen: 100})
	w.Attach(ppID, &core.Owner{PlayerID: 0, Faction: "Allied"})
	w.Attach(ppID, &core.FogVision{Range: 5})

	// Player 0: Barracks with production
	barID := w.Spawn()
	w.Attach(barID, &core.Position{X: 10, Y: 14})
	w.Attach(barID, &core.Health{Current: 500, Max: 500})
	w.Attach(barID, &core.Building{SizeX: 2, SizeY: 2, PowerDraw: 20})
	w.Attach(barID, &core.Production{Rate: 1.0, Rally: core.TilePos{X: 12, Y: 16}})
	w.Attach(barID, &core.Owner{PlayerID: 0, Faction: "Allied"})
	w.Attach(barID, &core.FogVision{Range: 5})
	w.Attach(barID, &core.Selectable{Radius: 1.0})

	// Player 0: Refinery
	refID := w.Spawn()
	w.Attach(refID, &core.Position{X: 14, Y: 14})
	w.Attach(refID, &core.Health{Current: 900, Max: 900})
	w.Attach(refID, &core.Building{SizeX: 3, SizeY: 3, PowerDraw: 30})
	w.Attach(refID, &core.Owner{PlayerID: 0, Faction: "Allied"})
	w.Attach(refID, &core.FogVision{Range: 5})

	// Player 0: Starting units
	for i := 0; i < 5; i++ {
		uid := w.Spawn()
		w.Attach(uid, &core.Position{X: float64(8 + i), Y: 13})
		w.Attach(uid, &core.Sprite{Width: 24, Height: 24, Visible: true, ScaleX: 1, ScaleY: 1})
		w.Attach(uid, &core.Health{Current: 125, Max: 125})
		w.Attach(uid, &core.Movable{Speed: 3.0, MoveType: core.MoveInfantry})
		w.Attach(uid, &core.Weapon{Name: "Rifle", Damage: 15, Range: 5, Cooldown: 1.0, DamageType: core.DmgKinetic, TargetType: core.TargetAll})
		w.Attach(uid, &core.Armor{ArmorType: core.ArmorLight})
		w.Attach(uid, &core.Selectable{Radius: 0.5})
		w.Attach(uid, &core.Owner{PlayerID: 0, Faction: "Allied"})
		w.Attach(uid, &core.FogVision{Range: 5})
	}

	// Player 0: Harvester
	harvID := w.Spawn()
	w.Attach(harvID, &core.Position{X: 15, Y: 16})
	w.Attach(harvID, &core.Sprite{Width: 28, Height: 28, Visible: true, ScaleX: 1, ScaleY: 1})
	w.Attach(harvID, &core.Health{Current: 600, Max: 600})
	w.Attach(harvID, &core.Movable{Speed: 1.5, MoveType: core.MoveVehicle})
	w.Attach(harvID, &core.Harvester{Capacity: 20, Rate: 2.0, Resource: "ore"})
	w.Attach(harvID, &core.Selectable{Radius: 0.6})
	w.Attach(harvID, &core.Owner{PlayerID: 0, Faction: "Allied"})
	w.Attach(harvID, &core.FogVision{Range: 4})

	// ---- AI Player 1 ----
	aicyID := w.Spawn()
	w.Attach(aicyID, &core.Position{X: 54, Y: 54})
	w.Attach(aicyID, &core.Health{Current: 1000, Max: 1000})
	w.Attach(aicyID, &core.Building{SizeX: 3, SizeY: 3, PowerGen: 0})
	w.Attach(aicyID, &core.Production{Rate: 1.0, Rally: core.TilePos{X: 52, Y: 52}})
	w.Attach(aicyID, &core.Owner{PlayerID: 1, Faction: "Soviet"})
	w.Attach(aicyID, &core.FogVision{Range: 8})

	aippID := w.Spawn()
	w.Attach(aippID, &core.Position{X: 50, Y: 54})
	w.Attach(aippID, &core.Health{Current: 750, Max: 750})
	w.Attach(aippID, &core.Building{SizeX: 2, SizeY: 2, PowerGen: 100})
	w.Attach(aippID, &core.Owner{PlayerID: 1, Faction: "Soviet"})
	w.Attach(aippID, &core.FogVision{Range: 5})

	aibarID := w.Spawn()
	w.Attach(aibarID, &core.Position{X: 54, Y: 50})
	w.Attach(aibarID, &core.Health{Current: 500, Max: 500})
	w.Attach(aibarID, &core.Building{SizeX: 2, SizeY: 2, PowerDraw: 20})
	w.Attach(aibarID, &core.Production{Rate: 1.0, Rally: core.TilePos{X: 52, Y: 48}})
	w.Attach(aibarID, &core.Owner{PlayerID: 1, Faction: "Soviet"})
	w.Attach(aibarID, &core.FogVision{Range: 5})

	// AI starting units
	for i := 0; i < 5; i++ {
		uid := w.Spawn()
		w.Attach(uid, &core.Position{X: float64(52 + i), Y: 52})
		w.Attach(uid, &core.Sprite{Width: 24, Height: 24, Visible: true, ScaleX: 1, ScaleY: 1})
		w.Attach(uid, &core.Health{Current: 100, Max: 100})
		w.Attach(uid, &core.Movable{Speed: 3.0, MoveType: core.MoveInfantry})
		w.Attach(uid, &core.Weapon{Name: "AK", Damage: 12, Range: 4.5, Cooldown: 1.0, DamageType: core.DmgKinetic, TargetType: core.TargetAll})
		w.Attach(uid, &core.Armor{ArmorType: core.ArmorNone})
		w.Attach(uid, &core.Selectable{Radius: 0.5})
		w.Attach(uid, &core.Owner{PlayerID: 1, Faction: "Soviet"})
		w.Attach(uid, &core.FogVision{Range: 5})
	}
}

func (g *Game) Update() error {
	g.input.Update()

	if g.input.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.gameState == "playing" {
			g.gameState = "paused"
			g.gameLoop.Pause()
		} else if g.gameState == "paused" {
			g.gameState = "playing"
			g.gameLoop.Play()
		}
	}

	if g.gameState == "paused" || g.gameState == "gameover" {
		return nil
	}

	g.handleCamera()

	// Toggles
	if g.input.IsKeyJustPressed(ebiten.KeyG) {
		g.showGrid = !g.showGrid
	}
	if g.input.IsKeyJustPressed(ebiten.KeyM) {
		g.showMinimap = !g.showMinimap
	}

	// Hover tile
	wx, wy := g.renderer.Camera.ScreenToWorld(g.input.MouseX, g.input.MouseY)
	g.hoverTileX = int(math.Floor(wx))
	g.hoverTileY = int(math.Floor(wy))

	// Control groups
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	for i := 0; i <= 9; i++ {
		key := ebiten.Key0 + ebiten.Key(i)
		if g.input.IsKeyJustPressed(key) {
			if ctrl {
				g.hud.AssignControlGroup(i)
			} else {
				g.hud.RecallControlGroup(i)
			}
		}
	}

	// Handle right click: move or attack-move selected units
	if g.input.RightJustPressed && !g.hud.IsInSidebar(g.input.MouseX, g.input.MouseY) {
		gx, gy := int(math.Floor(wx)), int(math.Floor(wy))
		w := g.gameLoop.World
		for _, id := range g.hud.SelectedIDs {
			if w.Has(id, core.CompMovable) {
				systems.OrderMove(w, g.navGrid, id, gx, gy)
			}
		}
		g.audioMgr.PlaySFX(audio.SndMove, wx, wy)
	}

	// Handle left click: select
	if g.input.LeftJustReleased && !g.input.Dragging {
		if !g.hud.HandleClick(g.input.MouseX, g.input.MouseY) {
			shift := ebiten.IsKeyPressed(ebiten.KeyShift)
			g.handleSelection(wx, wy, shift)
		}
	}

	// Box select
	if g.input.LeftJustReleased && g.input.Dragging {
		g.handleBoxSelect()
	}

	// Queue unit production from barracks (Q key)
	if g.input.IsKeyJustPressed(ebiten.KeyQ) {
		g.queueUnit("gi")
	}

	// Update audio listener
	g.audioMgr.SetCameraPos(g.renderer.Camera.X, g.renderer.Camera.Y)

	// Simulation tick
	g.gameLoop.Update()
	g.eventBus.Dispatch()

	// Check game over
	for _, p := range g.players.Players {
		if p.Defeated && p.ID == 0 {
			g.gameState = "gameover"
		}
	}

	return nil
}

func (g *Game) queueUnit(unitType string) {
	w := g.gameLoop.World
	player := g.players.GetPlayer(0)
	udef, ok := g.techTree.Units[unitType]
	if !ok || player.Credits < udef.Cost {
		return
	}
	// Find a production building
	for _, id := range w.Query(core.CompProduction, core.CompOwner) {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		if own.PlayerID != 0 {
			continue
		}
		prod := w.Get(id, core.CompProduction).(*core.Production)
		if len(prod.Queue) < 5 {
			player.Credits -= udef.Cost
			prod.Queue = append(prod.Queue, unitType)
			return
		}
	}
}

func (g *Game) handleSelection(wx, wy float64, shift bool) {
	w := g.gameLoop.World
	if !shift {
		g.hud.SelectedIDs = nil
	}
	units := w.Query(core.CompPosition, core.CompSelectable, core.CompOwner)
	for _, id := range units {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		if own.PlayerID != 0 {
			continue
		}
		pos := w.Get(id, core.CompPosition).(*core.Position)
		sx, sy := g.renderer.Camera.WorldToScreen(pos.X, pos.Y)
		dx := float64(g.input.MouseX - sx)
		dy := float64(g.input.MouseY - sy)
		if math.Sqrt(dx*dx+dy*dy) < 20 {
			g.hud.SelectedIDs = append(g.hud.SelectedIDs, id)
			g.audioMgr.PlaySFX(audio.SndSelect, pos.X, pos.Y)
			break
		}
	}
}

func (g *Game) handleBoxSelect() {
	x1, y1 := g.input.DragStartX, g.input.DragStartY
	x2, y2 := g.input.MouseX, g.input.MouseY
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	w := g.gameLoop.World
	g.hud.SelectedIDs = nil
	units := w.Query(core.CompPosition, core.CompSelectable, core.CompOwner)
	for _, id := range units {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		if own.PlayerID != 0 {
			continue
		}
		pos := w.Get(id, core.CompPosition).(*core.Position)
		sx, sy := g.renderer.Camera.WorldToScreen(pos.X, pos.Y)
		if sx >= x1 && sx <= x2 && sy >= y1 && sy <= y2 {
			g.hud.SelectedIDs = append(g.hud.SelectedIDs, id)
		}
	}
}

func (g *Game) handleCamera() {
	speed := g.scrollSpeed / 60.0
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
	cam := g.renderer.Camera
	if cam.EdgeScroll {
		edge := cam.EdgeSize
		if g.input.MouseX < edge {
			cam.Pan(-speed, 0)
		}
		if g.input.MouseX > ScreenWidth-edge {
			cam.Pan(speed, 0)
		}
		if g.input.MouseY < edge {
			cam.Pan(0, -speed)
		}
		if g.input.MouseY > ScreenHeight-edge {
			cam.Pan(0, speed)
		}
	}
	if g.input.ScrollY != 0 {
		cam.ZoomAt(g.input.ScrollY*0.1, g.input.MouseX, g.input.MouseY)
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		cam.Pan(float64(-g.input.MouseDX), float64(-g.input.MouseDY))
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 30, 255})

	// Draw map
	g.renderer.DrawMap(screen, g.tileMap)
	if g.showGrid {
		g.renderer.DrawGrid(screen, g.tileMap)
	}

	// Draw fog of war overlay
	g.drawFogOverlay(screen)

	// Draw hover tile
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

	// Draw entities
	g.drawEntities(screen)

	// Draw projectiles
	g.drawProjectiles(screen)

	// Draw selection box
	if x1, y1, x2, y2, active := g.input.DragRect(); active {
		g.renderer.DrawSelectionBox(screen, x1, y1, x2, y2)
	}

	// Minimap
	if g.showMinimap {
		g.renderer.DrawMinimap(screen, g.tileMap, ScreenWidth-g.hud.SidebarWidth-170, ScreenHeight-170, 160)
	}

	// HUD
	g.hud.Draw(screen, g.gameLoop.World)

	// Top-left debug info
	tile := g.tileMap.At(g.hoverTileX, g.hoverTileY)
	terrainName := "OOB"
	if tile != nil {
		terrainName = terrainTypeName(tile.Terrain)
	}
	info := fmt.Sprintf(
		"RTS Engine v0.2.0 | FPS: %.0f | Tick: %d | State: %s\n"+
			"Tile: (%d,%d) %s | Entities: %d | Selected: %d\n"+
			"Zoom: %.1fx | [WASD]Pan [G]Grid [M]Minimap [Q]Train GI [ESC]Pause",
		ebiten.ActualFPS(), g.gameLoop.CurrentTick(), g.gameState,
		g.hoverTileX, g.hoverTileY, terrainName,
		g.gameLoop.World.EntityCount(), len(g.hud.SelectedIDs),
		g.renderer.Camera.Zoom,
	)
	ebitenutil.DebugPrintAt(screen, info, 5, 35)

	// Pause/GameOver overlay
	if g.gameState == "paused" {
		g.drawOverlay(screen, "PAUSED", "Press ESC to resume")
	}
	if g.gameState == "gameover" {
		winner := "Enemy"
		p := g.players.GetPlayer(1)
		if p != nil && p.Defeated {
			winner = "You"
		}
		g.drawOverlay(screen, "GAME OVER", winner+" wins!")
	}
}

func (g *Game) drawOverlay(screen *ebiten.Image, title, subtitle string) {
	overlay := ebiten.NewImage(ScreenWidth, ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 150})
	screen.DrawImage(overlay, nil)
	ebitenutil.DebugPrintAt(screen, title, ScreenWidth/2-30, ScreenHeight/2-20)
	ebitenutil.DebugPrintAt(screen, subtitle, ScreenWidth/2-40, ScreenHeight/2)
}

func (g *Game) drawFogOverlay(screen *ebiten.Image) {
	fog := g.fogSys.Fogs[0]
	if fog == nil {
		return
	}
	minX, minY, maxX, maxY := g.renderer.Camera.VisibleTileRange(g.tileMap.Width, g.tileMap.Height)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			state := fog.At(x, y)
			if state == systems.FogVisible {
				continue
			}
			sx, sy := g.renderer.Camera.WorldToScreen(float64(x), float64(y))
			tw := g.tileMap.TileWidth
			th := g.tileMap.TileHeight
			sx -= tw / 2
			var alpha uint8
			if state == systems.FogShroud {
				alpha = 200
			} else {
				alpha = 100
			}
			fogImg := ebiten.NewImage(tw, th)
			fogImg.Fill(color.RGBA{0, 0, 0, alpha})
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(sx), float64(sy))
			screen.DrawImage(fogImg, op)
		}
	}
}

func (g *Game) drawEntities(screen *ebiten.Image) {
	w := g.gameLoop.World
	fog := g.fogSys.Fogs[0]

	// Draw buildings
	for _, id := range w.Query(core.CompPosition, core.CompBuilding, core.CompOwner) {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		own := w.Get(id, core.CompOwner).(*core.Owner)

		// Fog check
		if fog != nil && !fog.IsVisible(int(pos.X), int(pos.Y)) && own.PlayerID != 0 {
			continue
		}

		sx, sy := g.renderer.Camera.WorldToScreen(pos.X, pos.Y)
		bldg := w.Get(id, core.CompBuilding).(*core.Building)
		bw := float32(bldg.SizeX * 20)
		bh := float32(bldg.SizeY * 12)

		bcolor := color.RGBA{60, 60, 200, 255}
		if own.PlayerID == 1 {
			bcolor = color.RGBA{200, 60, 60, 255}
		}

		vector.DrawFilledRect(screen, float32(sx)-bw/2, float32(sy)-bh/2, bw, bh, bcolor, false)
		vector.StrokeRect(screen, float32(sx)-bw/2, float32(sy)-bh/2, bw, bh, 1, color.RGBA{255, 255, 255, 150}, false)

		// Health bar
		hp := w.Get(id, core.CompHealth)
		if hp != nil {
			h := hp.(*core.Health)
			ratio := float32(h.Ratio())
			vector.DrawFilledRect(screen, float32(sx)-bw/2, float32(sy)-bh/2-5, bw*ratio, 3, color.RGBA{0, 200, 0, 255}, false)
		}

		// Production progress
		if prod := w.Get(id, core.CompProduction); prod != nil {
			p := prod.(*core.Production)
			if len(p.Queue) > 0 {
				vector.DrawFilledRect(screen, float32(sx)-bw/2, float32(sy)+bh/2+2, bw*float32(p.Progress), 3, color.RGBA{255, 255, 0, 255}, false)
			}
		}

		// Check if selected
		for _, sid := range g.hud.SelectedIDs {
			if sid == id {
				vector.StrokeRect(screen, float32(sx)-bw/2-2, float32(sy)-bh/2-2, bw+4, bh+4, 2, color.RGBA{0, 255, 0, 200}, false)
				break
			}
		}
	}

	// Draw units (non-buildings with Position + Selectable)
	for _, id := range w.Query(core.CompPosition, core.CompSelectable, core.CompOwner) {
		if w.Has(id, core.CompBuilding) {
			continue
		}
		pos := w.Get(id, core.CompPosition).(*core.Position)
		own := w.Get(id, core.CompOwner).(*core.Owner)

		// Fog check
		if fog != nil && !fog.IsVisible(int(pos.X), int(pos.Y)) && own.PlayerID != 0 {
			continue
		}

		sx, sy := g.renderer.Camera.WorldToScreen(pos.X, pos.Y)

		selected := false
		for _, sid := range g.hud.SelectedIDs {
			if sid == id {
				selected = true
				break
			}
		}

		if selected {
			vector.DrawFilledCircle(screen, float32(sx), float32(sy), 16, color.RGBA{0, 255, 0, 60}, false)
			vector.StrokeCircle(screen, float32(sx), float32(sy), 16, 2, color.RGBA{0, 255, 0, 200}, false)
		}

		unitColor := color.RGBA{60, 120, 255, 255}
		if own.PlayerID == 1 {
			unitColor = color.RGBA{255, 60, 60, 255}
		}

		// Harvester: slightly bigger
		radius := float32(10)
		if w.Has(id, core.CompHarvester) {
			radius = 13
			unitColor.G = 200
		}

		vector.DrawFilledCircle(screen, float32(sx), float32(sy), radius, unitColor, false)
		vector.StrokeCircle(screen, float32(sx), float32(sy), radius, 1, color.RGBA{255, 255, 255, 180}, false)

		// Health bar
		if hp := w.Get(id, core.CompHealth); hp != nil {
			h := hp.(*core.Health)
			ratio := float32(h.Ratio())
			barW := float32(24)
			barX := float32(sx) - barW/2
			barY := float32(sy) - radius - 6
			vector.DrawFilledRect(screen, barX, barY, barW, 3, color.RGBA{40, 40, 40, 200}, false)
			barColor := color.RGBA{0, 200, 0, 255}
			if ratio < 0.5 {
				barColor = color.RGBA{255, 200, 0, 255}
			}
			if ratio < 0.25 {
				barColor = color.RGBA{255, 0, 0, 255}
			}
			vector.DrawFilledRect(screen, barX, barY, barW*ratio, 3, barColor, false)
		}

		// Harvester load indicator
		if harv := w.Get(id, core.CompHarvester); harv != nil {
			h := harv.(*core.Harvester)
			if h.Current > 0 {
				loadRatio := float32(h.Current) / float32(h.Capacity)
				vector.DrawFilledRect(screen, float32(sx)-12, float32(sy)+radius+2, 24*loadRatio, 2, color.RGBA{255, 215, 0, 255}, false)
			}
		}
	}
}

func (g *Game) drawProjectiles(screen *ebiten.Image) {
	w := g.gameLoop.World
	for _, id := range w.Query(core.CompPosition, core.CompProjectile) {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		sx, sy := g.renderer.Camera.WorldToScreen(pos.X, pos.Y)
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), 3, color.RGBA{255, 255, 100, 255}, false)
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func generateDemoMap() *maplib.TileMap {
	tm := maplib.NewTileMap("Demo Battlefield", MapSize, MapSize)
	tm.SetTerrain(0, 0, MapSize-1, MapSize-1, maplib.TerrainGrass)

	// River
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

	// Forests
	for _, f := range [][4]int{{5, 5, 12, 10}, {45, 8, 55, 15}, {20, 45, 30, 52}} {
		tm.SetTerrain(f[0], f[1], f[2], f[3], maplib.TerrainForest)
	}

	// Ore fields
	for _, pos := range [][2]int{{15, 15}, {16, 15}, {15, 16}, {16, 16}, {17, 15}, {45, 45}, {46, 45}, {45, 46}, {46, 46}, {47, 45}} {
		tm.PlaceOre(pos[0], pos[1], 1000)
	}

	// Terrain features
	tm.SetTerrain(30, 10, 35, 12, maplib.TerrainCliff)
	tm.SetTerrain(25, 50, 28, 55, maplib.TerrainRock)
	for x := 0; x < MapSize; x++ {
		tm.SetTerrain(x, MapSize/4, x, MapSize/4, maplib.TerrainRoad)
	}
	for y := 0; y < MapSize; y++ {
		tm.SetTerrain(MapSize/4, y, MapSize/4, y, maplib.TerrainRoad)
	}
	tm.SetTerrain(50, 50, 60, 60, maplib.TerrainSand)

	tm.StartPositions = []maplib.StartPos{
		{PlayerSlot: 0, X: 10, Y: 10},
		{PlayerSlot: 1, X: 54, Y: 54},
	}
	return tm
}

func terrainTypeName(t maplib.TerrainType) string {
	names := map[maplib.TerrainType]string{
		maplib.TerrainGrass: "Grass", maplib.TerrainDirt: "Dirt", maplib.TerrainSand: "Sand",
		maplib.TerrainWater: "Water", maplib.TerrainDeepWater: "Deep Water", maplib.TerrainRock: "Rock",
		maplib.TerrainCliff: "Cliff", maplib.TerrainRoad: "Road", maplib.TerrainBridge: "Bridge",
		maplib.TerrainOre: "Ore Field", maplib.TerrainGem: "Gem Field", maplib.TerrainSnow: "Snow",
		maplib.TerrainUrban: "Urban", maplib.TerrainForest: "Forest",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return "Unknown"
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("⚔️ RTS Engine v0.2.0 — Full Game")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

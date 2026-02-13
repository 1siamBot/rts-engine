package main

import (
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/1siamBot/rts-engine/engine/ai"
	"github.com/1siamBot/rts-engine/engine/audio"
	"github.com/1siamBot/rts-engine/engine/core"
	"github.com/1siamBot/rts-engine/engine/input"
	"github.com/1siamBot/rts-engine/engine/maplib"
	"github.com/1siamBot/rts-engine/engine/pathfind"
	"github.com/1siamBot/rts-engine/engine/render3d"
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

var (
	screenshotTarget string
	screenshotFrame  int
	frameCount       int
)

// Game implements ebiten.Game
type Game struct {
	renderer *render3d.Renderer3D
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
	gameState   string

	// Settings
	scrollSpeed float64

	// Cached images
	fogWhiteImg   *ebiten.Image
	selectionFill *ebiten.Image
}

func NewGame() *Game {
	g := &Game{
		renderer:    render3d.NewRenderer3D(ScreenWidth, ScreenHeight),
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

	g.navGrid = pathfind.NewNavGrid(g.tileMap)

	g.hud = ui.NewHUD(ScreenWidth, ScreenHeight, g.techTree, g.players, 0)

	// Wire up 3D sprite rendering callbacks (return false to use HUD default fallback)
	g.hud.UnitDrawFn = func(screen *ebiten.Image, w *core.World, id core.EntityID, sx, sy int, playerID int) bool {
		return false // Units are drawn by 3D renderer now
	}
	g.hud.BuildingDrawFn = func(screen *ebiten.Image, w *core.World, id core.EntityID, sx, sy int) bool {
		return false // Buildings are drawn by 3D renderer now
	}

	g.fogSys = systems.NewFogSystem(g.tileMap.Width, g.tileMap.Height, g.players)

	// Register systems
	w := g.gameLoop.World
	w.AddSystem(&systems.PowerSystem{Players: g.players})
	w.AddSystem(&systems.BuildingConstructionSystem{Players: g.players, EventBus: g.eventBus})
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

	g.renderer.Camera.SetMapSize(MapSize, MapSize)
	g.renderer.Camera.CenterOn(12, 12)

	g.spawnInitialEntities()

	// Mark initial building tiles as occupied
	g.markInitialBuildingTiles()

	g.gameLoop.Play()
	return g
}

func (g *Game) spawnInitialEntities() {
	w := g.gameLoop.World

	// Player 0: Construction Yard
	cyID := w.Spawn()
	w.Attach(cyID, &core.Position{X: 10, Y: 10})
	w.Attach(cyID, &core.Health{Current: 1000, Max: 1000})
	w.Attach(cyID, &core.Building{SizeX: 3, SizeY: 3, PowerGen: 0, IsConYard: true, Sellable: true})
	w.Attach(cyID, &core.Production{Rate: 1.0, Rally: core.TilePos{X: 13, Y: 13}})
	w.Attach(cyID, &core.Owner{PlayerID: 0, Faction: "Allied"})
	w.Attach(cyID, &core.FogVision{Range: 8})
	w.Attach(cyID, &core.Selectable{Radius: 1.5})
	w.Attach(cyID, &core.BuildingName{Key: "construction_yard"})

	// Player 0: Power Plant
	ppID := w.Spawn()
	w.Attach(ppID, &core.Position{X: 14, Y: 10})
	w.Attach(ppID, &core.Health{Current: 750, Max: 750})
	w.Attach(ppID, &core.Building{SizeX: 2, SizeY: 2, PowerGen: 100, Sellable: true})
	w.Attach(ppID, &core.Owner{PlayerID: 0, Faction: "Allied"})
	w.Attach(ppID, &core.FogVision{Range: 5})
	w.Attach(ppID, &core.Selectable{Radius: 1.0})
	w.Attach(ppID, &core.BuildingName{Key: "power_plant"})

	// Player 0: Barracks
	barID := w.Spawn()
	w.Attach(barID, &core.Position{X: 10, Y: 14})
	w.Attach(barID, &core.Health{Current: 500, Max: 500})
	w.Attach(barID, &core.Building{SizeX: 2, SizeY: 2, PowerDraw: 20, Sellable: true})
	w.Attach(barID, &core.Production{Rate: 1.0, Rally: core.TilePos{X: 12, Y: 16}})
	w.Attach(barID, &core.Owner{PlayerID: 0, Faction: "Allied"})
	w.Attach(barID, &core.FogVision{Range: 5})
	w.Attach(barID, &core.Selectable{Radius: 1.0})
	w.Attach(barID, &core.BuildingName{Key: "barracks"})

	// Player 0: Refinery
	refID := w.Spawn()
	w.Attach(refID, &core.Position{X: 14, Y: 14})
	w.Attach(refID, &core.Health{Current: 900, Max: 900})
	w.Attach(refID, &core.Building{SizeX: 3, SizeY: 3, PowerDraw: 30, Sellable: true})
	w.Attach(refID, &core.Owner{PlayerID: 0, Faction: "Allied"})
	w.Attach(refID, &core.FogVision{Range: 5})
	w.Attach(refID, &core.Selectable{Radius: 1.5})
	w.Attach(refID, &core.BuildingName{Key: "refinery"})

	// Player 0: Starting infantry
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
	w.Attach(aicyID, &core.Building{SizeX: 3, SizeY: 3, PowerGen: 0, IsConYard: true})
	w.Attach(aicyID, &core.Production{Rate: 1.0, Rally: core.TilePos{X: 52, Y: 52}})
	w.Attach(aicyID, &core.Owner{PlayerID: 1, Faction: "Soviet"})
	w.Attach(aicyID, &core.FogVision{Range: 8})
	w.Attach(aicyID, &core.BuildingName{Key: "construction_yard"})

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

func (g *Game) markInitialBuildingTiles() {
	w := g.gameLoop.World
	for _, id := range w.Query(core.CompBuilding, core.CompPosition) {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		bldg := w.Get(id, core.CompBuilding).(*core.Building)
		systems.OccupyTiles(g.tileMap, int(pos.X), int(pos.Y), bldg.SizeX, bldg.SizeY)
	}
}

func (g *Game) Update() error {
	g.input.Update()
	g.hud.Update(1.0 / 60.0)
	g.renderer.Update(1.0 / 60.0)
	g.renderer.Camera.SmoothUpdate(1.0 / 60.0)

	if g.input.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.hud.Placement.Active {
			g.cancelPlacementWithRefund()
		} else if g.gameState == "playing" {
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

	// Update placement ghost position
	if g.hud.Placement.Active {
		g.hud.Placement.TileX = g.hoverTileX
		g.hud.Placement.TileY = g.hoverTileY
		g.hud.Placement.Valid = g.canPlaceBuilding(g.hoverTileX, g.hoverTileY, g.hud.Placement.SizeX, g.hud.Placement.SizeY)
	}

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

	if g.input.IsKeyJustPressed(ebiten.KeyH) {
		g.tryDeployMCV()
	}
	if g.input.IsKeyJustPressed(ebiten.KeyDelete) {
		g.trySellBuilding()
	}

	// Handle right click
	if g.input.RightJustPressed {
		if g.hud.Placement.Active {
			g.cancelPlacementWithRefund()
		} else if !g.hud.IsInSidebar(g.input.MouseX, g.input.MouseY) {
			gx, gy := int(math.Floor(wx)), int(math.Floor(wy))
			w := g.gameLoop.World
			for _, id := range g.hud.SelectedIDs {
				if w.Has(id, core.CompMovable) {
					systems.OrderMove(w, g.navGrid, id, gx, gy)
				}
			}
			g.audioMgr.PlaySFX(audio.SndMove, wx, wy)
		}
	}

	// Handle left click
	if g.input.LeftJustReleased && !g.input.Dragging {
		if g.hud.Placement.Active && g.hud.Placement.Valid {
			g.placeBuilding()
		} else if g.hud.IsInMinimap(g.input.MouseX, g.input.MouseY) {
			wmx, wmy := g.hud.GetMinimapWorldPos(g.input.MouseX, g.input.MouseY, MapSize)
			g.renderer.Camera.CenterOn(wmx, wmy)
		} else if !g.hud.HandleClick(g.input.MouseX, g.input.MouseY) {
			if bKey := g.hud.GetSidebarBuildingClick(g.input.MouseX, g.input.MouseY, g.gameLoop.World); bKey != "" {
				g.startBuildingPurchase(bKey)
			} else if uKey := g.hud.GetSidebarUnitClick(g.input.MouseX, g.input.MouseY); uKey != "" {
				g.queueUnit(uKey)
			} else {
				shift := ebiten.IsKeyPressed(ebiten.KeyShift)
				g.handleSelection(wx, wy, shift)
			}
		}
	}

	if g.input.LeftJustReleased && g.input.Dragging && !g.hud.Placement.Active {
		g.handleBoxSelect()
	}

	if g.input.IsKeyJustPressed(ebiten.KeyQ) {
		g.queueUnit("gi")
	}

	g.audioMgr.SetCameraPos(g.renderer.Camera.TargetX, g.renderer.Camera.TargetY)

	g.gameLoop.Update()
	g.eventBus.Dispatch()

	for _, p := range g.players.Players {
		if p.Defeated && p.ID == 0 {
			g.gameState = "gameover"
		}
	}

	return nil
}

func (g *Game) startBuildingPurchase(key string) {
	bdef, ok := g.techTree.Buildings[key]
	if !ok {
		return
	}
	player := g.players.GetPlayer(0)
	if player == nil {
		return
	}

	// Check con yard exists
	if !g.hud.PlayerHasConYard(g.gameLoop.World) {
		g.hud.ShowMessage("Need Construction Yard", 2.0)
		return
	}

	// Check prerequisites
	if !g.techTree.HasPrereqs(g.gameLoop.World, 0, bdef.Prereqs) {
		g.hud.ShowMessage("Missing prerequisites", 2.0)
		return
	}

	// Check credits
	if player.Credits < bdef.Cost {
		g.hud.ShowMessage("Insufficient Funds", 2.0)
		return
	}

	player.Credits -= bdef.Cost
	g.hud.StartPlacement(key)
}

func (g *Game) placeBuilding() {
	key := g.hud.Placement.BuildingKey
	tx, ty := g.hud.Placement.TileX, g.hud.Placement.TileY
	player := g.players.GetPlayer(0)
	faction := "Allied"
	if player != nil {
		faction = player.Faction
	}

	systems.PlaceBuilding(g.gameLoop.World, key, g.techTree, 0, tx, ty, faction, g.eventBus)

	// Mark tiles occupied
	if bdef, ok := g.techTree.Buildings[key]; ok {
		systems.OccupyTiles(g.tileMap, tx, ty, bdef.SizeX, bdef.SizeY)
	}

	g.hud.CancelPlacement()
	g.audioMgr.PlaySFX(audio.SndBuild, float64(tx), float64(ty))
}

func (g *Game) cancelPlacementWithRefund() {
	if !g.hud.Placement.Active {
		return
	}
	key := g.hud.Placement.BuildingKey
	if bdef, ok := g.techTree.Buildings[key]; ok {
		player := g.players.GetPlayer(0)
		if player != nil {
			player.Credits += bdef.Cost
		}
	}
	g.hud.CancelPlacement()
}

func (g *Game) canPlaceBuilding(tileX, tileY, sizeX, sizeY int) bool {
	for dy := 0; dy < sizeY; dy++ {
		for dx := 0; dx < sizeX; dx++ {
			tx, ty := tileX+dx, tileY+dy
			if !g.tileMap.InBounds(tx, ty) {
				return false
			}
			tile := g.tileMap.At(tx, ty)
			if tile == nil {
				return false
			}
			// Can't build on water, deep water, cliffs
			if tile.Terrain == maplib.TerrainWater || tile.Terrain == maplib.TerrainDeepWater || tile.Terrain == maplib.TerrainCliff {
				return false
			}
			// Can't overlap existing buildings
			if tile.Occupied {
				return false
			}
		}
	}
	// Must be near an existing owned building (build radius ~10 tiles)
	w := g.gameLoop.World
	nearBuilding := false
	for _, bid := range w.Query(core.CompBuilding, core.CompOwner, core.CompPosition) {
		o := w.Get(bid, core.CompOwner).(*core.Owner)
		if o.PlayerID != 0 {
			continue
		}
		bp := w.Get(bid, core.CompPosition).(*core.Position)
		dx := float64(tileX) - bp.X
		dy := float64(tileY) - bp.Y
		if dx*dx+dy*dy < 100 {
			nearBuilding = true
			break
		}
	}
	return nearBuilding
}

func (g *Game) tryDeployMCV() {
	w := g.gameLoop.World
	for _, id := range g.hud.SelectedIDs {
		if w.Has(id, core.CompMCV) {
			systems.DeployMCV(w, id, g.eventBus)
			g.hud.SelectedIDs = nil
			return
		}
		if bldg := w.Get(id, core.CompBuilding); bldg != nil {
			b := bldg.(*core.Building)
			if b.IsConYard {
				systems.UndeployConYard(w, id, g.eventBus)
				g.hud.SelectedIDs = nil
				return
			}
		}
	}
}

func (g *Game) trySellBuilding() {
	w := g.gameLoop.World
	for _, id := range g.hud.SelectedIDs {
		if bldg := w.Get(id, core.CompBuilding); bldg != nil {
			b := bldg.(*core.Building)
			if b.Sellable {
				pos := w.Get(id, core.CompPosition).(*core.Position)
				g.hud.AddEffect(pos.X, pos.Y, "explosion", 15)
				g.renderer.Particles.AddExplosion(pos.X, pos.Y)
				// Free occupied tiles
				systems.FreeTiles(g.tileMap, int(pos.X), int(pos.Y), b.SizeX, b.SizeY)
				systems.SellBuilding(w, id, g.techTree, g.players)
			}
		}
	}
	g.hud.SelectedIDs = nil
}

func (g *Game) queueUnit(unitType string) {
	w := g.gameLoop.World
	player := g.players.GetPlayer(0)
	udef, ok := g.techTree.Units[unitType]
	if !ok {
		return
	}

	// Check credits
	if player.Credits < udef.Cost {
		g.hud.ShowMessage("Insufficient Funds", 2.0)
		return
	}

	// Check prereqs
	if !g.techTree.HasPrereqs(w, 0, udef.Prereqs) {
		g.hud.ShowMessage("Missing prerequisites", 2.0)
		return
	}

	// Find a production building that can produce this unit
	bid := systems.FindProductionBuilding(w, g.techTree, 0, unitType)
	if bid == 0 {
		g.hud.ShowMessage("No building can produce this unit", 2.0)
		return
	}

	prod := w.Get(bid, core.CompProduction).(*core.Production)
	player.Credits -= udef.Cost
	prod.Queue = append(prod.Queue, unitType)
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
		cam.ZoomAt(g.input.ScrollY, g.input.MouseX, g.input.MouseY)
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		cam.Pan(float64(-g.input.MouseDX), float64(-g.input.MouseDY))
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{12, 12, 20, 255})

	// Draw 3D scene (terrain + buildings + units + projectiles + particles)
	g.renderer.DrawScene(screen, g.tileMap, g.gameLoop.World, 0)

	if g.showGrid {
		g.renderer.DrawGrid(screen, g.tileMap)
	}

	// Fog of war overlay
	g.drawFogOverlay(screen)

	// Hover tile highlight in 3D
	if g.tileMap.InBounds(g.hoverTileX, g.hoverTileY) {
		g.drawHoverTile(screen)
	}

	// Health bars as 2D overlays at 3D projected positions
	g.drawHealthBars(screen)

	// Placement ghost in 3D
	if g.hud.Placement.Active {
		g.drawPlacementGhost(screen)
	}

	// Selection box
	if x1, y1, x2, y2, active := g.input.DragRect(); active && !g.hud.Placement.Active {
		g.renderer.DrawSelectionBox(screen, x1, y1, x2, y2)
	}

	// HUD panels (2D overlay)
	g.hud.Draw(screen, g.gameLoop.World)

	// Placement mode indicator
	if g.hud.Placement.Active {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Placing: %s (Click to place, ESC/Right-click to cancel)", g.hud.Placement.BuildingKey), 10, ScreenHeight-20)
	}

	// Pause/GameOver overlay
	if g.gameState == "paused" {
		g.drawOverlay(screen, "⏸ PAUSED", "Press ESC to resume")
	}
	if g.gameState == "gameover" {
		winner := "Enemy"
		p := g.players.GetPlayer(1)
		if p != nil && p.Defeated {
			winner = "You"
		}
		g.drawOverlay(screen, "GAME OVER", winner+" wins!")
	}

	// Screenshot capture
	frameCount++
	if screenshotTarget != "" && frameCount >= screenshotFrame {
		f, err := os.Create(screenshotTarget)
		if err != nil {
			log.Fatalf("Screenshot: %v", err)
		}
		if err := png.Encode(f, screen); err != nil {
			f.Close()
			log.Fatalf("Screenshot encode: %v", err)
		}
		f.Close()
		log.Printf("Screenshot saved to %s (%dx%d)", screenshotTarget, ScreenWidth, ScreenHeight)
		os.Exit(0)
	}
}

func (g *Game) drawHoverTile(screen *ebiten.Image) {
	x, y := g.hoverTileX, g.hoverTileY
	hoverColor := color.RGBA{255, 255, 0, 80}

	sx0, sy0, _ := g.renderer.Camera.Project3DToScreen(float64(x), 0.02, float64(y))
	sx1, sy1, _ := g.renderer.Camera.Project3DToScreen(float64(x+1), 0.02, float64(y))
	sx2, sy2, _ := g.renderer.Camera.Project3DToScreen(float64(x+1), 0.02, float64(y+1))
	sx3, sy3, _ := g.renderer.Camera.Project3DToScreen(float64(x), 0.02, float64(y+1))

	vector.StrokeLine(screen, float32(sx0), float32(sy0), float32(sx1), float32(sy1), 2, hoverColor, false)
	vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 2, hoverColor, false)
	vector.StrokeLine(screen, float32(sx2), float32(sy2), float32(sx3), float32(sy3), 2, hoverColor, false)
	vector.StrokeLine(screen, float32(sx3), float32(sy3), float32(sx0), float32(sy0), 2, hoverColor, false)
}

func (g *Game) drawHealthBars(screen *ebiten.Image) {
	w := g.gameLoop.World
	for _, id := range w.Query(core.CompPosition, core.CompHealth, core.CompOwner) {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		hp := w.Get(id, core.CompHealth).(*core.Health)
		if hp.Ratio() >= 1.0 {
			continue // Don't show full health bars
		}

		// Project to screen with slight Y offset above the entity
		heightOffset := 0.5
		if w.Has(id, core.CompBuilding) {
			heightOffset = 1.5
		}
		sx, sy, _ := g.renderer.Camera.Project3DToScreen(pos.X, heightOffset, pos.Y)

		barWidth := 30
		if w.Has(id, core.CompBuilding) {
			barWidth = 50
		}
		g.renderer.DrawHealthBar(screen, sx, sy, hp.Ratio(), barWidth)
	}
}

func (g *Game) drawPlacementGhost(screen *ebiten.Image) {
	tx, ty := g.hud.Placement.TileX, g.hud.Placement.TileY
	sx, sy := g.hud.Placement.SizeX, g.hud.Placement.SizeY

	var outlineColor color.RGBA
	if g.hud.Placement.Valid {
		outlineColor = color.RGBA{0, 255, 0, 150}
	} else {
		outlineColor = color.RGBA{255, 0, 0, 150}
	}

	// Draw outline of placement area
	for dx := 0; dx < sx; dx++ {
		for dy := 0; dy < sy; dy++ {
			fx, fy := float64(tx+dx), float64(ty+dy)
			s0x, s0y, _ := g.renderer.Camera.Project3DToScreen(fx, 0.03, fy)
			s1x, s1y, _ := g.renderer.Camera.Project3DToScreen(fx+1, 0.03, fy)
			s2x, s2y, _ := g.renderer.Camera.Project3DToScreen(fx+1, 0.03, fy+1)
			s3x, s3y, _ := g.renderer.Camera.Project3DToScreen(fx, 0.03, fy+1)

			vector.StrokeLine(screen, float32(s0x), float32(s0y), float32(s1x), float32(s1y), 2, outlineColor, false)
			vector.StrokeLine(screen, float32(s1x), float32(s1y), float32(s2x), float32(s2y), 2, outlineColor, false)
			vector.StrokeLine(screen, float32(s2x), float32(s2y), float32(s3x), float32(s3y), 2, outlineColor, false)
			vector.StrokeLine(screen, float32(s3x), float32(s3y), float32(s0x), float32(s0y), 2, outlineColor, false)
		}
	}
}

func (g *Game) drawFogOverlay(screen *ebiten.Image) {
	fog := g.fogSys.Fogs[0]
	if fog == nil {
		return
	}

	// Reuse cached white image
	if g.fogWhiteImg == nil {
		g.fogWhiteImg = ebiten.NewImage(4, 4)
		g.fogWhiteImg.Fill(color.White)
	}

	minX, minY, maxX, maxY := g.renderer.Camera.VisibleTileRange(g.tileMap.Width, g.tileMap.Height)

	// Batch fog triangles
	var vertices []ebiten.Vertex
	var indices []uint16

	shroudR := float32(5) / 255
	shroudG := float32(5) / 255
	shroudB := float32(15) / 255

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			state := fog.At(x, y)
			if state == systems.FogVisible {
				continue
			}
			var alphaF float32
			if state == systems.FogShroud {
				alphaF = float32(200) / 255
			} else {
				alphaF = float32(80) / 255
			}

			fx, fy := float64(x), float64(y)
			s0x, s0y, _ := g.renderer.Camera.Project3DToScreen(fx, 0.05, fy)
			s1x, s1y, _ := g.renderer.Camera.Project3DToScreen(fx+1, 0.05, fy)
			s2x, s2y, _ := g.renderer.Camera.Project3DToScreen(fx+1, 0.05, fy+1)
			s3x, s3y, _ := g.renderer.Camera.Project3DToScreen(fx, 0.05, fy+1)

			base := uint16(len(vertices))
			vertices = append(vertices,
				ebiten.Vertex{DstX: float32(s0x), DstY: float32(s0y), SrcX: 1, SrcY: 1, ColorR: shroudR, ColorG: shroudG, ColorB: shroudB, ColorA: alphaF},
				ebiten.Vertex{DstX: float32(s1x), DstY: float32(s1y), SrcX: 1, SrcY: 1, ColorR: shroudR, ColorG: shroudG, ColorB: shroudB, ColorA: alphaF},
				ebiten.Vertex{DstX: float32(s2x), DstY: float32(s2y), SrcX: 1, SrcY: 1, ColorR: shroudR, ColorG: shroudG, ColorB: shroudB, ColorA: alphaF},
				ebiten.Vertex{DstX: float32(s3x), DstY: float32(s3y), SrcX: 1, SrcY: 1, ColorR: shroudR, ColorG: shroudG, ColorB: shroudB, ColorA: alphaF},
			)
			indices = append(indices, base, base+1, base+2, base, base+2, base+3)

			if len(vertices) >= 65000 {
				op := &ebiten.DrawTrianglesOptions{}
				op.Blend = ebiten.BlendSourceOver
				screen.DrawTriangles(vertices, indices, g.fogWhiteImg, op)
				vertices = vertices[:0]
				indices = indices[:0]
			}
		}
	}

	if len(vertices) > 0 {
		op := &ebiten.DrawTrianglesOptions{}
		op.Blend = ebiten.BlendSourceOver
		screen.DrawTriangles(vertices, indices, g.fogWhiteImg, op)
	}
}

func (g *Game) drawOverlay(screen *ebiten.Image, title, subtitle string) {
	vector.DrawFilledRect(screen, 0, 0, float32(ScreenWidth), float32(ScreenHeight), color.RGBA{0, 0, 0, 150}, false)

	boxW := float32(300)
	boxH := float32(100)
	boxX := float32(ScreenWidth)/2 - boxW/2
	boxY := float32(ScreenHeight)/2 - boxH/2
	vector.DrawFilledRect(screen, boxX, boxY, boxW, boxH, color.RGBA{15, 15, 30, 240}, false)
	vector.StrokeRect(screen, boxX, boxY, boxW, boxH, 2, color.RGBA{0, 180, 220, 255}, false)

	ebitenutil.DebugPrintAt(screen, title, int(boxX)+boxW_center(title, boxW), int(boxY)+25)
	ebitenutil.DebugPrintAt(screen, subtitle, int(boxX)+boxW_center(subtitle, boxW), int(boxY)+50)
}

func boxW_center(text string, boxW float32) int {
	textW := len(text) * 6
	return int(boxW/2) - textW/2
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

func main() {
	headless := flag.Bool("headless", false, "Run in headless mode (no window)")
	screenshot := flag.String("screenshot", "", "Render one frame to PNG file and exit")
	flag.Parse()

	if os.Getenv("EBITENGINE_GRAPHICS_LIBRARY") == "" {
		os.Setenv("EBITENGINE_GRAPHICS_LIBRARY", "opengl")
	}

	if *screenshot != "" || *headless {
		screenshotPath := *screenshot
		if screenshotPath == "" {
			screenshotPath = "screenshot.png"
		}
		screenshotTarget = screenshotPath
		screenshotFrame = 30
	}

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("⚔️ RTS Engine v0.4.0 — Real 3D Isometric")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

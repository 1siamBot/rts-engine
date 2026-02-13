package ui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/1siamBot/rts-engine/engine/core"
	"github.com/1siamBot/rts-engine/engine/systems"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// CommandType represents a player command
type CommandType int

const (
	CmdNone CommandType = iota
	CmdMove
	CmdAttack
	CmdStop
	CmdGuard
	CmdBuild
	CmdDeploy
	CmdSell
	CmdRally
)

// BuildTab represents which tab is active
type BuildTab int

const (
	TabBuildings BuildTab = iota
	TabUnits
	TabDefense
)

// PlacementMode tracks building placement state
type PlacementMode struct {
	Active      bool
	BuildingKey string
	SizeX, SizeY int
	Valid       bool
	TileX, TileY int
}

// Effect represents a visual effect (explosion, smoke, etc.)
type Effect struct {
	X, Y    float64
	Timer   float64
	MaxTime float64
	Kind    string // "explosion", "smoke", "sparkle"
	Size    float64
}

// HUD is the main heads-up display
type HUD struct {
	ScreenW, ScreenH int
	SidebarWidth     int
	TopBarHeight     int
	BottomPanelH     int
	MinimapSize      int

	// State
	CurrentCommand CommandType
	BuildQueue     []string
	SelectedIDs    []core.EntityID
	ControlGroups  [10][]core.EntityID
	ActiveTab      BuildTab
	Placement      PlacementMode
	Effects        []Effect

	// Animated credits display
	DisplayCredits float64
	ActualCredits  int

	// Hover state
	HoverBuildIdx  int
	HoverCmdIdx    int
	HoverSidebar   bool

	// Build progress tracking for sidebar
	BuildProgress    map[string]float64 // building key -> progress 0-1
	BuildingQueue    []string           // queued building keys

	// References
	TechTree    *systems.TechTree
	Players     *core.PlayerManager
	LocalPlayer int

	// Cached images for rounded rects
	panelCache map[string]*ebiten.Image

	// UI Sprites (metallic panels, buttons, icons)
	Sprites *UISprites

	// Tick counter for animations
	tick float64

	// Sprite draw callbacks (set externally to use real sprites)
	UnitDrawFn     func(screen *ebiten.Image, w *core.World, id core.EntityID, sx, sy int, playerID int) bool
	BuildingDrawFn func(screen *ebiten.Image, w *core.World, id core.EntityID, sx, sy int) bool
}

func NewHUD(sw, sh int, tt *systems.TechTree, pm *core.PlayerManager, localPlayer int) *HUD {
	return &HUD{
		ScreenW:        sw,
		ScreenH:        sh,
		SidebarWidth:   200,
		TopBarHeight:   40,
		BottomPanelH:   100,
		MinimapSize:    160,
		TechTree:       tt,
		Players:        pm,
		LocalPlayer:    localPlayer,
		HoverBuildIdx:  -1,
		HoverCmdIdx:    -1,
		BuildProgress:  make(map[string]float64),
		panelCache:     make(map[string]*ebiten.Image),
		Sprites:        NewUISprites(),
	}
}

// ---- Drawing Helpers ----

func drawRoundedRect(screen *ebiten.Image, x, y, w, h float32, r float32, clr color.RGBA) {
	// Draw main body
	vector.DrawFilledRect(screen, x+r, y, w-2*r, h, clr, false)
	vector.DrawFilledRect(screen, x, y+r, w, h-2*r, clr, false)
	// Corners
	vector.DrawFilledCircle(screen, x+r, y+r, r, clr, false)
	vector.DrawFilledCircle(screen, x+w-r, y+r, r, clr, false)
	vector.DrawFilledCircle(screen, x+r, y+h-r, r, clr, false)
	vector.DrawFilledCircle(screen, x+w-r, y+h-r, r, clr, false)
}

func drawRoundedRectStroke(screen *ebiten.Image, x, y, w, h float32, r float32, clr color.RGBA) {
	// Top and bottom edges
	vector.StrokeLine(screen, x+r, y, x+w-r, y, 1, clr, false)
	vector.StrokeLine(screen, x+r, y+h, x+w-r, y+h, 1, clr, false)
	// Left and right edges
	vector.StrokeLine(screen, x, y+r, x, y+h-r, 1, clr, false)
	vector.StrokeLine(screen, x+w, y+r, x+w, y+h-r, 1, clr, false)
}

func lerpColor(a, b color.RGBA, t float32) color.RGBA {
	return color.RGBA{
		R: uint8(float32(a.R)*(1-t) + float32(b.R)*t),
		G: uint8(float32(a.G)*(1-t) + float32(b.G)*t),
		B: uint8(float32(a.B)*(1-t) + float32(b.B)*t),
		A: uint8(float32(a.A)*(1-t) + float32(b.A)*t),
	}
}

func healthBarColor(ratio float32) color.RGBA {
	if ratio > 0.5 {
		t := (ratio - 0.5) * 2
		return lerpColor(color.RGBA{255, 200, 0, 255}, color.RGBA{0, 220, 0, 255}, t)
	}
	t := ratio * 2
	return lerpColor(color.RGBA{220, 0, 0, 255}, color.RGBA{255, 200, 0, 255}, t)
}

// ---- Color Palette ----
var (
	panelBG      = color.RGBA{15, 15, 30, 210}
	panelBorder  = color.RGBA{40, 80, 140, 255}
	accentCyan   = color.RGBA{0, 180, 220, 255}
	accentBlue   = color.RGBA{40, 100, 200, 255}
	textWhite    = color.RGBA{220, 230, 255, 255}
	textDim      = color.RGBA{120, 140, 170, 255}
	btnNormal    = color.RGBA{30, 40, 65, 240}
	btnHover     = color.RGBA{45, 65, 100, 255}
	btnActive    = color.RGBA{0, 120, 180, 255}
	btnDisabled  = color.RGBA{25, 25, 35, 200}
	healthGreen  = color.RGBA{0, 220, 60, 255}
	healthYellow = color.RGBA{255, 200, 0, 255}
	healthRed    = color.RGBA{220, 30, 30, 255}
	powerGreen   = color.RGBA{60, 200, 60, 255}
	powerRed     = color.RGBA{220, 40, 40, 255}
	selectGreen  = color.RGBA{0, 255, 100, 180}
	minimapBG    = color.RGBA{10, 10, 20, 230}
)

// ---- Update (call each frame for animations) ----

func (h *HUD) Update(dt float64) {
	h.tick += dt

	// Animate credits counter
	player := h.Players.GetPlayer(h.LocalPlayer)
	if player != nil {
		h.ActualCredits = player.Credits
		diff := float64(h.ActualCredits) - h.DisplayCredits
		if math.Abs(diff) < 1 {
			h.DisplayCredits = float64(h.ActualCredits)
		} else {
			h.DisplayCredits += diff * dt * 5 // smooth lerp
		}
	}

	// Update effects
	alive := h.Effects[:0]
	for i := range h.Effects {
		h.Effects[i].Timer += dt
		if h.Effects[i].Timer < h.Effects[i].MaxTime {
			alive = append(alive, h.Effects[i])
		}
	}
	h.Effects = alive
}

func (h *HUD) AddEffect(x, y float64, kind string, size float64) {
	h.Effects = append(h.Effects, Effect{
		X: x, Y: y, Kind: kind, Size: size,
		MaxTime: 1.0,
	})
}

// ---- Draw ----

func (h *HUD) Draw(screen *ebiten.Image, w *core.World) {
	h.drawTopBar(screen)
	h.drawSidebar(screen, w)
	h.drawBottomPanel(screen, w)
	h.drawMinimap(screen, w)
}

// DrawWorldEffects draws selection circles, health bars above units, and effects
func (h *HUD) DrawWorldEffects(screen *ebiten.Image, w *core.World, worldToScreen func(float64, float64) (int, int)) {
	// Draw selection circles and health bars for all visible units
	for _, id := range w.Query(core.CompPosition, core.CompOwner) {
		if w.Has(id, core.CompBuilding) {
			continue
		}
		pos := w.Get(id, core.CompPosition).(*core.Position)
		own := w.Get(id, core.CompOwner).(*core.Owner)
		sx, sy := worldToScreen(pos.X, pos.Y)

		selected := false
		for _, sid := range h.SelectedIDs {
			if sid == id {
				selected = true
				break
			}
		}

		// Selection circle (green ellipse under unit)
		if selected {
			// Draw selection ellipse
			for angle := 0.0; angle < math.Pi*2; angle += 0.1 {
				x1 := float32(sx) + float32(math.Cos(angle)*18)
				y1 := float32(sy) + float32(math.Sin(angle)*9) + 4
				x2 := float32(sx) + float32(math.Cos(angle+0.1)*18)
				y2 := float32(sy) + float32(math.Sin(angle+0.1)*9) + 4
				vector.StrokeLine(screen, x1, y1, x2, y2, 2, selectGreen, false)
			}
		}

		// Unit body
		unitColor := color.RGBA{50, 120, 255, 255}
		if own.PlayerID != h.LocalPlayer {
			unitColor = color.RGBA{220, 50, 50, 255}
		}
		radius := float32(10)
		if w.Has(id, core.CompHarvester) {
			radius = 13
			if own.PlayerID == h.LocalPlayer {
				unitColor = color.RGBA{50, 200, 120, 255}
			}
		}
		// MCV is bigger
		if mcv := w.Get(id, core.CompMCV); mcv != nil {
			radius = 16
			if own.PlayerID == h.LocalPlayer {
				unitColor = color.RGBA{100, 80, 220, 255}
			}
		}

		// Try sprite first, fall back to circle
		spriteDrawn := false
		if h.UnitDrawFn != nil {
			spriteDrawn = h.UnitDrawFn(screen, w, id, sx, sy, own.PlayerID)
		}
		if !spriteDrawn {
			vector.DrawFilledCircle(screen, float32(sx), float32(sy), radius, unitColor, false)
			vector.StrokeCircle(screen, float32(sx), float32(sy), radius, 1.5, color.RGBA{255, 255, 255, 80}, false)
		}

		// Always show health bar above units
		if hp := w.Get(id, core.CompHealth); hp != nil {
			health := hp.(*core.Health)
			ratio := float32(health.Ratio())
			barW := float32(28)
			barH := float32(3)
			barX := float32(sx) - barW/2
			barY := float32(sy) - radius - 8
			// Background
			vector.DrawFilledRect(screen, barX-1, barY-1, barW+2, barH+2, color.RGBA{0, 0, 0, 160}, false)
			// Health fill
			vector.DrawFilledRect(screen, barX, barY, barW*ratio, barH, healthBarColor(ratio), false)
		}

		// Smoke for damaged units
		if hp := w.Get(id, core.CompHealth); hp != nil {
			health := hp.(*core.Health)
			if health.Ratio() < 0.5 {
				// Animated smoke puffs
				phase := h.tick*2 + float64(id)*0.7
				smokeAlpha := uint8(40 + 30*math.Sin(phase))
				smokeY := float32(sy) - radius - 12 - float32(math.Sin(phase*0.5)*3)
				vector.DrawFilledCircle(screen, float32(sx)+2, smokeY, 4, color.RGBA{80, 80, 80, smokeAlpha}, false)
				vector.DrawFilledCircle(screen, float32(sx)-2, smokeY-3, 3, color.RGBA{60, 60, 60, smokeAlpha / 2}, false)
			}
		}

		// Harvester load indicator
		if harv := w.Get(id, core.CompHarvester); harv != nil {
			hv := harv.(*core.Harvester)
			if hv.Current > 0 {
				loadRatio := float32(hv.Current) / float32(hv.Capacity)
				barW := float32(24)
				barX := float32(sx) - barW/2
				barY := float32(sy) + radius + 3
				vector.DrawFilledRect(screen, barX, barY, barW*loadRatio, 2, color.RGBA{255, 215, 0, 220}, false)
			}
		}
	}

	// Draw building effects
	for _, id := range w.Query(core.CompPosition, core.CompBuilding, core.CompOwner) {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		own := w.Get(id, core.CompOwner).(*core.Owner)
		bldg := w.Get(id, core.CompBuilding).(*core.Building)
		sx, sy := worldToScreen(pos.X, pos.Y)
		bw := float32(bldg.SizeX * 20)
		bh := float32(bldg.SizeY * 12)

		bcolor := color.RGBA{40, 60, 180, 255}
		borderColor := color.RGBA{80, 120, 255, 200}
		if own.PlayerID != h.LocalPlayer {
			bcolor = color.RGBA{180, 40, 40, 255}
			borderColor = color.RGBA{255, 80, 80, 200}
		}

		// Building under construction animation
		if bc := w.Get(id, core.CompBuildingConstruction); bc != nil {
			constr := bc.(*core.BuildingConstruction)
			if !constr.Complete {
				// Draw partial building (from bottom up)
				builtH := bh * float32(constr.Progress)
				vector.DrawFilledRect(screen, float32(sx)-bw/2, float32(sy)+bh/2-builtH, bw, builtH, bcolor, false)
				// Scaffold lines
				vector.StrokeRect(screen, float32(sx)-bw/2, float32(sy)-bh/2, bw, bh, 1, color.RGBA{200, 200, 100, 100}, false)
				// Progress text
				pctText := fmt.Sprintf("%d%%", int(constr.Progress*100))
				ebitenutil.DebugPrintAt(screen, pctText, sx-10, sy-5)
				continue
			}
		}

		// Try sprite first, fall back to colored rect
		bldgSpriteDrawn := false
		if h.BuildingDrawFn != nil {
			bldgSpriteDrawn = h.BuildingDrawFn(screen, w, id, sx, sy)
		}
		if !bldgSpriteDrawn {
			drawRoundedRect(screen, float32(sx)-bw/2, float32(sy)-bh/2, bw, bh, 3, bcolor)
			drawRoundedRectStroke(screen, float32(sx)-bw/2, float32(sy)-bh/2, bw, bh, 3, borderColor)
			if bldg.IsConYard {
				vector.DrawFilledCircle(screen, float32(sx), float32(sy), 8, color.RGBA{255, 220, 50, 200}, false)
				vector.StrokeCircle(screen, float32(sx), float32(sy), 8, 1, color.RGBA{255, 255, 200, 255}, false)
			}
		}

		// Health bar
		if hp := w.Get(id, core.CompHealth); hp != nil {
			health := hp.(*core.Health)
			ratio := float32(health.Ratio())
			barW := bw + 4
			barX := float32(sx) - barW/2
			barY := float32(sy) - bh/2 - 7
			vector.DrawFilledRect(screen, barX-1, barY-1, barW+2, 5, color.RGBA{0, 0, 0, 160}, false)
			vector.DrawFilledRect(screen, barX, barY, barW*ratio, 3, healthBarColor(ratio), false)

			// Smoke/fire for damaged buildings
			if health.Ratio() < 0.5 {
				phase := h.tick*1.5 + float64(id)
				smokeAlpha := uint8(60 + 40*math.Sin(phase))
				vector.DrawFilledCircle(screen, float32(sx)+5, float32(sy)-bh/2-10-float32(math.Sin(phase)*2), 5, color.RGBA{80, 80, 80, smokeAlpha}, false)
				if health.Ratio() < 0.25 {
					// Fire
					fireAlpha := uint8(120 + 60*math.Sin(phase*2))
					vector.DrawFilledCircle(screen, float32(sx)-3, float32(sy)-2, 4, color.RGBA{255, 100, 0, fireAlpha}, false)
				}
			}
		}

		// Production progress bar
		if prod := w.Get(id, core.CompProduction); prod != nil {
			p := prod.(*core.Production)
			if len(p.Queue) > 0 {
				barW := bw
				barX := float32(sx) - barW/2
				barY := float32(sy) + bh/2 + 3
				vector.DrawFilledRect(screen, barX-1, barY-1, barW+2, 5, color.RGBA{0, 0, 0, 160}, false)
				vector.DrawFilledRect(screen, barX, barY, barW*float32(p.Progress), 3, color.RGBA{255, 200, 0, 255}, false)
				// Queue count badge
				if len(p.Queue) > 1 {
					badgeX := float32(sx) + bw/2 - 5
					badgeY := float32(sy) - bh/2 - 5
					vector.DrawFilledCircle(screen, badgeX, badgeY, 7, color.RGBA{220, 50, 50, 240}, false)
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", len(p.Queue)), int(badgeX)-3, int(badgeY)-6)
				}
			}
		}

		// Selected indicator
		selected := false
		for _, sid := range h.SelectedIDs {
			if sid == id {
				selected = true
				break
			}
		}
		if selected {
			vector.StrokeRect(screen, float32(sx)-bw/2-3, float32(sy)-bh/2-3, bw+6, bh+6, 2, selectGreen, false)
		}
	}

	// Draw visual effects
	for _, eff := range h.Effects {
		sx, sy := worldToScreen(eff.X, eff.Y)
		t := float32(eff.Timer / eff.MaxTime)
		switch eff.Kind {
		case "explosion":
			r := float32(eff.Size) * (0.5 + t*1.5)
			alpha := uint8(255 * (1 - t))
			vector.DrawFilledCircle(screen, float32(sx), float32(sy), r, color.RGBA{255, 150, 0, alpha}, false)
			vector.DrawFilledCircle(screen, float32(sx), float32(sy), r*0.6, color.RGBA{255, 255, 100, alpha}, false)
		case "sparkle":
			alpha := uint8(200 * (0.5 + 0.5*math.Sin(h.tick*8+float64(sx))))
			vector.DrawFilledCircle(screen, float32(sx), float32(sy), 2, color.RGBA{255, 255, 100, alpha}, false)
		}
	}
}

// DrawPlacementGhost draws the building placement preview
func (h *HUD) DrawPlacementGhost(screen *ebiten.Image, worldToScreen func(float64, float64) (int, int)) {
	if !h.Placement.Active {
		return
	}
	sx, sy := worldToScreen(float64(h.Placement.TileX), float64(h.Placement.TileY))
	bw := float32(h.Placement.SizeX * 20)
	bh := float32(h.Placement.SizeY * 12)

	ghostColor := color.RGBA{0, 200, 0, 100}
	borderColor := color.RGBA{0, 255, 0, 200}
	if !h.Placement.Valid {
		ghostColor = color.RGBA{200, 0, 0, 100}
		borderColor = color.RGBA{255, 0, 0, 200}
	}

	drawRoundedRect(screen, float32(sx)-bw/2, float32(sy)-bh/2, bw, bh, 3, ghostColor)
	drawRoundedRectStroke(screen, float32(sx)-bw/2, float32(sy)-bh/2, bw, bh, 3, borderColor)

	// Building name
	if bdef, ok := h.TechTree.Buildings[h.Placement.BuildingKey]; ok {
		ebitenutil.DebugPrintAt(screen, bdef.Name, sx-int(bw/2), sy-int(bh/2)-14)
	}
}

func (h *HUD) drawTopBar(screen *ebiten.Image) {
	// Metallic panel background
	panel := h.Sprites.GenerateTopBarPanel(h.ScreenW, h.TopBarHeight)
	if panel != nil {
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(panel, op)
	} else {
		drawRoundedRect(screen, 0, 0, float32(h.ScreenW), float32(h.TopBarHeight), 0, panelBG)
	}

	player := h.Players.GetPlayer(h.LocalPlayer)
	if player == nil {
		return
	}

	// Credits section with gold icon
	x := 15
	y := 10
	// Gold coin (glossy circle)
	vector.DrawFilledCircle(screen, float32(x+8), float32(y+10), 9, color.RGBA{255, 200, 0, 255}, false)
	vector.DrawFilledCircle(screen, float32(x+7), float32(y+9), 6, color.RGBA{255, 230, 100, 255}, false)
	vector.StrokeCircle(screen, float32(x+8), float32(y+10), 9, 1.5, color.RGBA{180, 140, 0, 255}, false)
	ebitenutil.DebugPrintAt(screen, "$", x+5, y+4)

	creditStr := fmt.Sprintf("$%d", int(h.DisplayCredits))
	ebitenutil.DebugPrintAt(screen, creditStr, x+24, y+4)

	// Power section with icon sprite
	px := 200
	hasPower := player.HasPower()

	if h.Sprites.IconPower != nil {
		h.Sprites.DrawIcon(screen, h.Sprites.IconPower, px+10, y+10, 18)
	} else {
		pwrIconClr := powerGreen
		if !hasPower && int(h.tick*4)%2 == 0 {
			pwrIconClr = powerRed
		}
		vector.DrawFilledCircle(screen, float32(px+8), float32(y+10), 8, pwrIconClr, false)
	}

	powerStr := fmt.Sprintf("%d / %d", player.Power, player.PowerUse)
	ebitenutil.DebugPrintAt(screen, powerStr, px+24, y+4)

	// Power bar with glossy sprite
	barX := px + 100
	barY := y + 2
	barW := 120
	barH := 16
	ratio := 1.0
	if player.PowerUse > 0 {
		ratio = float64(player.Power) / float64(player.PowerUse)
		if ratio > 1 {
			ratio = 1
		}
	}
	barType := "power"
	if !hasPower {
		barType = "health" // will use red bar
	}
	h.Sprites.DrawBar(screen, barX, barY, barW, barH, ratio, barType)

	if !hasPower {
		// Blinking warning
		if int(h.tick*3)%2 == 0 {
			ebitenutil.DebugPrintAt(screen, "⚠ LOW POWER", px+230, y+4)
		}
	}

	// FPS on far right
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.0f", ebiten.ActualFPS()), h.ScreenW-80, y+4)
}

func (h *HUD) drawSidebar(screen *ebiten.Image, w *core.World) {
	sx := h.ScreenW - h.SidebarWidth
	sy := h.TopBarHeight
	sh := h.ScreenH - h.TopBarHeight

	// Metallic panel background
	panel := h.Sprites.GenerateSidebarPanel(h.SidebarWidth, sh)
	if panel != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sx), float64(sy))
		screen.DrawImage(panel, op)
	} else {
		drawRoundedRect(screen, float32(sx), float32(sy), float32(h.SidebarWidth), float32(sh), 0, panelBG)
	}

	// Tab buttons using sci-fi button sprites
	tabNames := []string{"BUILD", "UNITS", "DEF"}
	tabW := (h.SidebarWidth - 20) / 3
	for i, name := range tabNames {
		tx := sx + 10 + i*tabW
		ty := sy + 8
		isActive := BuildTab(i) == h.ActiveTab
		state := "normal"
		if isActive {
			state = "active"
		}

		h.Sprites.DrawRectButton(screen, tx, ty, tabW-4, 24, state)
		textX := tx + (tabW-4)/2 - len(name)*3
		ebitenutil.DebugPrintAt(screen, name, textX, ty+7)
	}

	// Build items
	y := sy + 40
	player := h.Players.GetPlayer(h.LocalPlayer)

	switch h.ActiveTab {
	case TabBuildings:
		h.drawBuildingButtons(screen, w, sx, y, player)
	case TabUnits:
		h.drawUnitButtons(screen, w, sx, y, player)
	case TabDefense:
		ebitenutil.DebugPrintAt(screen, "Coming soon...", sx+20, y+20)
	}
}

func (h *HUD) drawBuildingButtons(screen *ebiten.Image, w *core.World, sx, startY int, player *core.Player) {
	y := startY
	bIdx := 0

	// Check if player has construction yard
	hasConYard := h.playerHasConYard(w)

	for key, bdef := range h.TechTree.Buildings {
		if bIdx >= 10 {
			break
		}
		if key == "construction_yard" {
			continue // can't manually build con yard
		}

		canAfford := player != nil && player.Credits >= bdef.Cost
		hasPrereqs := true // simplified
		enabled := canAfford && hasPrereqs && hasConYard

		// Button
		btnX := float32(sx + 10)
		btnY := float32(y)
		btnW := float32(h.SidebarWidth - 20)
		btnH := float32(48)

		clr := btnNormal
		if !enabled {
			clr = btnDisabled
		} else if h.HoverBuildIdx == bIdx {
			clr = btnHover
		}

		drawRoundedRect(screen, btnX, btnY, btnW, btnH, 6, clr)

		// Icon placeholder (colored square)
		iconClr := accentBlue
		if !enabled {
			iconClr = color.RGBA{40, 40, 50, 200}
		}
		drawRoundedRect(screen, btnX+4, btnY+4, 40, 40, 4, iconClr)

		// Building initial letter as icon
		if len(bdef.Name) > 0 {
			ebitenutil.DebugPrintAt(screen, string(bdef.Name[0]), int(btnX)+18, int(btnY)+16)
		}

		// Name + Cost
		nameClr := textWhite
		if !enabled {
			nameClr = textDim
		}
		_ = nameClr
		ebitenutil.DebugPrintAt(screen, bdef.Name, int(btnX)+48, int(btnY)+8)
		costStr := fmt.Sprintf("$%d", bdef.Cost)
		costClr := textDim
		if !canAfford {
			costClr = healthRed
		}
		_ = costClr
		ebitenutil.DebugPrintAt(screen, costStr, int(btnX)+48, int(btnY)+22)

		// Power info
		if bdef.PowerGen > 0 {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("+%d⚡", bdef.PowerGen), int(btnX)+48, int(btnY)+34)
		} else if bdef.PowerDraw > 0 {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("-%d⚡", bdef.PowerDraw), int(btnX)+48, int(btnY)+34)
		}

		// Build progress overlay
		if prog, ok := h.BuildProgress[key]; ok && prog > 0 && prog < 1 {
			vector.DrawFilledRect(screen, btnX, btnY+btnH-4, btnW*float32(prog), 4, accentCyan, false)
		}

		// Border
		borderClr := panelBorder
		if enabled && h.HoverBuildIdx == bIdx {
			borderClr = accentCyan
		}
		drawRoundedRectStroke(screen, btnX, btnY, btnW, btnH, 6, borderClr)

		y += int(btnH) + 4
		bIdx++
	}

	if !hasConYard {
		ebitenutil.DebugPrintAt(screen, "⚠ Need Con. Yard", sx+20, y+10)
	}
}

func (h *HUD) drawUnitButtons(screen *ebiten.Image, w *core.World, sx, startY int, player *core.Player) {
	y := startY
	uIdx := 0
	for _, udef := range h.TechTree.Units {
		if uIdx >= 10 {
			break
		}

		canAfford := player != nil && player.Credits >= udef.Cost
		enabled := canAfford

		btnX := float32(sx + 10)
		btnY := float32(y)
		btnW := float32(h.SidebarWidth - 20)
		btnH := float32(40)

		clr := btnNormal
		if !enabled {
			clr = btnDisabled
		}

		drawRoundedRect(screen, btnX, btnY, btnW, btnH, 6, clr)

		// Unit icon (circle)
		iconClr := color.RGBA{50, 130, 50, 255}
		if !enabled {
			iconClr = color.RGBA{40, 40, 50, 200}
		}
		vector.DrawFilledCircle(screen, btnX+24, btnY+20, 14, iconClr, false)
		if len(udef.Name) > 0 {
			ebitenutil.DebugPrintAt(screen, string(udef.Name[0]), int(btnX)+20, int(btnY)+14)
		}

		ebitenutil.DebugPrintAt(screen, udef.Name, int(btnX)+44, int(btnY)+6)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("$%d", udef.Cost), int(btnX)+44, int(btnY)+20)

		drawRoundedRectStroke(screen, btnX, btnY, btnW, btnH, 6, panelBorder)

		y += int(btnH) + 4
		uIdx++
	}
}

func (h *HUD) drawBottomPanel(screen *ebiten.Image, w *core.World) {
	if len(h.SelectedIDs) == 0 && h.CurrentCommand == CmdNone {
		return
	}

	panelW := float32(h.ScreenW - h.SidebarWidth - h.MinimapSize - 20)
	panelX := float32(h.MinimapSize + 10)
	panelY := float32(h.ScreenH - h.BottomPanelH)

	drawRoundedRect(screen, panelX, panelY, panelW, float32(h.BottomPanelH), 8, panelBG)
	// Top accent
	vector.DrawFilledRect(screen, panelX+8, panelY, panelW-16, 2, accentCyan, false)

	if len(h.SelectedIDs) == 0 {
		return
	}

	// Single unit info or multi-select
	if len(h.SelectedIDs) == 1 {
		h.drawSingleUnitInfo(screen, w, int(panelX)+10, int(panelY)+10)
	} else {
		h.drawMultiSelectInfo(screen, w, int(panelX)+10, int(panelY)+10)
	}

	// Command buttons
	h.drawCommandButtons(screen, int(panelX)+int(panelW)-250, int(panelY)+15)
}

func (h *HUD) drawSingleUnitInfo(screen *ebiten.Image, w *core.World, x, y int) {
	id := h.SelectedIDs[0]

	// Unit portrait
	portraitClr := accentBlue
	if w.Has(id, core.CompBuilding) {
		portraitClr = color.RGBA{60, 60, 160, 255}
	}
	drawRoundedRect(screen, float32(x), float32(y), 60, 60, 6, portraitClr)

	// Name
	name := "Unit"
	if w.Has(id, core.CompBuilding) {
		name = "Building"
	}
	if w.Has(id, core.CompHarvester) {
		name = "Harvester"
	}
	if w.Has(id, core.CompMCV) {
		name = "MCV"
	}
	ebitenutil.DebugPrintAt(screen, name, x+70, y+5)

	// Health bar
	if hp := w.Get(id, core.CompHealth); hp != nil {
		health := hp.(*core.Health)
		ratio := float32(health.Ratio())
		barW := float32(120)
		barX := float32(x + 70)
		barY := float32(y + 22)
		vector.DrawFilledRect(screen, barX, barY, barW, 8, color.RGBA{20, 20, 30, 200}, false)
		vector.DrawFilledRect(screen, barX, barY, barW*ratio, 8, healthBarColor(ratio), false)
		vector.StrokeRect(screen, barX, barY, barW, 8, 1, panelBorder, false)
		hpText := fmt.Sprintf("%d / %d", health.Current, health.Max)
		ebitenutil.DebugPrintAt(screen, hpText, int(barX), int(barY)+12)
	}

	// Weapon info
	if wep := w.Get(id, core.CompWeapon); wep != nil {
		weapon := wep.(*core.Weapon)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("⚔ DMG: %d  RNG: %.1f", weapon.Damage, weapon.Range), x+70, y+45)
	}

	// MCV deploy button
	if w.Has(id, core.CompMCV) {
		btnX := float32(x + 70)
		btnY := float32(y + 58)
		clr := accentBlue
		if h.CurrentCommand == CmdDeploy {
			clr = btnActive
		}
		drawRoundedRect(screen, btnX, btnY, 70, 20, 4, clr)
		ebitenutil.DebugPrintAt(screen, "DEPLOY", int(btnX)+12, int(btnY)+4)
	}
}

func (h *HUD) drawMultiSelectInfo(screen *ebiten.Image, w *core.World, x, y int) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d units selected", len(h.SelectedIDs)), x+10, y+5)

	// Show unit type icons in a row
	ix := x + 10
	count := 0
	for _, id := range h.SelectedIDs {
		if count >= 20 {
			break
		}
		clr := color.RGBA{50, 120, 255, 200}
		if w.Has(id, core.CompHarvester) {
			clr = color.RGBA{50, 200, 120, 200}
		}
		if w.Has(id, core.CompMCV) {
			clr = color.RGBA{100, 80, 220, 200}
		}
		vector.DrawFilledCircle(screen, float32(ix+8), float32(y+30), 8, clr, false)

		// Mini health bar
		if hp := w.Get(id, core.CompHealth); hp != nil {
			health := hp.(*core.Health)
			ratio := float32(health.Ratio())
			vector.DrawFilledRect(screen, float32(ix), float32(y+42), 16*ratio, 2, healthBarColor(ratio), false)
		}

		ix += 20
		count++
	}
}

func (h *HUD) drawCommandButtons(screen *ebiten.Image, x, y int) {
	cmds := []struct {
		icon string
		name string
		cmd  CommandType
	}{
		{"🏃", "Move", CmdMove},
		{"⚔", "Attack", CmdAttack},
		{"■", "Stop", CmdStop},
		{"🛡", "Guard", CmdGuard},
		{"📍", "Rally", CmdRally},
	}

	for i, c := range cmds {
		bx := float32(x + i*48)
		by := float32(y)
		active := h.CurrentCommand == c.cmd

		clr := btnNormal
		if active {
			clr = btnActive
		} else if h.HoverCmdIdx == i {
			clr = btnHover
		}

		drawRoundedRect(screen, bx, by, 42, 42, 6, clr)
		drawRoundedRectStroke(screen, bx, by, 42, 42, 6, panelBorder)

		// Icon
		ebitenutil.DebugPrintAt(screen, c.icon, int(bx)+14, int(by)+8)
		// Label
		ebitenutil.DebugPrintAt(screen, c.name, int(bx)+4, int(by)+28)
	}
}

func (h *HUD) drawMinimap(screen *ebiten.Image, w *core.World) {
	mx := float32(5)
	my := float32(h.ScreenH - h.MinimapSize - 5)
	mw := float32(h.MinimapSize)
	mh := float32(h.MinimapSize)

	// Frame
	drawRoundedRect(screen, mx-2, my-18, mw+4, mh+22, 6, panelBG)
	drawRoundedRectStroke(screen, mx-2, my-18, mw+4, mh+22, 6, panelBorder)

	// Title
	ebitenutil.DebugPrintAt(screen, "TACTICAL MAP", int(mx)+30, int(my)-14)

	// Minimap content area
	vector.DrawFilledRect(screen, mx, my, mw, mh, minimapBG, false)

	// Units as dots
	for _, id := range w.Query(core.CompPosition, core.CompOwner) {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		own := w.Get(id, core.CompOwner).(*core.Owner)

		// Scale to minimap
		dotX := mx + float32(pos.X/64.0)*mw
		dotY := my + float32(pos.Y/64.0)*mh

		dotClr := color.RGBA{60, 140, 255, 255} // player = blue
		dotR := float32(2)
		if own.PlayerID != h.LocalPlayer {
			dotClr = color.RGBA{255, 60, 60, 255} // enemy = red
		}
		if w.Has(id, core.CompBuilding) {
			dotR = 3
		}

		vector.DrawFilledCircle(screen, dotX, dotY, dotR, dotClr, false)
	}
}

// ---- Ore Sparkle Drawing ----

func (h *HUD) DrawOreSparkles(screen *ebiten.Image, tileX, tileY int, oreAmount int, screenX, screenY int) {
	if oreAmount <= 0 {
		return
	}
	// Animated sparkle
	phase := h.tick*3 + float64(tileX*7+tileY*13)
	alpha := uint8(120 + 80*math.Sin(phase))
	vector.DrawFilledCircle(screen, float32(screenX)+2, float32(screenY)+2, 2, color.RGBA{255, 255, 100, alpha}, false)
	alpha2 := uint8(120 + 80*math.Sin(phase+1.5))
	vector.DrawFilledCircle(screen, float32(screenX)-3, float32(screenY)-1, 1.5, color.RGBA{255, 220, 50, alpha2}, false)
}

// DrawWaterEffect draws animated water overlay
func (h *HUD) DrawWaterEffect(screen *ebiten.Image, screenX, screenY, tw, th int) {
	phase := h.tick*1.5 + float64(screenX*3+screenY*5)*0.01
	shift := int(3 * math.Sin(phase))
	alpha := uint8(20 + 15*math.Sin(phase*0.7))
	vector.DrawFilledRect(screen, float32(screenX+shift), float32(screenY), float32(tw), float32(th), color.RGBA{100, 180, 255, alpha}, false)
}

// ---- Input Handling ----

func (h *HUD) HandleClick(mx, my int) bool {
	// Tab clicks in sidebar
	if mx >= h.ScreenW-h.SidebarWidth && my >= h.TopBarHeight && my < h.TopBarHeight+35 {
		tabW := (h.SidebarWidth - 20) / 3
		tabIdx := (mx - (h.ScreenW - h.SidebarWidth) - 10) / tabW
		if tabIdx >= 0 && tabIdx < 3 {
			h.ActiveTab = BuildTab(tabIdx)
			return true
		}
	}

	// Sidebar build buttons
	if mx >= h.ScreenW-h.SidebarWidth && my >= h.TopBarHeight+35 {
		return true
	}

	// Command buttons
	panelX := h.MinimapSize + 10
	panelY := h.ScreenH - h.BottomPanelH
	cmdX := panelX + (h.ScreenW - h.SidebarWidth - h.MinimapSize - 20) - 250
	cmdY := panelY + 15
	cmds := []CommandType{CmdMove, CmdAttack, CmdStop, CmdGuard, CmdRally}
	for i, c := range cmds {
		bx := cmdX + i*48
		if mx >= bx && mx < bx+42 && my >= cmdY && my < cmdY+42 {
			if h.CurrentCommand == c {
				h.CurrentCommand = CmdNone
			} else {
				h.CurrentCommand = c
			}
			return true
		}
	}

	// Minimap click
	if mx >= 5 && mx < 5+h.MinimapSize && my >= h.ScreenH-h.MinimapSize-5 && my < h.ScreenH-5 {
		return true // consumed, caller should move camera
	}

	return false
}

// GetMinimapWorldPos converts a minimap click to world coordinates
func (h *HUD) GetMinimapWorldPos(mx, my, mapSize int) (float64, float64) {
	relX := float64(mx-5) / float64(h.MinimapSize)
	relY := float64(my-(h.ScreenH-h.MinimapSize-5)) / float64(h.MinimapSize)
	return relX * float64(mapSize), relY * float64(mapSize)
}

// IsInMinimap checks if click is in minimap area
func (h *HUD) IsInMinimap(mx, my int) bool {
	return mx >= 5 && mx < 5+h.MinimapSize &&
		my >= h.ScreenH-h.MinimapSize-5 && my < h.ScreenH-5
}

// GetSidebarBuildingClick returns the building key if a building button was clicked
func (h *HUD) GetSidebarBuildingClick(mx, my int, w *core.World) string {
	if h.ActiveTab != TabBuildings {
		return ""
	}
	if mx < h.ScreenW-h.SidebarWidth || my < h.TopBarHeight+40 {
		return ""
	}

	relY := my - (h.TopBarHeight + 40)
	idx := relY / 52 // 48px button + 4px gap
	
	bIdx := 0
	for key := range h.TechTree.Buildings {
		if key == "construction_yard" {
			continue
		}
		if bIdx == idx {
			return key
		}
		bIdx++
	}
	return ""
}

// GetSidebarUnitClick returns the unit key if a unit button was clicked
func (h *HUD) GetSidebarUnitClick(mx, my int) string {
	if h.ActiveTab != TabUnits {
		return ""
	}
	if mx < h.ScreenW-h.SidebarWidth || my < h.TopBarHeight+40 {
		return ""
	}

	relY := my - (h.TopBarHeight + 40)
	idx := relY / 44 // 40px button + 4px gap

	uIdx := 0
	for key := range h.TechTree.Units {
		if uIdx == idx {
			return key
		}
		uIdx++
	}
	return ""
}

func (h *HUD) playerHasConYard(w *core.World) bool {
	for _, id := range w.Query(core.CompBuilding, core.CompOwner) {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		if own.PlayerID != h.LocalPlayer {
			continue
		}
		bldg := w.Get(id, core.CompBuilding).(*core.Building)
		if bldg.IsConYard {
			return true
		}
	}
	return false
}

// ---- Control Groups ----

func (h *HUD) AssignControlGroup(n int) {
	if n < 0 || n > 9 {
		return
	}
	h.ControlGroups[n] = make([]core.EntityID, len(h.SelectedIDs))
	copy(h.ControlGroups[n], h.SelectedIDs)
}

func (h *HUD) RecallControlGroup(n int) {
	if n < 0 || n > 9 {
		return
	}
	h.SelectedIDs = make([]core.EntityID, len(h.ControlGroups[n]))
	copy(h.SelectedIDs, h.ControlGroups[n])
}

func (h *HUD) IsInSidebar(mx, _ int) bool {
	return mx >= h.ScreenW-h.SidebarWidth
}

// CancelPlacement cancels building placement mode
func (h *HUD) CancelPlacement() {
	h.Placement.Active = false
	h.Placement.BuildingKey = ""
}

// StartPlacement enters building placement mode
func (h *HUD) StartPlacement(key string) {
	bdef, ok := h.TechTree.Buildings[key]
	if !ok {
		return
	}
	h.Placement.Active = true
	h.Placement.BuildingKey = key
	h.Placement.SizeX = bdef.SizeX
	h.Placement.SizeY = bdef.SizeY
}

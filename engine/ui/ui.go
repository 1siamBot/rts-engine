package ui

import (
	"fmt"
	"image/color"
	"math"
	"sort"

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
	CmdRepair
	CmdRally
	CmdWaypoint
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
	Active       bool
	BuildingKey  string
	SizeX, SizeY int
	Valid        bool
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

// SidebarBuildItem represents one item in the build grid
type SidebarBuildItem struct {
	Key       string
	Name      string
	Cost      int
	Enabled   bool
	CanAfford bool
	HasPrereqs bool
	Progress  float64 // 0-1, building/training progress
	Ready     bool    // construction complete, awaiting placement
	QueueCount int    // for units: how many in queue
	IsBuilding bool   // true = building, false = unit
	Tooltip   string  // requirement text
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

	// Sidebar mode: repair/sell cursor mode
	RepairMode bool
	SellMode   bool

	// Animated credits display
	DisplayCredits float64
	ActualCredits  int

	// Hover state
	HoverBuildIdx  int
	HoverCmdIdx    int
	HoverSidebar   bool

	// Build progress tracking for sidebar (building key -> progress 0-1)
	BuildProgress map[string]float64
	// Building construction state: key -> true means ready for placement
	BuildReady    map[string]bool

	// Scroll offset for build grid
	ScrollOffset int

	// Repair target tracking
	RepairTargetID core.EntityID

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

	// Status message (e.g. "Insufficient Funds")
	statusMsg     string
	statusMsgTime float64

	// Sprite draw callbacks (set externally to use real sprites)
	UnitDrawFn     func(screen *ebiten.Image, w *core.World, id core.EntityID, sx, sy int, playerID int) bool
	BuildingDrawFn func(screen *ebiten.Image, w *core.World, id core.EntityID, sx, sy int) bool
}

// RA2 sidebar layout constants
const (
	sidebarCreditsH   = 30  // credits + power display height
	sidebarCmdBtnH    = 32  // repair/sell/waypoint button row height
	sidebarTabH       = 28  // tab bar height
	sidebarSlotSize   = 86  // each build slot is square (was 80, slightly bigger)
	sidebarSlotGap    = 4   // gap between slots
	sidebarPadding    = 6   // padding inside sidebar
	sidebarPowerBarW  = 12  // power bar width on left edge
	sidebarScrollH    = 20  // scroll arrow height
)

func NewHUD(sw, sh int, tt *systems.TechTree, pm *core.PlayerManager, localPlayer int) *HUD {
	return &HUD{
		ScreenW:        sw,
		ScreenH:        sh,
		SidebarWidth:   200,
		TopBarHeight:   0, // No separate top bar; credits are in sidebar
		BottomPanelH:   100,
		MinimapSize:    160,
		TechTree:       tt,
		Players:        pm,
		LocalPlayer:    localPlayer,
		HoverBuildIdx:  -1,
		HoverCmdIdx:    -1,
		BuildProgress:  make(map[string]float64),
		BuildReady:     make(map[string]bool),
		panelCache:     make(map[string]*ebiten.Image),
		Sprites:        NewUISprites(),
	}
}

// ---- Drawing Helpers ----

func drawRoundedRect(screen *ebiten.Image, x, y, w, h float32, r float32, clr color.RGBA) {
	vector.DrawFilledRect(screen, x+r, y, w-2*r, h, clr, false)
	vector.DrawFilledRect(screen, x, y+r, w, h-2*r, clr, false)
	vector.DrawFilledCircle(screen, x+r, y+r, r, clr, false)
	vector.DrawFilledCircle(screen, x+w-r, y+r, r, clr, false)
	vector.DrawFilledCircle(screen, x+r, y+h-r, r, clr, false)
	vector.DrawFilledCircle(screen, x+w-r, y+h-r, r, clr, false)
}

func drawRoundedRectStroke(screen *ebiten.Image, x, y, w, h float32, r float32, clr color.RGBA) {
	vector.StrokeLine(screen, x+r, y, x+w-r, y, 1, clr, false)
	vector.StrokeLine(screen, x+r, y+h, x+w-r, y+h, 1, clr, false)
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

	// RA2 dark metal colors
	ra2MetalDark   = color.RGBA{35, 38, 42, 255}
	ra2MetalMid    = color.RGBA{55, 60, 65, 255}
	ra2MetalLight  = color.RGBA{75, 80, 88, 255}
	ra2SlotBG      = color.RGBA{22, 24, 28, 240}
	ra2SlotBorder  = color.RGBA{60, 65, 72, 255}
	ra2TabActive   = color.RGBA{70, 80, 95, 255}
	ra2TabInactive = color.RGBA{40, 44, 50, 255}
	ra2Gold        = color.RGBA{220, 190, 60, 255}
	ra2ReadyGreen  = color.RGBA{0, 255, 0, 255}
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
			h.DisplayCredits += diff * dt * 5
		}
	}

	// Update status message
	if h.statusMsgTime > 0 {
		h.statusMsgTime -= dt
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

// ShowMessage displays a temporary status message on screen
func (h *HUD) ShowMessage(msg string, duration float64) {
	h.statusMsg = msg
	h.statusMsgTime = duration
}

func (h *HUD) AddEffect(x, y float64, kind string, size float64) {
	h.Effects = append(h.Effects, Effect{
		X: x, Y: y, Kind: kind, Size: size,
		MaxTime: 1.0,
	})
}

// ---- Draw ----

func (h *HUD) Draw(screen *ebiten.Image, w *core.World) {
	h.drawSidebar(screen, w)
	h.drawBottomPanel(screen, w)
	h.drawMinimap(screen, w)

	// Status message (e.g. "Insufficient Funds")
	if h.statusMsgTime > 0 && h.statusMsg != "" {
		alpha := uint8(255)
		if h.statusMsgTime < 0.5 {
			alpha = uint8(h.statusMsgTime / 0.5 * 255)
		}
		msgW := len(h.statusMsg)*7 + 20
		msgX := (h.ScreenW-h.SidebarWidth)/2 - msgW/2
		msgY := h.ScreenH/2 - 40
		drawRoundedRect(screen, float32(msgX), float32(msgY), float32(msgW), 28, 6, color.RGBA{180, 30, 30, alpha})
		ebitenutil.DebugPrintAt(screen, h.statusMsg, msgX+10, msgY+8)
	}

	// Repair/Sell cursor indicator
	if h.RepairMode {
		ebitenutil.DebugPrintAt(screen, "🔧 REPAIR MODE - Click a building", 10, 10)
	}
	if h.SellMode {
		ebitenutil.DebugPrintAt(screen, "💰 SELL MODE - Click a building", 10, 10)
	}
}

// DrawWorldEffects draws selection circles, health bars above units, and effects
func (h *HUD) DrawWorldEffects(screen *ebiten.Image, w *core.World, worldToScreen func(float64, float64) (int, int)) {
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

		if selected {
			for angle := 0.0; angle < math.Pi*2; angle += 0.1 {
				x1 := float32(sx) + float32(math.Cos(angle)*18)
				y1 := float32(sy) + float32(math.Sin(angle)*9) + 4
				x2 := float32(sx) + float32(math.Cos(angle+0.1)*18)
				y2 := float32(sy) + float32(math.Sin(angle+0.1)*9) + 4
				vector.StrokeLine(screen, x1, y1, x2, y2, 2, selectGreen, false)
			}
		}

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
		if mcv := w.Get(id, core.CompMCV); mcv != nil {
			radius = 16
			if own.PlayerID == h.LocalPlayer {
				unitColor = color.RGBA{100, 80, 220, 255}
			}
		}

		spriteDrawn := false
		if h.UnitDrawFn != nil {
			spriteDrawn = h.UnitDrawFn(screen, w, id, sx, sy, own.PlayerID)
		}
		if !spriteDrawn {
			vector.DrawFilledCircle(screen, float32(sx), float32(sy), radius, unitColor, false)
			vector.StrokeCircle(screen, float32(sx), float32(sy), radius, 1.5, color.RGBA{255, 255, 255, 80}, false)
		}

		if hp := w.Get(id, core.CompHealth); hp != nil {
			health := hp.(*core.Health)
			ratio := float32(health.Ratio())
			barW := float32(28)
			barH := float32(3)
			barX := float32(sx) - barW/2
			barY := float32(sy) - radius - 8
			vector.DrawFilledRect(screen, barX-1, barY-1, barW+2, barH+2, color.RGBA{0, 0, 0, 160}, false)
			vector.DrawFilledRect(screen, barX, barY, barW*ratio, barH, healthBarColor(ratio), false)
		}

		if hp := w.Get(id, core.CompHealth); hp != nil {
			health := hp.(*core.Health)
			if health.Ratio() < 0.5 {
				phase := h.tick*2 + float64(id)*0.7
				smokeAlpha := uint8(40 + 30*math.Sin(phase))
				smokeY := float32(sy) - radius - 12 - float32(math.Sin(phase*0.5)*3)
				vector.DrawFilledCircle(screen, float32(sx)+2, smokeY, 4, color.RGBA{80, 80, 80, smokeAlpha}, false)
				vector.DrawFilledCircle(screen, float32(sx)-2, smokeY-3, 3, color.RGBA{60, 60, 60, smokeAlpha / 2}, false)
			}
		}

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

		if bc := w.Get(id, core.CompBuildingConstruction); bc != nil {
			constr := bc.(*core.BuildingConstruction)
			if !constr.Complete {
				builtH := bh * float32(constr.Progress)
				vector.DrawFilledRect(screen, float32(sx)-bw/2, float32(sy)+bh/2-builtH, bw, builtH, bcolor, false)
				vector.StrokeRect(screen, float32(sx)-bw/2, float32(sy)-bh/2, bw, bh, 1, color.RGBA{200, 200, 100, 100}, false)
				pctText := fmt.Sprintf("%d%%", int(constr.Progress*100))
				ebitenutil.DebugPrintAt(screen, pctText, sx-10, sy-5)
				continue
			}
		}

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

		if hp := w.Get(id, core.CompHealth); hp != nil {
			health := hp.(*core.Health)
			ratio := float32(health.Ratio())
			barW := bw + 4
			barX := float32(sx) - barW/2
			barY := float32(sy) - bh/2 - 7
			vector.DrawFilledRect(screen, barX-1, barY-1, barW+2, 5, color.RGBA{0, 0, 0, 160}, false)
			vector.DrawFilledRect(screen, barX, barY, barW*ratio, 3, healthBarColor(ratio), false)

			if health.Ratio() < 0.5 {
				phase := h.tick*1.5 + float64(id)
				smokeAlpha := uint8(60 + 40*math.Sin(phase))
				vector.DrawFilledCircle(screen, float32(sx)+5, float32(sy)-bh/2-10-float32(math.Sin(phase)*2), 5, color.RGBA{80, 80, 80, smokeAlpha}, false)
				if health.Ratio() < 0.25 {
					fireAlpha := uint8(120 + 60*math.Sin(phase*2))
					vector.DrawFilledCircle(screen, float32(sx)-3, float32(sy)-2, 4, color.RGBA{255, 100, 0, fireAlpha}, false)
				}
			}
		}

		if prod := w.Get(id, core.CompProduction); prod != nil {
			p := prod.(*core.Production)
			if len(p.Queue) > 0 {
				barW := bw
				barX := float32(sx) - barW/2
				barY := float32(sy) + bh/2 + 3
				vector.DrawFilledRect(screen, barX-1, barY-1, barW+2, 5, color.RGBA{0, 0, 0, 160}, false)
				vector.DrawFilledRect(screen, barX, barY, barW*float32(p.Progress), 3, color.RGBA{255, 200, 0, 255}, false)
				if len(p.Queue) > 1 {
					badgeX := float32(sx) + bw/2 - 5
					badgeY := float32(sy) - bh/2 - 5
					vector.DrawFilledCircle(screen, badgeX, badgeY, 7, color.RGBA{220, 50, 50, 240}, false)
					ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", len(p.Queue)), int(badgeX)-3, int(badgeY)-6)
				}
			}
		}

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

	if bdef, ok := h.TechTree.Buildings[h.Placement.BuildingKey]; ok {
		ebitenutil.DebugPrintAt(screen, bdef.Name, sx-int(bw/2), sy-int(bh/2)-14)
	}
}

// ======================== RA2-STYLE SIDEBAR ========================

func (h *HUD) drawSidebar(screen *ebiten.Image, w *core.World) {
	sx := h.ScreenW - h.SidebarWidth
	sy := 0
	sh := h.ScreenH

	// Dark brushed metal background
	panel := h.Sprites.GenerateSidebarPanel(h.SidebarWidth, sh)
	if panel != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sx), float64(sy))
		screen.DrawImage(panel, op)
	} else {
		vector.DrawFilledRect(screen, float32(sx), float32(sy), float32(h.SidebarWidth), float32(sh), ra2MetalDark, false)
		// Brushed metal lines
		for i := 0; i < sh; i += 3 {
			alpha := uint8(8 + 4*math.Sin(float64(i)*0.2))
			vector.DrawFilledRect(screen, float32(sx), float32(i), float32(h.SidebarWidth), 1, color.RGBA{255, 255, 255, alpha}, false)
		}
	}

	// Left edge bevel
	vector.DrawFilledRect(screen, float32(sx), 0, 2, float32(sh), color.RGBA{20, 22, 26, 255}, false)
	vector.DrawFilledRect(screen, float32(sx)+2, 0, 1, float32(sh), color.RGBA{80, 85, 95, 180}, false)

	curY := sy + sidebarPadding

	// ---- 1. Credits + Power display ----
	curY = h.drawSidebarCredits(screen, sx, curY)

	// ---- 2. Power bar (vertical, on left edge of sidebar) ----
	h.drawSidebarPowerBar(screen, sx, sy, sh)

	// ---- 3. Repair / Sell / Waypoint buttons ----
	curY = h.drawSidebarCmdButtons(screen, sx, curY)

	// ---- 4. Tab bar ----
	curY = h.drawSidebarTabs(screen, sx, curY)

	// ---- 5. Build grid (2 columns) ----
	h.drawSidebarBuildGrid(screen, w, sx, curY)
}

func (h *HUD) drawSidebarCredits(screen *ebiten.Image, sx, y int) int {
	player := h.Players.GetPlayer(h.LocalPlayer)
	if player == nil {
		return y + sidebarCreditsH
	}

	// Credits background strip
	vector.DrawFilledRect(screen, float32(sx+sidebarPowerBarW), float32(y), float32(h.SidebarWidth-sidebarPowerBarW), float32(sidebarCreditsH), color.RGBA{18, 20, 24, 240}, false)
	// Bottom divider
	vector.DrawFilledRect(screen, float32(sx+sidebarPowerBarW), float32(y+sidebarCreditsH-1), float32(h.SidebarWidth-sidebarPowerBarW), 1, ra2MetalLight, false)

	// Credits icon + scrolling number
	credX := sx + sidebarPowerBarW + 8
	// Gold coin
	vector.DrawFilledCircle(screen, float32(credX+6), float32(y+15), 7, color.RGBA{255, 200, 0, 255}, false)
	vector.DrawFilledCircle(screen, float32(credX+5), float32(y+14), 4, color.RGBA{255, 230, 100, 255}, false)
	ebitenutil.DebugPrintAt(screen, "$", credX+3, y+9)

	creditStr := fmt.Sprintf("$%d", int(h.DisplayCredits))
	ebitenutil.DebugPrintAt(screen, creditStr, credX+18, y+9)

	// Power display on right side
	pwrX := sx + h.SidebarWidth - 70
	hasPower := player.HasPower()
	pwrClr := powerGreen
	if !hasPower && int(h.tick*4)%2 == 0 {
		pwrClr = powerRed
	}
	vector.DrawFilledCircle(screen, float32(pwrX+6), float32(y+15), 5, pwrClr, false)
	ebitenutil.DebugPrintAt(screen, "⚡", pwrX, y+9)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d/%d", player.Power, player.PowerUse), pwrX+14, y+9)

	return y + sidebarCreditsH
}

func (h *HUD) drawSidebarPowerBar(screen *ebiten.Image, sx, sy, sh int) {
	player := h.Players.GetPlayer(h.LocalPlayer)
	if player == nil {
		return
	}

	barX := float32(sx + 2)
	barY := float32(sy + sidebarCreditsH + sidebarPadding)
	barH := float32(sh - sidebarCreditsH - sidebarPadding*2)
	barW := float32(sidebarPowerBarW - 4)

	// Background
	vector.DrawFilledRect(screen, barX, barY, barW, barH, color.RGBA{15, 15, 20, 240}, false)
	vector.StrokeRect(screen, barX, barY, barW, barH, 1, color.RGBA{50, 55, 62, 255}, false)

	// Power fill (bottom-up)
	ratio := 1.0
	if player.PowerUse > 0 {
		ratio = float64(player.Power) / float64(player.PowerUse)
		if ratio > 1 {
			ratio = 1
		}
	}
	fillH := barH * float32(ratio)
	fillColor := powerGreen
	if !player.HasPower() {
		fillColor = powerRed
		// Blink when low power
		if int(h.tick*3)%2 == 0 {
			fillColor = color.RGBA{255, 60, 60, 255}
		}
	}
	vector.DrawFilledRect(screen, barX+1, barY+barH-fillH, barW-2, fillH, fillColor, false)

	// Gradient highlight on the bar
	for i := 0; i < int(fillH); i++ {
		fy := barY + barH - float32(i)
		alpha := uint8(30 * math.Sin(float64(i)*0.1))
		vector.DrawFilledRect(screen, barX+1, fy, barW-2, 1, color.RGBA{255, 255, 255, alpha}, false)
	}
}

func (h *HUD) drawSidebarCmdButtons(screen *ebiten.Image, sx, y int) int {
	// Repair / Sell / Waypoint buttons row
	btnW := (h.SidebarWidth - sidebarPowerBarW - sidebarPadding*2 - 8) / 3
	btnH := sidebarCmdBtnH - 4
	startX := sx + sidebarPowerBarW + sidebarPadding

	type cmdBtn struct {
		label  string
		active bool
		icon   *ebiten.Image
	}
	buttons := []cmdBtn{
		{"🔧", h.RepairMode, h.Sprites.IconRepair},
		{"💰", h.SellMode, h.Sprites.IconSell},
		{"🏳", false, nil},
	}

	for i, btn := range buttons {
		bx := startX + i*(btnW+4)
		by := y + 2

		// Button background
		bgClr := ra2MetalMid
		if btn.active {
			bgClr = color.RGBA{0, 120, 180, 255}
		}
		vector.DrawFilledRect(screen, float32(bx), float32(by), float32(btnW), float32(btnH), bgClr, false)
		// Bevel
		vector.StrokeLine(screen, float32(bx), float32(by), float32(bx+btnW), float32(by), 1, ra2MetalLight, false)
		vector.StrokeLine(screen, float32(bx), float32(by), float32(bx), float32(by+btnH), 1, ra2MetalLight, false)
		vector.StrokeLine(screen, float32(bx+btnW), float32(by), float32(bx+btnW), float32(by+btnH), 1, color.RGBA{20, 22, 26, 255}, false)
		vector.StrokeLine(screen, float32(bx), float32(by+btnH), float32(bx+btnW), float32(by+btnH), 1, color.RGBA{20, 22, 26, 255}, false)

		// Icon or label
		if btn.icon != nil {
			h.Sprites.DrawIcon(screen, btn.icon, bx+btnW/2, by+btnH/2, 20)
		} else {
			ebitenutil.DebugPrintAt(screen, btn.label, bx+btnW/2-6, by+btnH/2-6)
		}
	}

	return y + sidebarCmdBtnH
}

func (h *HUD) drawSidebarTabs(screen *ebiten.Image, sx, y int) int {
	tabNames := []string{"BUILD", "UNITS", "DEF"}
	contentW := h.SidebarWidth - sidebarPowerBarW - sidebarPadding*2
	tabW := contentW / 3
	startX := sx + sidebarPowerBarW + sidebarPadding

	for i, name := range tabNames {
		tx := startX + i*tabW
		ty := y
		isActive := BuildTab(i) == h.ActiveTab

		// Tab background
		tabClr := ra2TabInactive
		if isActive {
			tabClr = ra2TabActive
		}

		// Use RA2 tab sprites if available
		h.Sprites.DrawTabButton(screen, tx, ty, tabW, sidebarTabH, isActive)

		// If no sprite was drawn, draw procedural tab
		if h.Sprites.RA2TabActive == nil {
			vector.DrawFilledRect(screen, float32(tx), float32(ty), float32(tabW), float32(sidebarTabH), tabClr, false)
			// Top highlight for active
			if isActive {
				vector.DrawFilledRect(screen, float32(tx), float32(ty), float32(tabW), 2, accentCyan, false)
			}
			// Bevels
			vector.StrokeLine(screen, float32(tx), float32(ty), float32(tx+tabW), float32(ty), 1, ra2MetalLight, false)
			vector.StrokeLine(screen, float32(tx+tabW-1), float32(ty), float32(tx+tabW-1), float32(ty+sidebarTabH), 1, color.RGBA{25, 28, 32, 255}, false)
		}

		textX := tx + tabW/2 - len(name)*3
		ebitenutil.DebugPrintAt(screen, name, textX, ty+9)
	}

	return y + sidebarTabH + 2
}

// getBuildItems returns the list of items for the current tab
func (h *HUD) getBuildItems(w *core.World) []SidebarBuildItem {
	player := h.Players.GetPlayer(h.LocalPlayer)
	hasConYard := h.PlayerHasConYard(w)

	var items []SidebarBuildItem

	switch h.ActiveTab {
	case TabBuildings:
		for _, key := range h.TechTree.BuildingKeyOrder() {
			bdef := h.TechTree.Buildings[key]
			canAfford := player != nil && player.Credits >= bdef.Cost
			hasPrereqs := h.TechTree.HasPrereqs(w, h.LocalPlayer, bdef.Prereqs)
			enabled := canAfford && hasPrereqs && hasConYard

			tooltip := ""
			if !hasConYard {
				tooltip = "Need Construction Yard"
			} else if !hasPrereqs {
				tooltip = "Requires: " + prereqNames(h.TechTree, bdef.Prereqs)
			} else if !canAfford {
				tooltip = "Insufficient Funds"
			}

			prog := h.BuildProgress[key]
			ready := h.BuildReady[key]

			items = append(items, SidebarBuildItem{
				Key: key, Name: bdef.Name, Cost: bdef.Cost,
				Enabled: enabled, CanAfford: canAfford, HasPrereqs: hasPrereqs && hasConYard,
				Progress: prog, Ready: ready, IsBuilding: true, Tooltip: tooltip,
			})
		}

	case TabUnits:
		for _, key := range h.TechTree.UnitKeyOrder() {
			udef := h.TechTree.Units[key]
			canAfford := player != nil && player.Credits >= udef.Cost
			hasPrereqs := h.TechTree.HasPrereqs(w, h.LocalPlayer, udef.Prereqs)
			hasProdBuilding := systems.FindProductionBuilding(w, h.TechTree, h.LocalPlayer, key) != 0
			enabled := canAfford && hasPrereqs && hasProdBuilding

			tooltip := ""
			if !hasProdBuilding {
				tooltip = "No production building"
			} else if !hasPrereqs {
				tooltip = "Requires: " + prereqNames(h.TechTree, udef.Prereqs)
			} else if !canAfford {
				tooltip = "Insufficient Funds"
			}

			// Get queue count and progress from production buildings
			queueCount, progress := h.getUnitQueueInfo(w, key)

			items = append(items, SidebarBuildItem{
				Key: key, Name: udef.Name, Cost: udef.Cost,
				Enabled: enabled, CanAfford: canAfford, HasPrereqs: hasPrereqs && hasProdBuilding,
				Progress: progress, QueueCount: queueCount, IsBuilding: false, Tooltip: tooltip,
			})
		}

	case TabDefense:
		for _, key := range h.TechTree.DefenseKeyOrder() {
			bdef := h.TechTree.Buildings[key]
			canAfford := player != nil && player.Credits >= bdef.Cost
			hasPrereqs := h.TechTree.HasPrereqs(w, h.LocalPlayer, bdef.Prereqs)
			enabled := canAfford && hasPrereqs && hasConYard

			tooltip := ""
			if !hasConYard {
				tooltip = "Need Construction Yard"
			} else if !hasPrereqs {
				tooltip = "Requires: " + prereqNames(h.TechTree, bdef.Prereqs)
			} else if !canAfford {
				tooltip = "Insufficient Funds"
			}

			prog := h.BuildProgress[key]
			ready := h.BuildReady[key]

			items = append(items, SidebarBuildItem{
				Key: key, Name: bdef.Name, Cost: bdef.Cost,
				Enabled: enabled, CanAfford: canAfford, HasPrereqs: hasPrereqs && hasConYard,
				Progress: prog, Ready: ready, IsBuilding: true, Tooltip: tooltip,
			})
		}
	}

	return items
}

func (h *HUD) getUnitQueueInfo(w *core.World, unitKey string) (int, float64) {
	totalQueue := 0
	var bestProgress float64

	for _, bid := range w.Query(core.CompProduction, core.CompOwner, core.CompBuildingName) {
		own := w.Get(bid, core.CompOwner).(*core.Owner)
		if own.PlayerID != h.LocalPlayer {
			continue
		}
		prod := w.Get(bid, core.CompProduction).(*core.Production)
		for i, qk := range prod.Queue {
			if qk == unitKey {
				totalQueue++
				if i == 0 {
					bestProgress = prod.Progress
				}
			}
		}
	}
	return totalQueue, bestProgress
}

func (h *HUD) drawSidebarBuildGrid(screen *ebiten.Image, w *core.World, sx, startY int) {
	items := h.getBuildItems(w)

	contentW := h.SidebarWidth - sidebarPowerBarW - sidebarPadding*2
	slotW := (contentW - sidebarSlotGap) / 2
	slotH := slotW // square slots
	gridStartX := sx + sidebarPowerBarW + sidebarPadding

	// Calculate visible slots
	availableH := h.ScreenH - startY - sidebarScrollH
	visibleRows := availableH / (slotH + sidebarSlotGap)
	if visibleRows < 1 {
		visibleRows = 1
	}
	visibleSlots := visibleRows * 2
	totalItems := len(items)

	// Clamp scroll
	maxScroll := totalItems - visibleSlots
	if maxScroll < 0 {
		maxScroll = 0
	}
	if h.ScrollOffset > maxScroll {
		h.ScrollOffset = maxScroll
	}
	if h.ScrollOffset < 0 {
		h.ScrollOffset = 0
	}

	// Draw build slots (2 columns)
	for i := 0; i < visibleSlots && i+h.ScrollOffset < totalItems; i++ {
		item := items[i+h.ScrollOffset]
		col := i % 2
		row := i / 2
		slotX := gridStartX + col*(slotW+sidebarSlotGap)
		slotY := startY + row*(slotH+sidebarSlotGap)

		h.drawBuildSlot(screen, slotX, slotY, slotW, slotH, item, i)
	}

	// Scroll arrows
	if totalItems > visibleSlots {
		arrowY := h.ScreenH - sidebarScrollH
		arrowX := gridStartX + contentW/2

		// Up arrow
		if h.ScrollOffset > 0 {
			vector.DrawFilledRect(screen, float32(gridStartX), float32(startY-16), float32(contentW), 14, ra2MetalMid, false)
			ebitenutil.DebugPrintAt(screen, "▲ scroll up", arrowX-30, startY-14)
		}
		// Down arrow
		if h.ScrollOffset < maxScroll {
			vector.DrawFilledRect(screen, float32(gridStartX), float32(arrowY), float32(contentW), float32(sidebarScrollH), ra2MetalMid, false)
			ebitenutil.DebugPrintAt(screen, "▼ scroll down", arrowX-36, arrowY+4)
		}
	}
}

func (h *HUD) drawBuildSlot(screen *ebiten.Image, x, y, w, hh int, item SidebarBuildItem, idx int) {
	// Slot background - use RA2 build slot textures
	isHovered := h.HoverBuildIdx == idx
	state := "normal"
	if !item.HasPrereqs {
		state = "disabled"
	} else if item.Progress > 0 && item.Progress < 1 {
		state = "active"
	} else if isHovered && item.Enabled {
		state = "hover"
	} else if !item.CanAfford {
		state = "disabled"
	}

	// Try RA2 slot texture
	var slotImg *ebiten.Image
	switch state {
	case "hover":
		slotImg = h.Sprites.RA2BuildSlotHover
	case "active":
		slotImg = h.Sprites.RA2BuildSlotActive
	case "disabled":
		slotImg = h.Sprites.RA2BuildSlotDisabled
	default:
		slotImg = h.Sprites.RA2BuildSlotNormal
	}

	if slotImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(w)/float64(slotImg.Bounds().Dx()), float64(hh)/float64(slotImg.Bounds().Dy()))
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(slotImg, op)
	} else {
		// Procedural slot
		bgClr := ra2SlotBG
		if state == "hover" {
			bgClr = color.RGBA{35, 40, 50, 255}
		} else if state == "disabled" {
			bgClr = color.RGBA{18, 18, 22, 200}
		} else if state == "active" {
			bgClr = color.RGBA{25, 35, 45, 255}
		}
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(hh), bgClr, false)
		// Beveled border
		vector.StrokeLine(screen, float32(x), float32(y), float32(x+w), float32(y), 1, ra2MetalLight, false)
		vector.StrokeLine(screen, float32(x), float32(y), float32(x), float32(y+hh), 1, ra2MetalLight, false)
		vector.StrokeLine(screen, float32(x+w), float32(y), float32(x+w), float32(y+hh), 1, color.RGBA{15, 16, 20, 255}, false)
		vector.StrokeLine(screen, float32(x), float32(y+hh), float32(x+w), float32(y+hh), 1, color.RGBA{15, 16, 20, 255}, false)
	}

	// Cameo icon
	if buildIcon := h.Sprites.GetBuildIcon(item.Key); buildIcon != nil {
		op := &ebiten.DrawImageOptions{}
		iw := buildIcon.Bounds().Dx()
		ih := buildIcon.Bounds().Dy()
		iconSize := w - 8
		op.GeoM.Scale(float64(iconSize)/float64(iw), float64(iconSize)/float64(ih))
		op.GeoM.Translate(float64(x)+4, float64(y)+2)
		if !item.HasPrereqs {
			op.ColorScale.Scale(0.3, 0.3, 0.3, 0.7)
		} else if !item.CanAfford {
			op.ColorScale.Scale(0.5, 0.5, 0.5, 0.9)
		}
		screen.DrawImage(buildIcon, op)
	} else {
		// Fallback: draw colored block with first letter
		iconClr := accentBlue
		if !item.HasPrereqs {
			iconClr = color.RGBA{30, 30, 40, 200}
		}
		iconM := 6
		drawIsoBlock(screen, float32(x+w/2), float32(y+hh/2-6), float32(w-iconM*2), float32((hh-iconM*2)/2), iconClr)
		if len(item.Name) > 0 {
			ebitenutil.DebugPrintAt(screen, string(item.Name[0]), x+w/2-3, y+hh/2-12)
		}
	}

	// Clock-wipe progress overlay (when building)
	if item.Progress > 0 && item.Progress < 1 {
		h.drawClockWipe(screen, x, y, w, hh, item.Progress)
	}

	// "READY" flashing text
	if item.Ready {
		if int(h.tick*4)%2 == 0 {
			readyX := x + w/2 - 18
			readyY := y + hh/2 - 6
			vector.DrawFilledRect(screen, float32(readyX-2), float32(readyY-2), 40, 16, color.RGBA{0, 0, 0, 180}, false)
			ebitenutil.DebugPrintAt(screen, "READY", readyX, readyY)
		}
	}

	// Queue count badge (for units)
	if item.QueueCount > 1 {
		badgeX := x + w - 12
		badgeY := y + 2
		vector.DrawFilledCircle(screen, float32(badgeX+6), float32(badgeY+6), 8, color.RGBA{220, 50, 50, 240}, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", item.QueueCount), badgeX+2, badgeY)
	}

	// Lock icon if prerequisites not met
	if !item.HasPrereqs {
		lockX := x + w/2 - 3
		lockY := y + hh/2 - 6
		ebitenutil.DebugPrintAt(screen, "🔒", lockX, lockY)
	}

	// Cost text at bottom of slot
	costStr := fmt.Sprintf("$%d", item.Cost)
	costClr := ra2Gold
	if !item.CanAfford {
		costClr = color.RGBA{180, 60, 60, 255}
	}
	_ = costClr // using debug print which is always white
	costX := x + w/2 - len(costStr)*3
	costY := y + hh - 12
	// Dark background strip for cost
	vector.DrawFilledRect(screen, float32(x), float32(costY-2), float32(w), 14, color.RGBA{0, 0, 0, 160}, false)
	ebitenutil.DebugPrintAt(screen, costStr, costX, costY)
}

// drawClockWipe draws a clock-wipe progress overlay on a build slot
func (h *HUD) drawClockWipe(screen *ebiten.Image, x, y, w, hh int, progress float64) {
	// Dark overlay for the un-built portion (clock sweep from top, clockwise)
	cx := float32(x) + float32(w)/2
	cy := float32(y) + float32(hh)/2
	r := float32(w) / 2

	// Draw dark overlay sectors for the remaining portion
	startAngle := -math.Pi / 2 // 12 o'clock
	sweepAngle := progress * 2 * math.Pi

	// Draw the "done" portion as slightly transparent
	// Draw the "remaining" portion as dark overlay
	remainStart := startAngle + sweepAngle
	remainEnd := startAngle + 2*math.Pi

	overlayClr := color.RGBA{0, 0, 0, 140}
	step := 0.05
	for angle := remainStart; angle < remainEnd; angle += step {
		x1 := cx + r*float32(math.Cos(angle))
		y1 := cy + r*float32(math.Sin(angle))
		x2 := cx + r*float32(math.Cos(angle+step))
		y2 := cy + r*float32(math.Sin(angle+step))

		// Draw triangle from center to arc segment
		vertices := []ebiten.Vertex{
			{DstX: cx, DstY: cy, SrcX: 0, SrcY: 0, ColorR: 0, ColorG: 0, ColorB: 0, ColorA: float32(overlayClr.A) / 255},
			{DstX: x1, DstY: y1, SrcX: 0, SrcY: 0, ColorR: 0, ColorG: 0, ColorB: 0, ColorA: float32(overlayClr.A) / 255},
			{DstX: x2, DstY: y2, SrcX: 0, SrcY: 0, ColorR: 0, ColorG: 0, ColorB: 0, ColorA: float32(overlayClr.A) / 255},
		}
		indices := []uint16{0, 1, 2}

		whiteImg := ebiten.NewImage(1, 1)
		whiteImg.Fill(color.White)
		opt := &ebiten.DrawTrianglesOptions{}
		opt.Blend = ebiten.BlendSourceOver
		screen.DrawTriangles(vertices, indices, whiteImg, opt)
	}

	// Progress text
	pctText := fmt.Sprintf("%d%%", int(progress*100))
	ebitenutil.DebugPrintAt(screen, pctText, x+w/2-12, y+hh/2-6)
}

// drawIsoBlock draws a small 3D isometric block for building icons
func drawIsoBlock(screen *ebiten.Image, cx, cy, w, h float32, clr color.RGBA) {
	topClr := color.RGBA{
		uint8(min(int(clr.R)+40, 255)),
		uint8(min(int(clr.G)+40, 255)),
		uint8(min(int(clr.B)+40, 255)),
		clr.A,
	}
	hw := w / 2
	hh := h / 4
	vector.DrawFilledRect(screen, cx-hw, cy-hh-h/2, w, h/2, topClr, false)
	vector.DrawFilledRect(screen, cx-hw, cy-hh, w/2, h/2, clr, false)
	sideClr := color.RGBA{
		uint8(max(int(clr.R)-30, 0)),
		uint8(max(int(clr.G)-30, 0)),
		uint8(max(int(clr.B)-30, 0)),
		clr.A,
	}
	vector.DrawFilledRect(screen, cx, cy-hh, w/2, h/2, sideClr, false)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ======================== BOTTOM PANEL ========================

func (h *HUD) drawBottomPanel(screen *ebiten.Image, w *core.World) {
	if len(h.SelectedIDs) == 0 && h.CurrentCommand == CmdNone {
		return
	}

	panelW := h.ScreenW - h.SidebarWidth - h.MinimapSize - 20
	panelX := h.MinimapSize + 10
	panelY := h.ScreenH - h.BottomPanelH

	panel := h.Sprites.GenerateBottomPanel(panelW, h.BottomPanelH)
	if panel != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(panelX), float64(panelY))
		screen.DrawImage(panel, op)
	} else {
		drawRoundedRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(h.BottomPanelH), 8, panelBG)
	}

	if len(h.SelectedIDs) == 0 {
		return
	}

	if len(h.SelectedIDs) == 1 {
		h.drawSingleUnitInfo(screen, w, panelX+10, panelY+10)
	} else {
		h.drawMultiSelectInfo(screen, w, panelX+10, panelY+10)
	}

	h.drawCommandButtons(screen, panelX+panelW-250, panelY+15)
}

func (h *HUD) drawSingleUnitInfo(screen *ebiten.Image, w *core.World, x, y int) {
	id := h.SelectedIDs[0]

	vector.DrawFilledRect(screen, float32(x), float32(y), 64, 64, color.RGBA{10, 15, 25, 240}, false)
	vector.StrokeLine(screen, float32(x), float32(y), float32(x+64), float32(y), 1, color.RGBA{70, 80, 100, 200}, false)
	vector.StrokeLine(screen, float32(x), float32(y), float32(x), float32(y+64), 1, color.RGBA{70, 80, 100, 200}, false)
	vector.StrokeLine(screen, float32(x+64), float32(y), float32(x+64), float32(y+64), 1, color.RGBA{15, 20, 30, 200}, false)
	vector.StrokeLine(screen, float32(x), float32(y+64), float32(x+64), float32(y+64), 1, color.RGBA{15, 20, 30, 200}, false)

	portraitClr := accentBlue
	if w.Has(id, core.CompBuilding) {
		portraitClr = color.RGBA{60, 60, 160, 255}
	}
	if w.Has(id, core.CompHarvester) {
		portraitClr = color.RGBA{50, 180, 100, 255}
	}
	if w.Has(id, core.CompMCV) {
		portraitClr = color.RGBA{120, 80, 200, 255}
	}
	cx, cy := float32(x+32), float32(y+32)
	vector.DrawFilledCircle(screen, cx+1, cy+1, 22, color.RGBA{0, 0, 0, 100}, false)
	vector.DrawFilledCircle(screen, cx, cy, 22, portraitClr, false)
	vector.DrawFilledCircle(screen, cx-6, cy-6, 10, color.RGBA{
		uint8(min(int(portraitClr.R)+60, 255)),
		uint8(min(int(portraitClr.G)+60, 255)),
		uint8(min(int(portraitClr.B)+60, 255)),
		100,
	}, false)
	vector.StrokeCircle(screen, cx, cy, 22, 1.5, color.RGBA{90, 100, 120, 200}, false)

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
	// Try to get building name
	if bn := w.Get(id, core.CompBuildingName); bn != nil {
		key := bn.(*core.BuildingName).Key
		if bdef, ok := h.TechTree.Buildings[key]; ok {
			name = bdef.Name
		}
	}
	ebitenutil.DebugPrintAt(screen, name, x+72, y+5)

	if hp := w.Get(id, core.CompHealth); hp != nil {
		health := hp.(*core.Health)
		ratio := health.Ratio()
		h.Sprites.DrawBar(screen, x+72, y+20, 130, 12, ratio, "health")
		hpText := fmt.Sprintf("%d / %d", health.Current, health.Max)
		ebitenutil.DebugPrintAt(screen, hpText, x+72, y+34)
	}

	if wep := w.Get(id, core.CompWeapon); wep != nil {
		weapon := wep.(*core.Weapon)
		if h.Sprites.IconAttack != nil {
			h.Sprites.DrawIcon(screen, h.Sprites.IconAttack, x+80, y+55, 12)
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("DMG:%d RNG:%.0f", weapon.Damage, weapon.Range), x+90, y+50)
	}

	if w.Has(id, core.CompMCV) {
		state := "normal"
		if h.CurrentCommand == CmdDeploy {
			state = "active"
		}
		h.Sprites.DrawRectButton(screen, x+72, y+62, 80, 22, state)
		if h.Sprites.IconDeploy != nil {
			h.Sprites.DrawIcon(screen, h.Sprites.IconDeploy, x+84, y+73, 14)
		}
		ebitenutil.DebugPrintAt(screen, "DEPLOY [H]", x+94, y+67)
	}
}

func (h *HUD) drawMultiSelectInfo(screen *ebiten.Image, w *core.World, x, y int) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d units selected", len(h.SelectedIDs)), x+10, y+5)

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
		name string
		cmd  CommandType
	}{
		{"Move", CmdMove},
		{"Attack", CmdAttack},
		{"Stop", CmdStop},
		{"Guard", CmdGuard},
		{"Rally", CmdRally},
	}

	for i, c := range cmds {
		bx := x + i*48
		by := y
		active := h.CurrentCommand == c.cmd

		state := "normal"
		if active {
			state = "active"
		} else if h.HoverCmdIdx == i {
			state = "hover"
		}

		h.Sprites.DrawButton(screen, bx, by, 44, 44, state)

		icon := h.Sprites.GetCommandIcon(c.cmd)
		if icon != nil {
			h.Sprites.DrawIcon(screen, icon, bx+22, by+18, 20)
		}

		textX := bx + 22 - len(c.name)*3
		ebitenutil.DebugPrintAt(screen, c.name, textX, by+32)
	}
}

// ======================== MINIMAP ========================

func (h *HUD) drawMinimap(screen *ebiten.Image, w *core.World) {
	mx := 5
	my := h.ScreenH - h.MinimapSize - 5
	mw := h.MinimapSize
	mh := h.MinimapSize

	frame := h.Sprites.GenerateMinimapFrame(h.MinimapSize)
	if frame != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(mx-4), float64(my-20))
		screen.DrawImage(frame, op)
	} else {
		drawRoundedRect(screen, float32(mx-2), float32(my-18), float32(mw+4), float32(mh+22), 6, panelBG)
		drawRoundedRectStroke(screen, float32(mx-2), float32(my-18), float32(mw+4), float32(mh+22), 6, panelBorder)
	}

	ebitenutil.DebugPrintAt(screen, "TACTICAL MAP", mx+30, my-16)
	vector.DrawFilledRect(screen, float32(mx), float32(my), float32(mw), float32(mh), minimapBG, false)

	// Radar sweep effect
	sweepAngle := h.tick * 0.8
	sweepCx := float32(mx) + float32(mw)/2
	sweepCy := float32(my) + float32(mh)/2
	sweepR := float32(mw) / 2
	endX := sweepCx + sweepR*float32(math.Cos(sweepAngle))
	endY := sweepCy + sweepR*float32(math.Sin(sweepAngle))
	vector.StrokeLine(screen, sweepCx, sweepCy, endX, endY, 1, color.RGBA{0, 200, 100, 60}, false)
	for i := 1; i < 8; i++ {
		trailAngle := sweepAngle - float64(i)*0.08
		tEndX := sweepCx + sweepR*float32(math.Cos(trailAngle))
		tEndY := sweepCy + sweepR*float32(math.Sin(trailAngle))
		alpha := uint8(40 - i*5)
		if alpha > 40 {
			alpha = 0
		}
		vector.StrokeLine(screen, sweepCx, sweepCy, tEndX, tEndY, 1, color.RGBA{0, 200, 100, alpha}, false)
	}

	for _, id := range w.Query(core.CompPosition, core.CompOwner) {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		own := w.Get(id, core.CompOwner).(*core.Owner)

		dotX := float32(mx) + float32(pos.X/64.0)*float32(mw)
		dotY := float32(my) + float32(pos.Y/64.0)*float32(mh)

		dotClr := color.RGBA{60, 140, 255, 255}
		dotR := float32(2)
		if own.PlayerID != h.LocalPlayer {
			dotClr = color.RGBA{255, 60, 60, 255}
		}
		if w.Has(id, core.CompBuilding) {
			dotR = 3
			vector.DrawFilledCircle(screen, dotX, dotY, dotR+2, color.RGBA{dotClr.R, dotClr.G, dotClr.B, 40}, false)
		}

		vector.DrawFilledCircle(screen, dotX, dotY, dotR, dotClr, false)
	}

	scanY := float32(my) + float32(mh)*float32(math.Mod(h.tick*0.3, 1.0))
	vector.DrawFilledRect(screen, float32(mx), scanY, float32(mw), 1, color.RGBA{0, 255, 0, 15}, false)
}

// ---- Ore Sparkle Drawing ----

func (h *HUD) DrawOreSparkles(screen *ebiten.Image, tileX, tileY int, oreAmount int, screenX, screenY int) {
	if oreAmount <= 0 {
		return
	}
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

// ======================== INPUT HANDLING ========================

func (h *HUD) HandleClick(mx, my int) bool {
	sidebarX := h.ScreenW - h.SidebarWidth

	// ---- Sidebar clicks ----
	if mx >= sidebarX {
		relX := mx - sidebarX
		curY := sidebarPadding

		// Credits area
		if my < curY+sidebarCreditsH {
			return true // consumed
		}
		curY += sidebarCreditsH

		// Repair/Sell/Waypoint buttons
		if my >= curY && my < curY+sidebarCmdBtnH {
			contentW := h.SidebarWidth - sidebarPowerBarW - sidebarPadding*2
			btnW := (contentW - 8) / 3
			btnRelX := relX - sidebarPowerBarW - sidebarPadding
			if btnRelX >= 0 {
				btnIdx := btnRelX / (btnW + 4)
				if btnIdx == 0 {
					// Repair
					h.RepairMode = !h.RepairMode
					h.SellMode = false
				} else if btnIdx == 1 {
					// Sell
					h.SellMode = !h.SellMode
					h.RepairMode = false
				} else if btnIdx == 2 {
					// Waypoint (toggle)
					h.RepairMode = false
					h.SellMode = false
				}
			}
			return true
		}
		curY += sidebarCmdBtnH

		// Tab clicks
		if my >= curY && my < curY+sidebarTabH {
			contentW := h.SidebarWidth - sidebarPowerBarW - sidebarPadding*2
			tabW := contentW / 3
			tabRelX := relX - sidebarPowerBarW - sidebarPadding
			if tabRelX >= 0 {
				tabIdx := tabRelX / tabW
				if tabIdx >= 0 && tabIdx < 3 {
					h.ActiveTab = BuildTab(tabIdx)
					h.ScrollOffset = 0
				}
			}
			return true
		}
		curY += sidebarTabH + 2

		// Build grid clicks are handled by GetSidebarBuildClick
		// Scroll arrows
		if my >= h.ScreenH-sidebarScrollH {
			h.ScrollOffset += 2
			return true
		}

		return true // consume all sidebar clicks
	}

	// Command buttons in bottom panel
	panelX := h.MinimapSize + 10
	panelY := h.ScreenH - h.BottomPanelH
	panelW := h.ScreenW - h.SidebarWidth - h.MinimapSize - 20
	cmdX := panelX + panelW - 250
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
		return true
	}

	return false
}

// HandleScroll handles mouse wheel for sidebar scrolling
func (h *HUD) HandleScroll(mx, my int, scrollY float64) bool {
	if mx >= h.ScreenW-h.SidebarWidth {
		if scrollY > 0 {
			h.ScrollOffset -= 2
		} else if scrollY < 0 {
			h.ScrollOffset += 2
		}
		if h.ScrollOffset < 0 {
			h.ScrollOffset = 0
		}
		return true
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

// GetSidebarBuildClick returns the build item key if a build slot was clicked
func (h *HUD) GetSidebarBuildClick(mx, my int, w *core.World) string {
	sidebarX := h.ScreenW - h.SidebarWidth
	if mx < sidebarX {
		return ""
	}

	// Calculate grid start Y
	curY := sidebarPadding + sidebarCreditsH + sidebarCmdBtnH + sidebarTabH + 2
	if my < curY {
		return ""
	}

	contentW := h.SidebarWidth - sidebarPowerBarW - sidebarPadding*2
	slotW := (contentW - sidebarSlotGap) / 2
	slotH := slotW

	relY := my - curY
	relX := mx - (sidebarX + sidebarPowerBarW + sidebarPadding)

	if relX < 0 || relX >= contentW {
		return ""
	}

	col := relX / (slotW + sidebarSlotGap)
	row := relY / (slotH + sidebarSlotGap)

	idx := row*2 + col + h.ScrollOffset
	items := h.getBuildItems(w)
	if idx >= 0 && idx < len(items) {
		return items[idx].Key
	}
	return ""
}

// GetSidebarBuildingClick returns the building key if a building button was clicked (for backwards compat)
func (h *HUD) GetSidebarBuildingClick(mx, my int, w *core.World) string {
	if h.ActiveTab != TabBuildings && h.ActiveTab != TabDefense {
		return ""
	}
	return h.GetSidebarBuildClick(mx, my, w)
}

// GetSidebarUnitClick returns the unit key if a unit button was clicked
func (h *HUD) GetSidebarUnitClick(mx, my int, w *core.World) string {
	if h.ActiveTab != TabUnits {
		return ""
	}
	return h.GetSidebarBuildClick(mx, my, w)
}

func (h *HUD) PlayerHasConYard(w *core.World) bool {
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

// ---- Helpers ----

func prereqNames(tt *systems.TechTree, prereqs []string) string {
	if len(prereqs) == 0 {
		return "none"
	}
	result := ""
	for i, p := range prereqs {
		if i > 0 {
			result += ", "
		}
		if bdef, ok := tt.Buildings[p]; ok {
			result += bdef.Name
		} else {
			result += p
		}
	}
	return result
}

// GetBuildItems is public access to build items for external use
func (h *HUD) GetBuildItems(w *core.World) []SidebarBuildItem {
	return h.getBuildItems(w)
}

// SortedBuildingKeys returns building keys in sorted order (for deterministic iteration)
func SortedBuildingKeys(m map[string]*systems.BuildingDef) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// SortedUnitKeys returns unit keys in sorted order
func SortedUnitKeys(m map[string]*systems.UnitDef) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

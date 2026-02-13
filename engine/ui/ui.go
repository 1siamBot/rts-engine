package ui

import (
	"fmt"
	"image/color"

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
)

// HUD is the main heads-up display
type HUD struct {
	ScreenW, ScreenH int
	SidebarWidth     int
	TopBarHeight     int

	// State
	CurrentCommand CommandType
	BuildQueue     []string
	SelectedIDs    []core.EntityID
	ControlGroups  [10][]core.EntityID

	// References
	TechTree *systems.TechTree
	Players  *core.PlayerManager
	LocalPlayer int
}

func NewHUD(sw, sh int, tt *systems.TechTree, pm *core.PlayerManager, localPlayer int) *HUD {
	return &HUD{
		ScreenW:      sw,
		ScreenH:      sh,
		SidebarWidth: 200,
		TopBarHeight: 30,
		TechTree:     tt,
		Players:      pm,
		LocalPlayer:  localPlayer,
	}
}

// Draw renders the entire HUD
func (h *HUD) Draw(screen *ebiten.Image, w *core.World) {
	h.drawTopBar(screen)
	h.drawSidebar(screen, w)
	h.drawUnitInfo(screen, w)
	h.drawCommandButtons(screen)
}

func (h *HUD) drawTopBar(screen *ebiten.Image) {
	// Resource bar
	vector.DrawFilledRect(screen, 0, 0, float32(h.ScreenW), float32(h.TopBarHeight), color.RGBA{0, 0, 0, 180}, false)
	player := h.Players.GetPlayer(h.LocalPlayer)
	if player == nil {
		return
	}
	info := fmt.Sprintf("Credits: $%d | Power: %d/%d", player.Credits, player.Power, player.PowerUse)
	if !player.HasPower() {
		info += " ⚠ LOW POWER"
	}
	ebitenutil.DebugPrintAt(screen, info, 10, 8)
}

func (h *HUD) drawSidebar(screen *ebiten.Image, w *core.World) {
	sx := float32(h.ScreenW - h.SidebarWidth)
	vector.DrawFilledRect(screen, sx, float32(h.TopBarHeight), float32(h.SidebarWidth), float32(h.ScreenH-h.TopBarHeight), color.RGBA{20, 20, 40, 220}, false)

	// Build buttons
	y := h.TopBarHeight + 10
	ebitenutil.DebugPrintAt(screen, "=== BUILD ===", int(sx)+10, y)
	y += 20

	// Show available buildings
	bIdx := 0
	for key, bdef := range h.TechTree.Buildings {
		if bIdx >= 8 {
			break
		}
		btnColor := color.RGBA{60, 60, 100, 255}
		vector.DrawFilledRect(screen, sx+10, float32(y), float32(h.SidebarWidth-20), 24, btnColor, false)
		vector.StrokeRect(screen, sx+10, float32(y), float32(h.SidebarWidth-20), 24, 1, color.RGBA{100, 100, 160, 255}, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s $%d", bdef.Name, bdef.Cost), int(sx)+15, y+5)
		_ = key
		y += 28
		bIdx++
	}

	y += 10
	ebitenutil.DebugPrintAt(screen, "=== UNITS ===", int(sx)+10, y)
	y += 20

	uIdx := 0
	for key, udef := range h.TechTree.Units {
		if uIdx >= 8 {
			break
		}
		btnColor := color.RGBA{60, 80, 60, 255}
		vector.DrawFilledRect(screen, sx+10, float32(y), float32(h.SidebarWidth-20), 24, btnColor, false)
		vector.StrokeRect(screen, sx+10, float32(y), float32(h.SidebarWidth-20), 24, 1, color.RGBA{100, 140, 100, 255}, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s $%d", udef.Name, udef.Cost), int(sx)+15, y+5)
		_ = key
		y += 28
		uIdx++
	}
}

func (h *HUD) drawUnitInfo(screen *ebiten.Image, w *core.World) {
	if len(h.SelectedIDs) == 0 {
		return
	}
	// Draw info panel at bottom
	panelH := 80
	py := h.ScreenH - panelH
	pw := h.ScreenW - h.SidebarWidth
	vector.DrawFilledRect(screen, 0, float32(py), float32(pw), float32(panelH), color.RGBA{0, 0, 0, 180}, false)

	x := 10
	for i, id := range h.SelectedIDs {
		if i >= 12 {
			break
		}
		hp := w.Get(id, core.CompHealth)
		if hp == nil {
			continue
		}
		h2 := hp.(*core.Health)
		// Small unit portrait
		vector.DrawFilledRect(screen, float32(x), float32(py+5), 40, 40, color.RGBA{60, 120, 255, 200}, false)
		// HP bar under portrait
		ratio := float32(h2.Ratio())
		barColor := color.RGBA{0, 200, 0, 255}
		if ratio < 0.5 {
			barColor = color.RGBA{255, 200, 0, 255}
		}
		if ratio < 0.25 {
			barColor = color.RGBA{255, 0, 0, 255}
		}
		vector.DrawFilledRect(screen, float32(x), float32(py+48), 40*ratio, 4, barColor, false)

		// Show weapon info for first selected
		if i == 0 {
			wep := w.Get(id, core.CompWeapon)
			info := fmt.Sprintf("HP: %d/%d", h2.Current, h2.Max)
			if wep != nil {
				w2 := wep.(*core.Weapon)
				info += fmt.Sprintf(" | DMG: %d | RNG: %.1f", w2.Damage, w2.Range)
			}
			ebitenutil.DebugPrintAt(screen, info, 10, py+58)
		}
		x += 45
	}
}

func (h *HUD) drawCommandButtons(screen *ebiten.Image) {
	// Command buttons at bottom-right (above sidebar)
	bx := h.ScreenW - h.SidebarWidth
	by := h.ScreenH - 80
	cmds := []struct {
		name string
		cmd  CommandType
	}{
		{"Move", CmdMove},
		{"Attack", CmdAttack},
		{"Stop", CmdStop},
		{"Guard", CmdGuard},
	}
	for i, c := range cmds {
		x := float32(bx - 220 + i*55)
		y := float32(by + 10)
		active := h.CurrentCommand == c.cmd
		clr := color.RGBA{50, 50, 80, 255}
		if active {
			clr = color.RGBA{100, 100, 200, 255}
		}
		vector.DrawFilledRect(screen, x, y, 50, 25, clr, false)
		vector.StrokeRect(screen, x, y, 50, 25, 1, color.RGBA{150, 150, 200, 255}, false)
		ebitenutil.DebugPrintAt(screen, c.name, int(x)+5, int(y)+6)
	}
}

// HandleClick processes sidebar/HUD clicks. Returns true if click was consumed.
func (h *HUD) HandleClick(mx, my int) bool {
	// Check sidebar area
	if mx >= h.ScreenW-h.SidebarWidth && my >= h.TopBarHeight {
		return true // consumed by sidebar
	}
	// Check command buttons
	bx := h.ScreenW - h.SidebarWidth
	by := h.ScreenH - 80
	cmds := []CommandType{CmdMove, CmdAttack, CmdStop, CmdGuard}
	for i, c := range cmds {
		x := bx - 220 + i*55
		y := by + 10
		if mx >= x && mx < x+50 && my >= y && my < y+25 {
			h.CurrentCommand = c
			return true
		}
	}
	return false
}

// AssignControlGroup assigns selected units to group n
func (h *HUD) AssignControlGroup(n int) {
	if n < 0 || n > 9 {
		return
	}
	h.ControlGroups[n] = make([]core.EntityID, len(h.SelectedIDs))
	copy(h.ControlGroups[n], h.SelectedIDs)
}

// RecallControlGroup selects units from group n
func (h *HUD) RecallControlGroup(n int) {
	if n < 0 || n > 9 {
		return
	}
	h.SelectedIDs = make([]core.EntityID, len(h.ControlGroups[n]))
	copy(h.SelectedIDs, h.ControlGroups[n])
}

// IsInSidebar returns true if the mouse position is over the sidebar
func (h *HUD) IsInSidebar(mx, _ int) bool {
	return mx >= h.ScreenW-h.SidebarWidth
}

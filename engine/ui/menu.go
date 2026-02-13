package ui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// GameState represents the current UI state
type GameState int

const (
	StateMainMenu GameState = iota
	StateSkirmishSetup
	StatePlaying
	StatePaused
	StateSettings
	StateGameOver
)

// SkirmishSettings holds skirmish configuration
type SkirmishSettings struct {
	MapIndex       int
	Faction        int // 0=Allied, 1=Soviet
	AIDifficulty   int // 0=Easy, 1=Medium, 2=Hard
	StartingCredits int // index into creditOptions
	MapSize        int // 0=Small, 1=Medium, 2=Large
}

// GameOverStats holds end-game statistics
type GameOverStats struct {
	Victory           bool
	UnitsBuilt        int
	UnitsLost         int
	BuildingsBuilt    int
	BuildingsDestroyed int
	CreditsEarned     int
}

// MenuButton represents a clickable menu button
type MenuButton struct {
	X, Y, W, H int
	Text        string
	Hovered     bool
	Disabled    bool
}

// MenuSystem manages all game menus
type MenuSystem struct {
	State       GameState
	PrevState   GameState // for settings back button
	ScreenW     int
	ScreenH     int
	Tick        float64
	Sprites     *UISprites

	// Skirmish
	Skirmish SkirmishSettings

	// Settings
	Settings     GameSettings
	TempSettings GameSettings // edited but not applied

	// Game Over
	GameOverData GameOverStats

	// Internal
	buttons     []MenuButton
	hoverIdx    int
	settingsTab int // 0=Graphics, 1=Audio, 2=Game, 3=Controls

	// Callbacks
	OnStartGame   func(SkirmishSettings)
	OnResumeGame  func()
	OnRestartGame func()
	OnQuitToMenu  func()
	OnExitGame    func()
	OnApplySettings func(GameSettings)
}

// GameSettings holds configurable settings
type GameSettings struct {
	Fullscreen    bool
	VSync         bool
	MusicVolume   float64 // 0-1
	SFXVolume     float64 // 0-1
	ScrollSpeed   float64 // 1-10
	ShowHealthBars bool
	ShowMinimap   bool
}

var (
	mapNames      = []string{"Riverside", "Desert Storm", "Arctic Front", "Island Fortress"}
	factionNames  = []string{"Allied", "Soviet"}
	diffNames     = []string{"Easy", "Medium", "Hard"}
	creditOptions = []int{5000, 10000, 20000}
	mapSizeNames  = []string{"Small", "Medium", "Large"}

	menuBG      = color.RGBA{8, 8, 16, 255}
	menuPanel   = color.RGBA{15, 15, 30, 230}
	menuBorder  = color.RGBA{0, 140, 200, 255}
	menuAccent  = color.RGBA{0, 200, 255, 255}
	menuBtnNorm = color.RGBA{25, 35, 55, 240}
	menuBtnHov  = color.RGBA{35, 55, 90, 255}
	menuBtnAct  = color.RGBA{0, 100, 160, 255}
	menuBtnDis  = color.RGBA{20, 20, 30, 200}
	menuText    = color.RGBA{200, 220, 255, 255}
	menuTextDim = color.RGBA{100, 120, 150, 255}
	menuGold    = color.RGBA{255, 200, 50, 255}
	menuRed     = color.RGBA{220, 50, 50, 255}
	menuGreen   = color.RGBA{50, 220, 80, 255}
)

func NewMenuSystem(screenW, screenH int, sprites *UISprites) *MenuSystem {
	return &MenuSystem{
		State:   StateMainMenu,
		ScreenW: screenW,
		ScreenH: screenH,
		Sprites: sprites,
		Skirmish: SkirmishSettings{
			MapIndex:        0,
			Faction:         0,
			AIDifficulty:    1,
			StartingCredits: 1, // 10000
			MapSize:         1, // Medium
		},
		Settings: GameSettings{
			VSync:          true,
			MusicVolume:    0.7,
			SFXVolume:      0.8,
			ScrollSpeed:    5,
			ShowHealthBars: true,
			ShowMinimap:    true,
		},
		hoverIdx: -1,
	}
}

func (m *MenuSystem) Update(dt float64) {
	m.Tick += dt
	mx, my := ebiten.CursorPosition()

	switch m.State {
	case StateMainMenu:
		m.updateMainMenu(mx, my)
	case StateSkirmishSetup:
		m.updateSkirmishSetup(mx, my)
	case StatePaused:
		m.updatePauseMenu(mx, my)
	case StateSettings:
		m.updateSettings(mx, my)
	case StateGameOver:
		m.updateGameOver(mx, my)
	}
}

func (m *MenuSystem) Draw(screen *ebiten.Image) {
	switch m.State {
	case StateMainMenu:
		m.drawMainMenu(screen)
	case StateSkirmishSetup:
		m.drawSkirmishSetup(screen)
	case StatePaused:
		m.drawPauseMenu(screen)
	case StateSettings:
		m.drawSettings(screen)
	case StateGameOver:
		m.drawGameOver(screen)
	}
}

// ==================== MAIN MENU ====================

func (m *MenuSystem) updateMainMenu(mx, my int) {
	buttons := m.mainMenuButtons()
	m.hoverIdx = -1
	for i, b := range buttons {
		if mx >= b.X && mx < b.X+b.W && my >= b.Y && my < b.Y+b.H && !b.Disabled {
			m.hoverIdx = i
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && m.hoverIdx >= 0 {
		switch m.hoverIdx {
		case 0: // SKIRMISH
			m.State = StateSkirmishSetup
		case 1: // MULTIPLAYER (placeholder)
			// no-op
		case 2: // MAP EDITOR (placeholder)
			// no-op
		case 3: // SETTINGS
			m.PrevState = StateMainMenu
			m.TempSettings = m.Settings
			m.settingsTab = 0
			m.State = StateSettings
		case 4: // EXIT
			if m.OnExitGame != nil {
				m.OnExitGame()
			}
		}
	}
}

func (m *MenuSystem) mainMenuButtons() []MenuButton {
	cx := m.ScreenW / 2
	startY := m.ScreenH/2 - 20
	bw, bh, gap := 260, 40, 8
	names := []string{"SKIRMISH", "MULTIPLAYER", "MAP EDITOR", "SETTINGS", "EXIT"}
	disabled := []bool{false, true, true, false, false}
	buttons := make([]MenuButton, len(names))
	for i, name := range names {
		buttons[i] = MenuButton{
			X: cx - bw/2, Y: startY + i*(bh+gap),
			W: bw, H: bh, Text: name, Disabled: disabled[i],
		}
	}
	return buttons
}

func (m *MenuSystem) drawMainMenu(screen *ebiten.Image) {
	screen.Fill(menuBG)
	m.drawAnimatedBG(screen)

	// Title
	m.drawTitle(screen)

	// Buttons
	buttons := m.mainMenuButtons()
	for i, b := range buttons {
		m.drawMenuButton(screen, b, i == m.hoverIdx)
	}

	// Version
	ebitenutil.DebugPrintAt(screen, "RTS Engine v0.5.0", 10, m.ScreenH-20)
}

func (m *MenuSystem) drawTitle(screen *ebiten.Image) {
	cx := m.ScreenW / 2
	title := "COMMAND & CONQUER"
	subtitle := "RTS ENGINE"

	// Title with glow effect
	titleW := len(title) * 12
	// Glow behind
	pulse := 0.7 + 0.3*math.Sin(m.Tick*2)
	glowAlpha := uint8(40 * pulse)
	drawRoundedRect(screen, float32(cx-titleW/2-20), float32(60), float32(titleW+40), 70, 8,
		color.RGBA{0, 100, 180, glowAlpha})

	// Title text (large - use debug print scaled)
	// Since we only have debug text, draw it big by repeating with offsets for "bold"
	ty := 75
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			ebitenutil.DebugPrintAt(screen, title, cx-len(title)*3+dx, ty+dy)
		}
	}
	// Cyan accent line under title
	lineY := float32(ty + 20)
	vector.DrawFilledRect(screen, float32(cx-120), lineY, 240, 2, menuAccent, false)
	// Glow
	vector.DrawFilledRect(screen, float32(cx-120), lineY-1, 240, 4, color.RGBA{0, 180, 255, 40}, false)

	// Subtitle
	ebitenutil.DebugPrintAt(screen, subtitle, cx-len(subtitle)*3, ty+30)
}

func (m *MenuSystem) drawAnimatedBG(screen *ebiten.Image) {
	// Animated grid lines
	t := m.Tick
	gridAlpha := uint8(15)
	for i := 0; i < 20; i++ {
		x := float32(math.Mod(float64(i)*70+t*20, float64(m.ScreenW)))
		vector.StrokeLine(screen, x, 0, x, float32(m.ScreenH), 1, color.RGBA{0, 80, 120, gridAlpha}, false)
	}
	for i := 0; i < 12; i++ {
		y := float32(math.Mod(float64(i)*65+t*15, float64(m.ScreenH)))
		vector.StrokeLine(screen, 0, y, float32(m.ScreenW), y, 1, color.RGBA{0, 80, 120, gridAlpha}, false)
	}

	// Floating particles
	for i := 0; i < 30; i++ {
		px := float32(math.Mod(float64(i)*43.7+t*10+float64(i*i)*0.3, float64(m.ScreenW)))
		py := float32(math.Mod(float64(i)*67.3+t*5+float64(i)*1.7, float64(m.ScreenH)))
		alpha := uint8(20 + 20*math.Sin(t*2+float64(i)))
		vector.DrawFilledCircle(screen, px, py, 1.5, color.RGBA{0, 180, 255, alpha}, false)
	}
}

// ==================== SKIRMISH SETUP ====================

func (m *MenuSystem) updateSkirmishSetup(mx, my int) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		m.State = StateMainMenu
		return
	}

	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	cx := m.ScreenW / 2
	panelX := cx - 200
	y := 130

	// Map selection arrows
	if m.clickInRect(mx, my, panelX, y+20, 30, 24) {
		m.Skirmish.MapIndex = (m.Skirmish.MapIndex - 1 + len(mapNames)) % len(mapNames)
	}
	if m.clickInRect(mx, my, panelX+370, y+20, 30, 24) {
		m.Skirmish.MapIndex = (m.Skirmish.MapIndex + 1) % len(mapNames)
	}
	y += 60

	// Faction
	if m.clickInRect(mx, my, panelX, y+20, 30, 24) {
		m.Skirmish.Faction = (m.Skirmish.Faction - 1 + len(factionNames)) % len(factionNames)
	}
	if m.clickInRect(mx, my, panelX+370, y+20, 30, 24) {
		m.Skirmish.Faction = (m.Skirmish.Faction + 1) % len(factionNames)
	}
	y += 60

	// AI Difficulty
	if m.clickInRect(mx, my, panelX, y+20, 30, 24) {
		m.Skirmish.AIDifficulty = (m.Skirmish.AIDifficulty - 1 + len(diffNames)) % len(diffNames)
	}
	if m.clickInRect(mx, my, panelX+370, y+20, 30, 24) {
		m.Skirmish.AIDifficulty = (m.Skirmish.AIDifficulty + 1) % len(diffNames)
	}
	y += 60

	// Starting Credits
	if m.clickInRect(mx, my, panelX, y+20, 30, 24) {
		m.Skirmish.StartingCredits = (m.Skirmish.StartingCredits - 1 + len(creditOptions)) % len(creditOptions)
	}
	if m.clickInRect(mx, my, panelX+370, y+20, 30, 24) {
		m.Skirmish.StartingCredits = (m.Skirmish.StartingCredits + 1) % len(creditOptions)
	}
	y += 60

	// Map Size
	if m.clickInRect(mx, my, panelX, y+20, 30, 24) {
		m.Skirmish.MapSize = (m.Skirmish.MapSize - 1 + len(mapSizeNames)) % len(mapSizeNames)
	}
	if m.clickInRect(mx, my, panelX+370, y+20, 30, 24) {
		m.Skirmish.MapSize = (m.Skirmish.MapSize + 1) % len(mapSizeNames)
	}
	y += 80

	// START GAME button
	btnW, btnH := 260, 44
	btnX := cx - btnW/2
	if m.clickInRect(mx, my, btnX, y, btnW, btnH) {
		if m.OnStartGame != nil {
			m.OnStartGame(m.Skirmish)
		}
		m.State = StatePlaying
	}

	// BACK button
	if m.clickInRect(mx, my, btnX, y+btnH+12, btnW, 36) {
		m.State = StateMainMenu
	}
}

func (m *MenuSystem) drawSkirmishSetup(screen *ebiten.Image) {
	screen.Fill(menuBG)
	m.drawAnimatedBG(screen)

	cx := m.ScreenW / 2

	// Title
	ebitenutil.DebugPrintAt(screen, "SKIRMISH SETUP", cx-len("SKIRMISH SETUP")*3, 30)
	vector.DrawFilledRect(screen, float32(cx-80), 48, 160, 2, menuAccent, false)

	// Panel background
	panelX := cx - 210
	panelW := 420
	drawRoundedRect(screen, float32(panelX-10), 70, float32(panelW+20), 440, 8, menuPanel)
	drawRoundedRectStroke(screen, float32(panelX-10), 70, float32(panelW+20), 440, 8, menuBorder)

	y := 130

	// Options with left/right arrows
	m.drawOption(screen, panelX, y, "MAP", mapNames[m.Skirmish.MapIndex])
	y += 60
	m.drawOption(screen, panelX, y, "FACTION", factionNames[m.Skirmish.Faction])
	y += 60
	m.drawOption(screen, panelX, y, "AI DIFFICULTY", diffNames[m.Skirmish.AIDifficulty])
	y += 60
	m.drawOption(screen, panelX, y, "CREDITS", fmt.Sprintf("$%d", creditOptions[m.Skirmish.StartingCredits]))
	y += 60
	m.drawOption(screen, panelX, y, "MAP SIZE", mapSizeNames[m.Skirmish.MapSize])
	y += 80

	// START GAME button
	btnW, btnH := 260, 44
	btnX := cx - btnW/2
	m.drawBigButton(screen, btnX, y, btnW, btnH, "START GAME", menuGreen)

	// BACK
	m.drawBigButton(screen, btnX, y+btnH+12, btnW, 36, "BACK", menuBtnNorm)
}

func (m *MenuSystem) drawOption(screen *ebiten.Image, x, y int, label, value string) {
	ebitenutil.DebugPrintAt(screen, label, x+40, y)

	// Value display with arrows
	valX := x + 40
	valW := 320
	drawRoundedRect(screen, float32(valX), float32(y+16), float32(valW), 28, 4, color.RGBA{20, 25, 40, 240})
	drawRoundedRectStroke(screen, float32(valX), float32(y+16), float32(valW), 28, 4, color.RGBA{40, 60, 100, 200})

	// Left arrow
	m.drawArrowButton(screen, x, y+20, false)
	// Right arrow
	m.drawArrowButton(screen, x+370, y+20, true)

	// Centered value text
	textX := valX + valW/2 - len(value)*3
	ebitenutil.DebugPrintAt(screen, value, textX, y+22)
}

func (m *MenuSystem) drawArrowButton(screen *ebiten.Image, x, y int, right bool) {
	drawRoundedRect(screen, float32(x), float32(y), 30, 24, 4, menuBtnHov)
	arrow := "<"
	if right {
		arrow = ">"
	}
	ebitenutil.DebugPrintAt(screen, arrow, x+12, y+6)
}

// ==================== PAUSE MENU ====================

func (m *MenuSystem) updatePauseMenu(mx, my int) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		m.State = StatePlaying
		if m.OnResumeGame != nil {
			m.OnResumeGame()
		}
		return
	}

	buttons := m.pauseMenuButtons()
	m.hoverIdx = -1
	for i, b := range buttons {
		if mx >= b.X && mx < b.X+b.W && my >= b.Y && my < b.Y+b.H {
			m.hoverIdx = i
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && m.hoverIdx >= 0 {
		switch m.hoverIdx {
		case 0: // RESUME
			m.State = StatePlaying
			if m.OnResumeGame != nil {
				m.OnResumeGame()
			}
		case 1: // SETTINGS
			m.PrevState = StatePaused
			m.TempSettings = m.Settings
			m.settingsTab = 0
			m.State = StateSettings
		case 2: // RESTART
			if m.OnRestartGame != nil {
				m.OnRestartGame()
			}
		case 3: // SURRENDER
			m.GameOverData = GameOverStats{Victory: false}
			m.State = StateGameOver
		case 4: // QUIT TO MENU
			m.State = StateMainMenu
			if m.OnQuitToMenu != nil {
				m.OnQuitToMenu()
			}
		}
	}
}

func (m *MenuSystem) pauseMenuButtons() []MenuButton {
	cx := m.ScreenW / 2
	startY := m.ScreenH/2 - 80
	bw, bh, gap := 220, 36, 8
	names := []string{"RESUME", "SETTINGS", "RESTART", "SURRENDER", "QUIT TO MENU"}
	buttons := make([]MenuButton, len(names))
	for i, name := range names {
		buttons[i] = MenuButton{
			X: cx - bw/2, Y: startY + i*(bh+gap),
			W: bw, H: bh, Text: name,
		}
	}
	return buttons
}

func (m *MenuSystem) drawPauseMenu(screen *ebiten.Image) {
	// Dark overlay
	vector.DrawFilledRect(screen, 0, 0, float32(m.ScreenW), float32(m.ScreenH), color.RGBA{0, 0, 0, 160}, false)

	cx := m.ScreenW / 2

	// Panel
	panelW, panelH := 300, 320
	px := float32(cx - panelW/2)
	py := float32(m.ScreenH/2 - panelH/2 - 20)
	drawRoundedRect(screen, px, py, float32(panelW), float32(panelH), 10, menuPanel)
	drawRoundedRectStroke(screen, px, py, float32(panelW), float32(panelH), 10, menuBorder)

	// Title
	title := "PAUSED"
	ebitenutil.DebugPrintAt(screen, title, cx-len(title)*3, int(py)+20)
	vector.DrawFilledRect(screen, px+20, py+38, float32(panelW-40), 2, menuAccent, false)

	// Buttons
	buttons := m.pauseMenuButtons()
	for i, b := range buttons {
		m.drawMenuButton(screen, b, i == m.hoverIdx)
	}
}

// ==================== SETTINGS ====================

func (m *MenuSystem) updateSettings(mx, my int) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		m.State = m.PrevState
		return
	}

	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	cx := m.ScreenW / 2

	// Tab buttons
	tabNames := []string{"Graphics", "Audio", "Game", "Controls"}
	tabW := 90
	tabStartX := cx - (len(tabNames)*tabW)/2
	for i := range tabNames {
		if m.clickInRect(mx, my, tabStartX+i*tabW, 90, tabW-4, 28) {
			m.settingsTab = i
		}
	}

	panelX := cx - 200
	y := 160

	switch m.settingsTab {
	case 0: // Graphics
		if m.clickInRect(mx, my, panelX+250, y, 100, 24) {
			m.TempSettings.Fullscreen = !m.TempSettings.Fullscreen
		}
		y += 50
		if m.clickInRect(mx, my, panelX+250, y, 100, 24) {
			m.TempSettings.VSync = !m.TempSettings.VSync
		}
	case 1: // Audio
		// Music slider area
		if mx >= panelX+150 && mx < panelX+380 && my >= y && my < y+24 {
			m.TempSettings.MusicVolume = float64(mx-panelX-150) / 230.0
			if m.TempSettings.MusicVolume < 0 { m.TempSettings.MusicVolume = 0 }
			if m.TempSettings.MusicVolume > 1 { m.TempSettings.MusicVolume = 1 }
		}
		y += 50
		if mx >= panelX+150 && mx < panelX+380 && my >= y && my < y+24 {
			m.TempSettings.SFXVolume = float64(mx-panelX-150) / 230.0
			if m.TempSettings.SFXVolume < 0 { m.TempSettings.SFXVolume = 0 }
			if m.TempSettings.SFXVolume > 1 { m.TempSettings.SFXVolume = 1 }
		}
	case 2: // Game
		// Scroll speed slider
		if mx >= panelX+150 && mx < panelX+380 && my >= y && my < y+24 {
			m.TempSettings.ScrollSpeed = float64(mx-panelX-150) / 230.0 * 10
			if m.TempSettings.ScrollSpeed < 1 { m.TempSettings.ScrollSpeed = 1 }
			if m.TempSettings.ScrollSpeed > 10 { m.TempSettings.ScrollSpeed = 10 }
		}
		y += 50
		if m.clickInRect(mx, my, panelX+250, y, 100, 24) {
			m.TempSettings.ShowHealthBars = !m.TempSettings.ShowHealthBars
		}
		y += 50
		if m.clickInRect(mx, my, panelX+250, y, 100, 24) {
			m.TempSettings.ShowMinimap = !m.TempSettings.ShowMinimap
		}
	}

	// APPLY / BACK buttons
	btnY := 420
	if m.clickInRect(mx, my, cx-130, btnY, 120, 36) {
		// APPLY
		m.Settings = m.TempSettings
		if m.OnApplySettings != nil {
			m.OnApplySettings(m.Settings)
		}
		m.State = m.PrevState
	}
	if m.clickInRect(mx, my, cx+10, btnY, 120, 36) {
		// BACK
		m.State = m.PrevState
	}
}

func (m *MenuSystem) drawSettings(screen *ebiten.Image) {
	if m.PrevState == StatePaused || m.PrevState == StatePlaying {
		vector.DrawFilledRect(screen, 0, 0, float32(m.ScreenW), float32(m.ScreenH), color.RGBA{0, 0, 0, 180}, false)
	} else {
		screen.Fill(menuBG)
		m.drawAnimatedBG(screen)
	}

	cx := m.ScreenW / 2

	// Panel
	panelW, panelH := 440, 380
	px := float32(cx - panelW/2)
	py := float32(50)
	drawRoundedRect(screen, px, py, float32(panelW), float32(panelH), 10, menuPanel)
	drawRoundedRectStroke(screen, px, py, float32(panelW), float32(panelH), 10, menuBorder)

	// Title
	ebitenutil.DebugPrintAt(screen, "SETTINGS", cx-len("SETTINGS")*3, 60)
	vector.DrawFilledRect(screen, px+20, 78, float32(panelW-40), 2, menuAccent, false)

	// Tabs
	tabNames := []string{"Graphics", "Audio", "Game", "Controls"}
	tabW := 90
	tabStartX := cx - (len(tabNames)*tabW)/2
	for i, name := range tabNames {
		tx := tabStartX + i*tabW
		clr := menuBtnNorm
		if i == m.settingsTab {
			clr = menuBtnAct
		}
		drawRoundedRect(screen, float32(tx), 90, float32(tabW-4), 28, 4, clr)
		ebitenutil.DebugPrintAt(screen, name, tx+(tabW-4)/2-len(name)*3, 97)
	}

	panelX := cx - 200
	y := 160

	switch m.settingsTab {
	case 0: // Graphics
		ebitenutil.DebugPrintAt(screen, "Fullscreen", panelX+20, y+4)
		m.drawToggle(screen, panelX+250, y, m.TempSettings.Fullscreen)
		y += 50
		ebitenutil.DebugPrintAt(screen, "VSync", panelX+20, y+4)
		m.drawToggle(screen, panelX+250, y, m.TempSettings.VSync)
	case 1: // Audio
		ebitenutil.DebugPrintAt(screen, "Music Volume", panelX+20, y+4)
		m.drawSlider(screen, panelX+150, y, 230, m.TempSettings.MusicVolume)
		y += 50
		ebitenutil.DebugPrintAt(screen, "SFX Volume", panelX+20, y+4)
		m.drawSlider(screen, panelX+150, y, 230, m.TempSettings.SFXVolume)
	case 2: // Game
		ebitenutil.DebugPrintAt(screen, "Scroll Speed", panelX+20, y+4)
		m.drawSlider(screen, panelX+150, y, 230, m.TempSettings.ScrollSpeed/10)
		y += 50
		ebitenutil.DebugPrintAt(screen, "Health Bars", panelX+20, y+4)
		m.drawToggle(screen, panelX+250, y, m.TempSettings.ShowHealthBars)
		y += 50
		ebitenutil.DebugPrintAt(screen, "Show Minimap", panelX+20, y+4)
		m.drawToggle(screen, panelX+250, y, m.TempSettings.ShowMinimap)
	case 3: // Controls
		keys := []string{
			"W/A/S/D  — Camera Pan",
			"Mouse Wheel — Zoom",
			"Left Click — Select / Place",
			"Right Click — Move / Cancel",
			"ESC — Menu / Cancel",
			"G — Toggle Grid",
			"M — Toggle Minimap",
			"H — Deploy MCV",
			"DEL — Sell Building",
			"Q — Quick Train Infantry",
			"Ctrl+0-9 — Set Group",
			"0-9 — Recall Group",
		}
		for i, k := range keys {
			ebitenutil.DebugPrintAt(screen, k, panelX+30, y+i*18)
		}
	}

	// APPLY / BACK
	btnY := 420
	m.drawBigButton(screen, cx-130, btnY, 120, 36, "APPLY", menuGreen)
	m.drawBigButton(screen, cx+10, btnY, 120, 36, "BACK", menuBtnNorm)
}

// ==================== GAME OVER ====================

func (m *MenuSystem) updateGameOver(mx, my int) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	cx := m.ScreenW / 2
	btnW, btnH := 200, 40
	btnY := m.ScreenH/2 + 120

	if m.clickInRect(mx, my, cx-btnW-10, btnY, btnW, btnH) {
		// PLAY AGAIN
		if m.OnRestartGame != nil {
			m.OnRestartGame()
		}
		m.State = StatePlaying
	}
	if m.clickInRect(mx, my, cx+10, btnY, btnW, btnH) {
		// MAIN MENU
		m.State = StateMainMenu
		if m.OnQuitToMenu != nil {
			m.OnQuitToMenu()
		}
	}
}

func (m *MenuSystem) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, float32(m.ScreenW), float32(m.ScreenH), color.RGBA{0, 0, 0, 180}, false)

	cx := m.ScreenW / 2
	cy := m.ScreenH / 2

	// Panel
	panelW, panelH := 400, 340
	px := float32(cx - panelW/2)
	py := float32(cy - panelH/2)
	drawRoundedRect(screen, px, py, float32(panelW), float32(panelH), 12, menuPanel)
	drawRoundedRectStroke(screen, px, py, float32(panelW), float32(panelH), 12, menuBorder)

	// VICTORY / DEFEAT
	var resultText string
	var resultClr color.RGBA
	if m.GameOverData.Victory {
		resultText = "VICTORY"
		resultClr = menuGreen
	} else {
		resultText = "DEFEAT"
		resultClr = menuRed
	}

	// Big text with glow
	tx := cx - len(resultText)*3
	ty := int(py) + 30
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			ebitenutil.DebugPrintAt(screen, resultText, tx+dx, ty+dy)
		}
	}
	// Color underline
	vector.DrawFilledRect(screen, float32(cx-60), float32(ty+18), 120, 3, resultClr, false)

	// Stats
	stats := m.GameOverData
	sy := ty + 40
	statLines := []string{
		fmt.Sprintf("Units Built:         %d", stats.UnitsBuilt),
		fmt.Sprintf("Units Lost:          %d", stats.UnitsLost),
		fmt.Sprintf("Buildings Built:     %d", stats.BuildingsBuilt),
		fmt.Sprintf("Buildings Destroyed: %d", stats.BuildingsDestroyed),
		fmt.Sprintf("Credits Earned:      $%d", stats.CreditsEarned),
	}
	for i, line := range statLines {
		ebitenutil.DebugPrintAt(screen, line, cx-len(line)*3, sy+i*22)
	}

	// Buttons
	btnW, btnH := 200, 40
	btnY := cy + 120
	m.drawBigButton(screen, cx-btnW-10, btnY, btnW, btnH, "PLAY AGAIN", menuGreen)
	m.drawBigButton(screen, cx+10, btnY, btnW, btnH, "MAIN MENU", menuBtnNorm)
}

// ==================== DRAWING HELPERS ====================

func (m *MenuSystem) drawMenuButton(screen *ebiten.Image, b MenuButton, hovered bool) {
	clr := menuBtnNorm
	if b.Disabled {
		clr = menuBtnDis
	} else if hovered {
		clr = menuBtnHov
	}
	drawRoundedRect(screen, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), 6, clr)

	borderClr := color.RGBA{40, 70, 120, 200}
	if hovered && !b.Disabled {
		borderClr = menuAccent
	}
	drawRoundedRectStroke(screen, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), 6, borderClr)

	textClr := menuText
	if b.Disabled {
		textClr = menuTextDim
	}
	_ = textClr // debug print is always white; we use it directly
	tx := b.X + b.W/2 - len(b.Text)*3
	ty := b.Y + b.H/2 - 6
	ebitenutil.DebugPrintAt(screen, b.Text, tx, ty)
}

func (m *MenuSystem) drawBigButton(screen *ebiten.Image, x, y, w, h int, text string, clr color.RGBA) {
	mx, my := ebiten.CursorPosition()
	hovered := mx >= x && mx < x+w && my >= y && my < y+h

	bgClr := clr
	if hovered {
		// Lighten
		bgClr.R = uint8(min(int(bgClr.R)+30, 255))
		bgClr.G = uint8(min(int(bgClr.G)+30, 255))
		bgClr.B = uint8(min(int(bgClr.B)+30, 255))
	}
	drawRoundedRect(screen, float32(x), float32(y), float32(w), float32(h), 6, bgClr)

	borderClr := color.RGBA{60, 100, 160, 200}
	if hovered {
		borderClr = menuAccent
	}
	drawRoundedRectStroke(screen, float32(x), float32(y), float32(w), float32(h), 6, borderClr)

	tx := x + w/2 - len(text)*3
	ty := y + h/2 - 6
	ebitenutil.DebugPrintAt(screen, text, tx, ty)
}

func (m *MenuSystem) drawToggle(screen *ebiten.Image, x, y int, on bool) {
	w, h := 50, 24
	drawRoundedRect(screen, float32(x), float32(y), float32(w), float32(h), float32(h/2), color.RGBA{30, 35, 50, 240})

	knobX := x + 5
	clr := color.RGBA{100, 100, 120, 255}
	if on {
		knobX = x + w - 19
		clr = menuAccent
	}
	vector.DrawFilledCircle(screen, float32(knobX+7), float32(y+h/2), 8, clr, false)

	label := "OFF"
	if on {
		label = "ON"
	}
	ebitenutil.DebugPrintAt(screen, label, x+w+8, y+5)
}

func (m *MenuSystem) drawSlider(screen *ebiten.Image, x, y, w int, value float64) {
	h := 24
	// Track
	trackY := y + h/2 - 2
	drawRoundedRect(screen, float32(x), float32(trackY), float32(w), 4, 2, color.RGBA{30, 35, 50, 240})

	// Fill
	fillW := int(float64(w) * value)
	drawRoundedRect(screen, float32(x), float32(trackY), float32(fillW), 4, 2, menuAccent)

	// Knob
	knobX := float32(x) + float32(fillW)
	vector.DrawFilledCircle(screen, knobX, float32(y+h/2), 8, menuAccent, false)
	vector.StrokeCircle(screen, knobX, float32(y+h/2), 8, 1.5, color.RGBA{255, 255, 255, 100}, false)

	// Value text
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d%%", int(value*100)), x+w+10, y+5)
}

func (m *MenuSystem) clickInRect(mx, my, x, y, w, h int) bool {
	return mx >= x && mx < x+w && my >= y && my < y+h
}

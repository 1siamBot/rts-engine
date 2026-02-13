package ui

import (
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
)

// UISprites holds all loaded UI sprite images
type UISprites struct {
	// Buttons
	BtnNormal   *ebiten.Image
	BtnHover    *ebiten.Image
	BtnActive   *ebiten.Image
	BtnDisabled *ebiten.Image

	// Rectangle buttons (for sidebar)
	BtnRectNormal *ebiten.Image
	BtnRectHover  *ebiten.Image
	BtnRectActive *ebiten.Image

	// Bars
	BarHealth    *ebiten.Image
	BarHealthLow *ebiten.Image
	BarPower     *ebiten.Image
	BarProgress  *ebiten.Image
	BarBG        *ebiten.Image

	// Icons
	IconAttack  *ebiten.Image
	IconMove    *ebiten.Image
	IconStop    *ebiten.Image
	IconGuard   *ebiten.Image
	IconRally   *ebiten.Image
	IconDeploy  *ebiten.Image
	IconPower   *ebiten.Image
	IconSell    *ebiten.Image
	IconRepair  *ebiten.Image
	IconCredits *ebiten.Image
	Crosshair   *ebiten.Image

	// RA2-style build icons (per building key)
	BuildIcons map[string]*ebiten.Image

	// RA2-style panel textures
	RA2SidebarBG   *ebiten.Image
	RA2TopBarBG    *ebiten.Image
	RA2BottomPanel *ebiten.Image
	RA2MinimapFrame *ebiten.Image
	RA2TabActive   *ebiten.Image
	RA2TabInactive *ebiten.Image
	RA2BuildSlotNormal   *ebiten.Image
	RA2BuildSlotHover    *ebiten.Image
	RA2BuildSlotDisabled *ebiten.Image
	RA2BuildSlotActive   *ebiten.Image
	RA2DarkSteelTile     *ebiten.Image
	RA2PanelDivider      *ebiten.Image

	// Generated panel textures
	PanelDark     *ebiten.Image // dark metallic panel (tileable)
	PanelMetal    *ebiten.Image // brushed metal panel
	PanelFrame    *ebiten.Image // border frame piece
	RivetImg      *ebiten.Image // single rivet
	GlowLine      *ebiten.Image // cyan glow line (horizontal)

	// Cached composited panels
	SidebarPanel  *ebiten.Image
	TopBarPanel   *ebiten.Image
	BottomPanel   *ebiten.Image
	MinimapFrame  *ebiten.Image
}

func NewUISprites() *UISprites {
	us := &UISprites{
		BuildIcons: make(map[string]*ebiten.Image),
	}
	uiDir := getUIAssetsDir()
	log.Printf("Loading UI assets from: %s", uiDir)

	// Load RA2-style UI assets first (priority)
	us.loadRA2UIAssets()

	// Load button sprites
	us.BtnNormal = loadUI(filepath.Join(uiDir, "buttons", "btn_normal.png"))
	us.BtnHover = loadUI(filepath.Join(uiDir, "buttons", "btn_hover.png"))
	us.BtnActive = loadUI(filepath.Join(uiDir, "buttons", "btn_active.png"))
	us.BtnDisabled = loadUI(filepath.Join(uiDir, "buttons", "btn_disabled.png"))
	us.BtnRectNormal = loadUI(filepath.Join(uiDir, "buttons", "btn_rect_normal.png"))
	us.BtnRectHover = loadUI(filepath.Join(uiDir, "buttons", "btn_rect_hover.png"))
	us.BtnRectActive = loadUI(filepath.Join(uiDir, "buttons", "btn_rect_active.png"))

	// Load bars
	us.BarHealth = loadUI(filepath.Join(uiDir, "bars", "bar_health.png"))
	us.BarHealthLow = loadUI(filepath.Join(uiDir, "bars", "bar_health_low.png"))
	us.BarPower = loadUI(filepath.Join(uiDir, "bars", "bar_power.png"))
	us.BarProgress = loadUI(filepath.Join(uiDir, "bars", "bar_progress.png"))
	us.BarBG = loadUI(filepath.Join(uiDir, "bars", "bar_bg.png"))

	// Load icons
	us.IconAttack = loadUI(filepath.Join(uiDir, "icons", "icon_attack.png"))
	us.IconMove = loadUI(filepath.Join(uiDir, "icons", "icon_move.png"))
	us.IconStop = loadUI(filepath.Join(uiDir, "icons", "icon_stop.png"))
	us.IconGuard = loadUI(filepath.Join(uiDir, "icons", "icon_guard.png"))
	us.IconRally = loadUI(filepath.Join(uiDir, "icons", "icon_rally.png"))
	us.IconDeploy = loadUI(filepath.Join(uiDir, "icons", "icon_deploy.png"))
	us.IconPower = loadUI(filepath.Join(uiDir, "icons", "icon_power.png"))
	us.IconSell = loadUI(filepath.Join(uiDir, "icons", "icon_sell.png"))
	us.Crosshair = loadUI(filepath.Join(uiDir, "icons", "crosshair.png"))

	// Generate procedural metallic textures
	us.generateMetallicTextures()

	loaded := 0
	for _, img := range []*ebiten.Image{
		us.BtnNormal, us.BtnHover, us.BtnActive, us.BtnDisabled,
		us.BtnRectNormal, us.BtnRectHover, us.BtnRectActive,
		us.BarHealth, us.BarHealthLow, us.BarPower, us.BarProgress, us.BarBG,
		us.IconAttack, us.IconMove, us.IconStop, us.IconGuard,
		us.IconRally, us.IconDeploy, us.IconPower, us.IconSell,
	} {
		if img != nil {
			loaded++
		}
	}
	log.Printf("UISprites: loaded %d sprite files + procedural textures", loaded)

	return us
}

func (us *UISprites) loadRA2UIAssets() {
	// Find assets/ra2/ui directory
	candidates := []string{
		"assets/ra2/ui",
		"../assets/ra2/ui",
		"../../assets/ra2/ui",
	}
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		srcDir := filepath.Dir(filename)
		candidates = append(candidates,
			filepath.Join(srcDir, "../../assets/ra2/ui"),
			filepath.Join(srcDir, "../../../assets/ra2/ui"),
		)
	}

	var ra2Dir string
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			ra2Dir = c
			break
		}
	}
	if ra2Dir == "" {
		return
	}

	log.Printf("Loading RA2 UI assets from: %s", ra2Dir)
	loaded := 0

	// Panel textures
	if img := loadUI(filepath.Join(ra2Dir, "panels", "sidebar_bg.png")); img != nil { us.RA2SidebarBG = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "panels", "topbar_bg.png")); img != nil { us.RA2TopBarBG = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "panels", "bottom_panel.png")); img != nil { us.RA2BottomPanel = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "panels", "minimap_frame.png")); img != nil { us.RA2MinimapFrame = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "panels", "dark_steel_tile.png")); img != nil { us.RA2DarkSteelTile = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "panels", "panel_divider.png")); img != nil { us.RA2PanelDivider = img; loaded++ }

	// Tab buttons
	if img := loadUI(filepath.Join(ra2Dir, "sidebar", "tab_active.png")); img != nil { us.RA2TabActive = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "sidebar", "tab_inactive.png")); img != nil { us.RA2TabInactive = img; loaded++ }

	// Build slot buttons
	if img := loadUI(filepath.Join(ra2Dir, "sidebar", "build_slot_normal.png")); img != nil { us.RA2BuildSlotNormal = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "sidebar", "build_slot_hover.png")); img != nil { us.RA2BuildSlotHover = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "sidebar", "build_slot_disabled.png")); img != nil { us.RA2BuildSlotDisabled = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "sidebar", "build_slot_active.png")); img != nil { us.RA2BuildSlotActive = img; loaded++ }

	// Command icons (override existing)
	if img := loadUI(filepath.Join(ra2Dir, "icons", "cmd_move.png")); img != nil { us.IconMove = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "icons", "cmd_attack.png")); img != nil { us.IconAttack = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "icons", "cmd_stop.png")); img != nil { us.IconStop = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "icons", "cmd_guard.png")); img != nil { us.IconGuard = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "icons", "cmd_deploy.png")); img != nil { us.IconDeploy = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "icons", "cmd_sell.png")); img != nil { us.IconSell = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "icons", "cmd_repair.png")); img != nil { us.IconRepair = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "icons", "cmd_rally.png")); img != nil { us.IconRally = img; loaded++ }

	// Resource icons
	if img := loadUI(filepath.Join(ra2Dir, "icons", "credits.png")); img != nil { us.IconCredits = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "icons", "power.png")); img != nil { us.IconPower = img; loaded++ }

	// Build icons
	buildingKeys := []string{"construction_yard", "power_plant", "barracks", "refinery", "war_factory"}
	for _, key := range buildingKeys {
		if img := loadUI(filepath.Join(ra2Dir, "icons", "build_"+key+".png")); img != nil {
			us.BuildIcons[key] = img
			loaded++
		}
	}

	// Bar textures (override existing)
	if img := loadUI(filepath.Join(ra2Dir, "bars", "health_green.png")); img != nil { us.BarHealth = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "bars", "health_red.png")); img != nil { us.BarHealthLow = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "bars", "power_bar.png")); img != nil { us.BarPower = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "bars", "progress_bar.png")); img != nil { us.BarProgress = img; loaded++ }
	if img := loadUI(filepath.Join(ra2Dir, "bars", "bar_bg.png")); img != nil { us.BarBG = img; loaded++ }

	log.Printf("RA2 UI: loaded %d assets", loaded)
}

func (us *UISprites) generateMetallicTextures() {
	// Dark metallic panel texture (256x256, tileable)
	us.PanelDark = generateDarkMetalPanel(256, 256)

	// Brushed metal panel (256x64)
	us.PanelMetal = generateBrushedMetal(256, 64)

	// Rivet (8x8)
	us.RivetImg = generateRivet(8)

	// Cyan glow line (256x4)
	us.GlowLine = generateGlowLine(256, 4, color.NRGBA{0, 180, 255, 255})

	// Panel frame border piece (16x16)
	us.PanelFrame = generateFramePiece(16)
}

// GenerateSidebarPanel creates a cached sidebar panel image
func (us *UISprites) GenerateSidebarPanel(w, h int) *ebiten.Image {
	if us.SidebarPanel != nil {
		sw := us.SidebarPanel.Bounds().Dx()
		sh := us.SidebarPanel.Bounds().Dy()
		if sw == w && sh == h {
			return us.SidebarPanel
		}
	}
	// Use RA2 sidebar texture if available (scale to fit)
	if us.RA2SidebarBG != nil {
		panel := ebiten.NewImage(w, h)
		op := &ebiten.DrawImageOptions{}
		srcW := us.RA2SidebarBG.Bounds().Dx()
		srcH := us.RA2SidebarBG.Bounds().Dy()
		op.GeoM.Scale(float64(w)/float64(srcW), float64(h)/float64(srcH))
		panel.DrawImage(us.RA2SidebarBG, op)
		us.SidebarPanel = panel
		return us.SidebarPanel
	}
	us.SidebarPanel = us.compositePanel(w, h)
	return us.SidebarPanel
}

// GenerateTopBarPanel creates a cached top bar panel image
func (us *UISprites) GenerateTopBarPanel(w, h int) *ebiten.Image {
	if us.TopBarPanel != nil {
		sw := us.TopBarPanel.Bounds().Dx()
		sh := us.TopBarPanel.Bounds().Dy()
		if sw == w && sh == h {
			return us.TopBarPanel
		}
	}
	if us.RA2TopBarBG != nil {
		panel := ebiten.NewImage(w, h)
		op := &ebiten.DrawImageOptions{}
		srcW := us.RA2TopBarBG.Bounds().Dx()
		srcH := us.RA2TopBarBG.Bounds().Dy()
		op.GeoM.Scale(float64(w)/float64(srcW), float64(h)/float64(srcH))
		panel.DrawImage(us.RA2TopBarBG, op)
		us.TopBarPanel = panel
		return us.TopBarPanel
	}
	us.TopBarPanel = us.compositePanel(w, h)
	return us.TopBarPanel
}

// GenerateBottomPanel creates a cached bottom panel image
func (us *UISprites) GenerateBottomPanel(w, h int) *ebiten.Image {
	if us.BottomPanel != nil {
		sw := us.BottomPanel.Bounds().Dx()
		sh := us.BottomPanel.Bounds().Dy()
		if sw == w && sh == h {
			return us.BottomPanel
		}
	}
	if us.RA2BottomPanel != nil {
		panel := ebiten.NewImage(w, h)
		op := &ebiten.DrawImageOptions{}
		srcW := us.RA2BottomPanel.Bounds().Dx()
		srcH := us.RA2BottomPanel.Bounds().Dy()
		op.GeoM.Scale(float64(w)/float64(srcW), float64(h)/float64(srcH))
		panel.DrawImage(us.RA2BottomPanel, op)
		us.BottomPanel = panel
		return us.BottomPanel
	}
	us.BottomPanel = us.compositePanel(w, h)
	return us.BottomPanel
}

// GenerateMinimapFrame creates a cached minimap frame
func (us *UISprites) GenerateMinimapFrame(size int) *ebiten.Image {
	if us.MinimapFrame != nil {
		sw := us.MinimapFrame.Bounds().Dx()
		if sw == size+8 {
			return us.MinimapFrame
		}
	}
	if us.RA2MinimapFrame != nil {
		frameW := size + 8
		frameH := size + 24
		panel := ebiten.NewImage(frameW, frameH)
		op := &ebiten.DrawImageOptions{}
		srcW := us.RA2MinimapFrame.Bounds().Dx()
		srcH := us.RA2MinimapFrame.Bounds().Dy()
		op.GeoM.Scale(float64(frameW)/float64(srcW), float64(frameH)/float64(srcH))
		panel.DrawImage(us.RA2MinimapFrame, op)
		us.MinimapFrame = panel
		return us.MinimapFrame
	}
	frameW := size + 8
	frameH := size + 24
	frame := ebiten.NewImage(frameW, frameH)

	// Dark background
	tilePanel(frame, us.PanelDark, frameW, frameH)

	// Inner cut-out for minimap (slightly inset)
	// Draw beveled border
	drawBevelRect(frame, 3, 19, size+2, size+2, color.NRGBA{60, 80, 100, 255}, color.NRGBA{20, 30, 40, 255})

	// Top label area glow
	if us.GlowLine != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(frameW)/float64(us.GlowLine.Bounds().Dx()), 1)
		op.GeoM.Translate(0, 16)
		frame.DrawImage(us.GlowLine, op)
	}

	// Corner rivets
	if us.RivetImg != nil {
		drawRivetAt(frame, us.RivetImg, 2, 2)
		drawRivetAt(frame, us.RivetImg, frameW-10, 2)
		drawRivetAt(frame, us.RivetImg, 2, frameH-10)
		drawRivetAt(frame, us.RivetImg, frameW-10, frameH-10)
	}

	us.MinimapFrame = frame
	return us.MinimapFrame
}

func (us *UISprites) compositePanel(w, h int) *ebiten.Image {
	panel := ebiten.NewImage(w, h)

	// Tile dark metal texture
	tilePanel(panel, us.PanelDark, w, h)

	// Top highlight line
	if us.GlowLine != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(w)/float64(us.GlowLine.Bounds().Dx()), 1)
		panel.DrawImage(us.GlowLine, op)
	}

	// Bottom subtle line
	if us.GlowLine != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(w)/float64(us.GlowLine.Bounds().Dx()), 0.5)
		op.GeoM.Translate(0, float64(h-2))
		op.ColorScale.Scale(0.5, 0.5, 0.5, 0.5)
		panel.DrawImage(us.GlowLine, op)
	}

	// Corner rivets
	if us.RivetImg != nil {
		drawRivetAt(panel, us.RivetImg, 4, 6)
		drawRivetAt(panel, us.RivetImg, w-12, 6)
		drawRivetAt(panel, us.RivetImg, 4, h-12)
		drawRivetAt(panel, us.RivetImg, w-12, h-12)
	}

	// Beveled outer edge
	drawBevelBorder(panel, w, h)

	return panel
}

// DrawButton draws a button with the appropriate state sprite
func (us *UISprites) DrawButton(screen *ebiten.Image, x, y, w, h int, state string) {
	var btn *ebiten.Image
	switch state {
	case "hover":
		btn = us.BtnNormal // Will use BtnHover when square
	case "active":
		btn = us.BtnActive
	case "disabled":
		btn = us.BtnDisabled
	default:
		btn = us.BtnNormal
	}
	if btn == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/float64(btn.Bounds().Dx()), float64(h)/float64(btn.Bounds().Dy()))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)
}

// DrawRectButton draws a rectangular button with state
func (us *UISprites) DrawRectButton(screen *ebiten.Image, x, y, w, h int, state string) {
	// Try RA2 build slot textures first
	var ra2Btn *ebiten.Image
	switch state {
	case "hover":
		ra2Btn = us.RA2BuildSlotHover
	case "active":
		ra2Btn = us.RA2BuildSlotActive
	case "disabled":
		ra2Btn = us.RA2BuildSlotDisabled
	default:
		ra2Btn = us.RA2BuildSlotNormal
	}
	if ra2Btn != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(w)/float64(ra2Btn.Bounds().Dx()), float64(h)/float64(ra2Btn.Bounds().Dy()))
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(ra2Btn, op)
		return
	}

	var btn *ebiten.Image
	switch state {
	case "hover":
		btn = us.BtnRectHover
	case "active":
		btn = us.BtnRectActive
	case "disabled":
		btn = us.BtnRectNormal
	default:
		btn = us.BtnRectNormal
	}
	if btn == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/float64(btn.Bounds().Dx()), float64(h)/float64(btn.Bounds().Dy()))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)
}

// DrawBar draws a bar (health/power/progress) with fill ratio
func (us *UISprites) DrawBar(screen *ebiten.Image, x, y, w, h int, ratio float64, barType string) {
	// Background
	if us.BarBG != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(w)/float64(us.BarBG.Bounds().Dx()), float64(h)/float64(us.BarBG.Bounds().Dy()))
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(us.BarBG, op)
	}

	// Fill
	var fill *ebiten.Image
	switch barType {
	case "health":
		if ratio > 0.5 {
			fill = us.BarHealth
		} else {
			fill = us.BarHealthLow
		}
	case "power":
		fill = us.BarPower
	case "progress":
		fill = us.BarProgress
	default:
		fill = us.BarHealth
	}

	if fill != nil && ratio > 0 {
		fillW := int(float64(w) * ratio)
		if fillW < 1 {
			fillW = 1
		}

		// Create a sub-image of the fill bar
		op := &ebiten.DrawImageOptions{}
		srcW := fill.Bounds().Dx()
		srcH := fill.Bounds().Dy()
		subW := int(float64(srcW) * ratio)
		if subW < 1 {
			subW = 1
		}
		sub := fill.SubImage(image.Rect(0, 0, subW, srcH)).(*ebiten.Image)
		op.GeoM.Scale(float64(fillW)/float64(subW), float64(h)/float64(srcH))
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(sub, op)
	}
}

// DrawIcon draws a command icon centered at position
func (us *UISprites) DrawIcon(screen *ebiten.Image, icon *ebiten.Image, x, y, size int) {
	if icon == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	iw := icon.Bounds().Dx()
	ih := icon.Bounds().Dy()
	scale := float64(size) / float64(iw)
	if ih > iw {
		scale = float64(size) / float64(ih)
	}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x)-float64(iw)*scale/2, float64(y)-float64(ih)*scale/2)
	screen.DrawImage(icon, op)
}

// GetBuildIcon returns the build icon for a building key
func (us *UISprites) GetBuildIcon(key string) *ebiten.Image {
	return us.BuildIcons[key]
}

// DrawTabButton draws a tab button using RA2 textures if available
func (us *UISprites) DrawTabButton(screen *ebiten.Image, x, y, w, h int, active bool) {
	var tab *ebiten.Image
	if active {
		tab = us.RA2TabActive
	} else {
		tab = us.RA2TabInactive
	}
	if tab != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(w)/float64(tab.Bounds().Dx()), float64(h)/float64(tab.Bounds().Dy()))
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(tab, op)
		return
	}
	// Fallback to DrawRectButton
	state := "normal"
	if active {
		state = "active"
	}
	us.DrawRectButton(screen, x, y, w, h, state)
}

// GetCommandIcon returns the icon for a command type
func (us *UISprites) GetCommandIcon(cmd CommandType) *ebiten.Image {
	switch cmd {
	case CmdMove:
		return us.IconMove
	case CmdAttack:
		return us.IconAttack
	case CmdStop:
		return us.IconStop
	case CmdGuard:
		return us.IconGuard
	case CmdRally:
		return us.IconRally
	case CmdDeploy:
		return us.IconDeploy
	case CmdSell:
		return us.IconSell
	}
	return nil
}

// ---- Procedural texture generation ----

func generateDarkMetalPanel(w, h int) *ebiten.Image {
	img := ebiten.NewImage(w, h)

	// Create metallic gradient with noise
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Base dark blue-grey
			base := 22.0
			// Vertical gradient (slightly lighter at top)
			grad := base + 8.0*(1.0-float64(y)/float64(h))
			// Brushed metal horizontal lines
			lineNoise := 2.0 * math.Sin(float64(y)*0.8+float64(x)*0.01)
			// Subtle noise
			noise := 3.0 * math.Sin(float64(x*7919+y*7927)*0.001)

			v := grad + lineNoise + noise
			r := uint8(math.Max(0, math.Min(255, v*0.9)))
			g := uint8(math.Max(0, math.Min(255, v*0.95)))
			b := uint8(math.Max(0, math.Min(255, v*1.2)))

			img.Set(x, y, color.NRGBA{r, g, b, 230})
		}
	}
	return img
}

func generateBrushedMetal(w, h int) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			base := 50.0
			brush := 10.0 * math.Sin(float64(x)*0.3+float64(y)*0.1)
			v := base + brush
			r := uint8(math.Max(0, math.Min(255, v)))
			g := uint8(math.Max(0, math.Min(255, v*1.05)))
			b := uint8(math.Max(0, math.Min(255, v*1.15)))
			img.Set(x, y, color.NRGBA{r, g, b, 255})
		}
	}
	return img
}

func generateRivet(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	cx, cy := float64(size)/2, float64(size)/2
	r := float64(size)/2 - 1

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist <= r {
				// 3D shading
				nx := dx / r
				ny := dy / r
				light := 0.5 + 0.5*(nx*0.3-ny*0.5)
				v := uint8(math.Max(0, math.Min(255, light*180)))
				img.Set(x, y, color.NRGBA{v, v, uint8(float64(v) * 1.1), 255})
			}
		}
	}
	return img
}

func generateGlowLine(w, h int, clr color.NRGBA) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	cy := float64(h) / 2
	for y := 0; y < h; y++ {
		dist := math.Abs(float64(y) - cy) / cy
		alpha := uint8(float64(clr.A) * (1.0 - dist*dist))
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{clr.R, clr.G, clr.B, alpha})
		}
	}
	return img
}

func generateFramePiece(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if x < 2 || y < 2 || x >= size-2 || y >= size-2 {
				// Outer border: light bevel
				v := uint8(80)
				if x < 1 || y < 1 {
					v = 100 // highlight
				}
				img.Set(x, y, color.NRGBA{v, v, uint8(float64(v) * 1.2), 255})
			}
		}
	}
	return img
}

// ---- Helper functions ----

func tilePanel(dst *ebiten.Image, src *ebiten.Image, w, h int) {
	if src == nil {
		return
	}
	sw := src.Bounds().Dx()
	sh := src.Bounds().Dy()
	for y := 0; y < h; y += sh {
		for x := 0; x < w; x += sw {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			dst.DrawImage(src, op)
		}
	}
}

func drawRivetAt(dst *ebiten.Image, rivet *ebiten.Image, x, y int) {
	if rivet == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(rivet, op)
}

func drawBevelRect(dst *ebiten.Image, x, y, w, h int, highlight, shadow color.NRGBA) {
	// Top + Left = highlight
	for i := 0; i < w; i++ {
		dst.Set(x+i, y, highlight)
		dst.Set(x+i, y+1, color.NRGBA{highlight.R, highlight.G, highlight.B, highlight.A / 2})
	}
	for i := 0; i < h; i++ {
		dst.Set(x, y+i, highlight)
		dst.Set(x+1, y+i, color.NRGBA{highlight.R, highlight.G, highlight.B, highlight.A / 2})
	}
	// Bottom + Right = shadow
	for i := 0; i < w; i++ {
		dst.Set(x+i, y+h-1, shadow)
		dst.Set(x+i, y+h-2, color.NRGBA{shadow.R, shadow.G, shadow.B, shadow.A / 2})
	}
	for i := 0; i < h; i++ {
		dst.Set(x+w-1, y+i, shadow)
		dst.Set(x+w-2, y+i, color.NRGBA{shadow.R, shadow.G, shadow.B, shadow.A / 2})
	}
}

func drawBevelBorder(dst *ebiten.Image, w, h int) {
	highlight := color.NRGBA{70, 85, 110, 200}
	shadow := color.NRGBA{10, 15, 25, 200}
	drawBevelRect(dst, 0, 0, w, h, highlight, shadow)
}

func getUIAssetsDir() string {
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Join(filepath.Dir(exe), "assets", "ui")
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "..", "..", "assets", "ui")
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	if _, err := os.Stat(filepath.Join("assets", "ui")); err == nil {
		return filepath.Join("assets", "ui")
	}
	return filepath.Join("assets", "ui")
}

func loadUI(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		log.Printf("Warning: could not decode UI sprite %s: %v", path, err)
		return nil
	}
	return ebiten.NewImageFromImage(img)
}

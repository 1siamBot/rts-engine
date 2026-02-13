// generate_ui creates RA2-inspired UI texture assets programmatically.
// These are original creations inspired by the dark-steel military aesthetic
// of classic isometric RTS games. No copyrighted assets are reproduced.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

func main() {
	outDir := "assets/ra2/ui"
	os.MkdirAll(filepath.Join(outDir, "sidebar"), 0755)
	os.MkdirAll(filepath.Join(outDir, "icons"), 0755)
	os.MkdirAll(filepath.Join(outDir, "bars"), 0755)
	os.MkdirAll(filepath.Join(outDir, "panels"), 0755)
	os.MkdirAll(filepath.Join(outDir, "cursors"), 0755)

	// === SIDEBAR PANEL TEXTURES ===
	fmt.Println("Generating sidebar textures...")
	savePNG(filepath.Join(outDir, "panels", "sidebar_bg.png"), genSidebarBG(220, 680))
	savePNG(filepath.Join(outDir, "panels", "topbar_bg.png"), genTopBarBG(1280, 44))
	savePNG(filepath.Join(outDir, "panels", "bottom_panel.png"), genBottomPanel(800, 100))
	savePNG(filepath.Join(outDir, "panels", "minimap_frame.png"), genMinimapFrame(172, 192))
	savePNG(filepath.Join(outDir, "panels", "dark_steel_tile.png"), genDarkSteelTile(64, 64))
	savePNG(filepath.Join(outDir, "panels", "panel_divider.png"), genPanelDivider(200, 3))

	// === TAB BUTTONS ===
	fmt.Println("Generating tab buttons...")
	savePNG(filepath.Join(outDir, "sidebar", "tab_buildings.png"), genTabButton(64, 28, "BLD", true))
	savePNG(filepath.Join(outDir, "sidebar", "tab_units.png"), genTabButton(64, 28, "UNT", false))
	savePNG(filepath.Join(outDir, "sidebar", "tab_defense.png"), genTabButton(64, 28, "DEF", false))
	savePNG(filepath.Join(outDir, "sidebar", "tab_active.png"), genTabButtonState(64, 28, true))
	savePNG(filepath.Join(outDir, "sidebar", "tab_inactive.png"), genTabButtonState(64, 28, false))

	// === BUILD SLOT BUTTONS ===
	fmt.Println("Generating build slot buttons...")
	savePNG(filepath.Join(outDir, "sidebar", "build_slot_normal.png"), genBuildSlot(190, 52, "normal"))
	savePNG(filepath.Join(outDir, "sidebar", "build_slot_hover.png"), genBuildSlot(190, 52, "hover"))
	savePNG(filepath.Join(outDir, "sidebar", "build_slot_disabled.png"), genBuildSlot(190, 52, "disabled"))
	savePNG(filepath.Join(outDir, "sidebar", "build_slot_active.png"), genBuildSlot(190, 52, "active"))

	// === BUILD ICONS (rendered 3D-preview-style) ===
	fmt.Println("Generating build icons...")
	genBuildingIcons(filepath.Join(outDir, "icons"))

	// === COMMAND ICONS ===
	fmt.Println("Generating command icons...")
	savePNG(filepath.Join(outDir, "icons", "cmd_move.png"), genCommandIcon(32, "move"))
	savePNG(filepath.Join(outDir, "icons", "cmd_attack.png"), genCommandIcon(32, "attack"))
	savePNG(filepath.Join(outDir, "icons", "cmd_stop.png"), genCommandIcon(32, "stop"))
	savePNG(filepath.Join(outDir, "icons", "cmd_guard.png"), genCommandIcon(32, "guard"))
	savePNG(filepath.Join(outDir, "icons", "cmd_deploy.png"), genCommandIcon(32, "deploy"))
	savePNG(filepath.Join(outDir, "icons", "cmd_sell.png"), genCommandIcon(32, "sell"))
	savePNG(filepath.Join(outDir, "icons", "cmd_repair.png"), genCommandIcon(32, "repair"))
	savePNG(filepath.Join(outDir, "icons", "cmd_rally.png"), genCommandIcon(32, "rally"))

	// === RESOURCE ICONS ===
	fmt.Println("Generating resource icons...")
	savePNG(filepath.Join(outDir, "icons", "credits.png"), genCreditsIcon(24))
	savePNG(filepath.Join(outDir, "icons", "power.png"), genPowerIcon(24))

	// === BARS ===
	fmt.Println("Generating bar textures...")
	savePNG(filepath.Join(outDir, "bars", "health_green.png"), genBarTexture(128, 12, colGreen))
	savePNG(filepath.Join(outDir, "bars", "health_yellow.png"), genBarTexture(128, 12, colYellow))
	savePNG(filepath.Join(outDir, "bars", "health_red.png"), genBarTexture(128, 12, colRed))
	savePNG(filepath.Join(outDir, "bars", "power_bar.png"), genBarTexture(128, 16, colCyan))
	savePNG(filepath.Join(outDir, "bars", "progress_bar.png"), genBarTexture(128, 8, colGold))
	savePNG(filepath.Join(outDir, "bars", "bar_bg.png"), genBarBG(128, 16))
	savePNG(filepath.Join(outDir, "bars", "power_bar_bg.png"), genPowerBarBG(128, 20))

	// === CURSORS ===
	fmt.Println("Generating cursor sprites...")
	savePNG(filepath.Join(outDir, "cursors", "cursor_default.png"), genCursor(24, "default"))
	savePNG(filepath.Join(outDir, "cursors", "cursor_move.png"), genCursor(24, "move"))
	savePNG(filepath.Join(outDir, "cursors", "cursor_attack.png"), genCursor(24, "attack"))
	savePNG(filepath.Join(outDir, "cursors", "cursor_deploy.png"), genCursor(24, "deploy"))
	savePNG(filepath.Join(outDir, "cursors", "cursor_select.png"), genCursor(24, "select"))

	// === SELECTION BOX ===
	savePNG(filepath.Join(outDir, "panels", "selection_corner.png"), genSelectionCorner(8))

	fmt.Println("✅ All RA2-inspired UI assets generated!")
}

// Color palette
var (
	colSteel     = color.NRGBA{45, 50, 62, 255}
	colSteelDark = color.NRGBA{22, 25, 35, 255}
	colSteelLit  = color.NRGBA{68, 75, 90, 255}
	colRivet     = color.NRGBA{55, 60, 72, 255}
	colGlow      = color.NRGBA{0, 160, 220, 255}
	colGlowDim   = color.NRGBA{0, 80, 120, 120}
	colGreen     = color.NRGBA{30, 200, 50, 255}
	colYellow    = color.NRGBA{240, 200, 20, 255}
	colRed       = color.NRGBA{220, 40, 30, 255}
	colCyan      = color.NRGBA{0, 200, 240, 255}
	colGold      = color.NRGBA{255, 200, 50, 255}
	colBevel     = color.NRGBA{80, 88, 105, 200}
	colShadow    = color.NRGBA{10, 12, 18, 200}
)

// === PANEL GENERATORS ===

func genSidebarBG(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Dark steel gradient with subtle vertical noise
			base := 24.0 + 6.0*float64(x)/float64(w)
			noise := 2.0 * math.Sin(float64(y)*0.7+float64(x)*0.03)
			brushed := 1.5 * math.Sin(float64(y)*2.1+float64(x)*0.02)
			v := base + noise + brushed
			r := clamp8(v * 0.88)
			g := clamp8(v * 0.92)
			b := clamp8(v * 1.15)
			img.Set(x, y, color.NRGBA{r, g, b, 240})
		}
	}
	// Left edge bevel (highlight)
	for y := 0; y < h; y++ {
		img.Set(0, y, colBevel)
		img.Set(1, y, color.NRGBA{65, 72, 88, 180})
	}
	// Right edge shadow
	for y := 0; y < h; y++ {
		img.Set(w-1, y, colShadow)
		img.Set(w-2, y, color.NRGBA{15, 18, 26, 160})
	}
	// Rivets along edges
	for y := 20; y < h-20; y += 40 {
		drawRivet(img, 5, y, 6)
		drawRivet(img, w-11, y, 6)
	}
	// Horizontal divider lines (every ~120px for panel sections)
	for y := 80; y < h-40; y += 120 {
		drawHLine(img, 8, w-8, y, color.NRGBA{50, 56, 68, 200})
		drawHLine(img, 8, w-8, y+1, color.NRGBA{18, 20, 28, 200})
	}
	// Glow line at top
	drawGlowLineH(img, 0, w, 0, colGlow, 3)
	return img
}

func genTopBarBG(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			base := 18.0 + 12.0*(1.0-float64(y)/float64(h))
			noise := 1.5 * math.Sin(float64(x)*0.05+float64(y)*0.8)
			v := base + noise
			img.Set(x, y, color.NRGBA{clamp8(v * 0.85), clamp8(v * 0.9), clamp8(v * 1.2), 235})
		}
	}
	// Bottom glow line
	drawGlowLineH(img, 0, w, h-3, colGlow, 3)
	// Top highlight
	drawHLine(img, 0, w, 0, color.NRGBA{70, 78, 95, 200})
	// Rivets
	drawRivet(img, 6, h/2-3, 6)
	drawRivet(img, w-12, h/2-3, 6)
	return img
}

func genBottomPanel(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			base := 20.0 + 8.0*float64(y)/float64(h)
			v := base + 1.5*math.Sin(float64(x)*0.04+float64(y)*0.5)
			img.Set(x, y, color.NRGBA{clamp8(v * 0.85), clamp8(v * 0.9), clamp8(v * 1.15), 230})
		}
	}
	drawGlowLineH(img, 0, w, 0, colGlow, 3)
	drawBevelRectImg(img, 0, 0, w, h)
	return img
}

func genMinimapFrame(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// Steel frame
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			base := 30.0
			v := base + 3.0*math.Sin(float64(x)*0.1+float64(y)*0.08)
			img.Set(x, y, color.NRGBA{clamp8(v * 0.88), clamp8(v * 0.92), clamp8(v * 1.15), 240})
		}
	}
	// Inner cutout (transparent for minimap content)
	border := 6
	for y := border + 20; y < h-border; y++ {
		for x := border; x < w-border; x++ {
			img.Set(x, y, color.NRGBA{8, 10, 18, 240})
		}
	}
	// Inner bevel
	for x := border; x < w-border; x++ {
		img.Set(x, border+20, color.NRGBA{15, 18, 28, 255})
		img.Set(x, border+21, color.NRGBA{12, 14, 22, 255})
		img.Set(x, h-border-1, colBevel)
	}
	for y := border + 20; y < h-border; y++ {
		img.Set(border, y, color.NRGBA{15, 18, 28, 255})
		img.Set(w-border-1, y, colBevel)
	}
	// Top label area
	drawGlowLineH(img, 4, w-4, 18, colGlow, 2)
	// Corner rivets
	drawRivet(img, 3, 3, 5)
	drawRivet(img, w-8, 3, 5)
	drawRivet(img, 3, h-8, 5)
	drawRivet(img, w-8, h-8, 5)
	drawBevelRectImg(img, 0, 0, w, h)
	return img
}

func genDarkSteelTile(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			base := 22.0
			v := base + 3.0*math.Sin(float64(y)*1.5+float64(x)*0.05) + 2.0*math.Sin(float64(x*7919+y*7927)*0.001)
			img.Set(x, y, color.NRGBA{clamp8(v * 0.88), clamp8(v * 0.92), clamp8(v * 1.15), 230})
		}
	}
	return img
}

func genPanelDivider(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	drawHLine(img, 0, w, 0, color.NRGBA{50, 56, 68, 200})
	drawHLine(img, 0, w, 1, colGlowDim)
	drawHLine(img, 0, w, 2, color.NRGBA{15, 18, 26, 200})
	return img
}

// === TAB BUTTONS ===

func genTabButton(w, h int, label string, active bool) *image.RGBA {
	return genTabButtonState(w, h, active)
}

func genTabButtonState(w, h int, active bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var base float64
			if active {
				base = 38.0 + 8.0*(1.0-float64(y)/float64(h))
			} else {
				base = 25.0 + 4.0*(1.0-float64(y)/float64(h))
			}
			v := base + 1.5*math.Sin(float64(x)*0.1)
			img.Set(x, y, color.NRGBA{clamp8(v * 0.85), clamp8(v * 0.9), clamp8(v * 1.2), 240})
		}
	}
	if active {
		drawGlowLineH(img, 0, w, h-2, colGlow, 2)
	}
	drawBevelRectImg(img, 0, 0, w, h)
	return img
}

// === BUILD SLOT BUTTONS ===

func genBuildSlot(w, h int, state string) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var baseVal float64
	var tintR, tintG, tintB float64

	switch state {
	case "hover":
		baseVal = 35.0
		tintR, tintG, tintB = 0.9, 1.0, 1.3
	case "active":
		baseVal = 32.0
		tintR, tintG, tintB = 0.7, 1.0, 1.5
	case "disabled":
		baseVal = 18.0
		tintR, tintG, tintB = 0.8, 0.8, 0.9
	default: // normal
		baseVal = 28.0
		tintR, tintG, tintB = 0.88, 0.92, 1.15
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			grad := baseVal + 4.0*(1.0-float64(y)/float64(h))
			v := grad + 1.0*math.Sin(float64(y)*0.8)
			img.Set(x, y, color.NRGBA{clamp8(v * tintR), clamp8(v * tintG), clamp8(v * tintB), 240})
		}
	}

	// Icon area inset (left side, 44x44)
	insetX, insetY, insetW, insetH := 3, 3, 46, h-6
	for y := insetY; y < insetY+insetH; y++ {
		for x := insetX; x < insetX+insetW; x++ {
			img.Set(x, y, color.NRGBA{12, 14, 22, 250})
		}
	}
	// Inset bevel
	for x := insetX; x < insetX+insetW; x++ {
		img.Set(x, insetY, color.NRGBA{8, 10, 16, 255})
		img.Set(x, insetY+insetH-1, color.NRGBA{55, 60, 72, 180})
	}
	for y := insetY; y < insetY+insetH; y++ {
		img.Set(insetX, y, color.NRGBA{8, 10, 16, 255})
		img.Set(insetX+insetW-1, y, color.NRGBA{55, 60, 72, 180})
	}

	if state == "active" {
		drawGlowLineH(img, 0, w, h-2, colCyan, 2)
	} else if state == "hover" {
		drawGlowLineH(img, 0, w, h-2, color.NRGBA{0, 120, 180, 150}, 2)
	}

	drawBevelRectImg(img, 0, 0, w, h)
	return img
}

// === BUILD ICONS (procedurally rendered) ===

func genBuildingIcons(outDir string) {
	icons := []struct{ name string; r, g, b uint8; shape string }{
		{"construction_yard", 180, 160, 60, "conyard"},
		{"power_plant", 60, 180, 60, "power"},
		{"barracks", 60, 100, 180, "barracks"},
		{"refinery", 180, 140, 40, "refinery"},
		{"war_factory", 120, 120, 140, "factory"},
	}
	for _, ic := range icons {
		img := genBuildIcon(44, ic.r, ic.g, ic.b, ic.shape)
		savePNG(filepath.Join(outDir, "build_"+ic.name+".png"), img)
	}
}

func genBuildIcon(size int, r, g, b uint8, shape string) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// Dark background
	fill(img, color.NRGBA{12, 14, 22, 255})
	cx, cy := float64(size)/2, float64(size)/2

	switch shape {
	case "conyard":
		// Construction Yard: large building with crane
		drawIsoBox(img, cx, cy+4, 28, 16, 12, color.NRGBA{r, g, b, 255})
		// Crane arm
		for i := 0; i < 18; i++ {
			px := int(cx) - 4 + i
			py := int(cy) - 10
			if px >= 0 && px < size && py >= 0 && py < size {
				img.Set(px, py, color.NRGBA{200, 180, 60, 255})
				img.Set(px, py+1, color.NRGBA{160, 140, 40, 255})
			}
		}
	case "power":
		// Power Plant: building with cooling tower
		drawIsoBox(img, cx-2, cy+6, 22, 12, 10, color.NRGBA{r, g, b, 255})
		// Cooling tower (cylinder)
		drawCircleFill(img, int(cx)+6, int(cy)-2, 7, color.NRGBA{180, 180, 180, 255})
		drawCircleFill(img, int(cx)+5, int(cy)-3, 4, color.NRGBA{200, 200, 200, 200})
	case "barracks":
		drawIsoBox(img, cx, cy+4, 24, 14, 14, color.NRGBA{r, g, b, 255})
		// Door
		for y := int(cy) + 2; y < int(cy)+10; y++ {
			for x := int(cx) - 3; x < int(cx)+3; x++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, color.NRGBA{30, 25, 20, 255})
				}
			}
		}
	case "refinery":
		drawIsoBox(img, cx, cy+6, 26, 14, 10, color.NRGBA{r, g, b, 255})
		// Silos
		drawCircleFill(img, int(cx)-5, int(cy)-4, 5, color.NRGBA{160, 160, 160, 255})
		drawCircleFill(img, int(cx)+5, int(cy)-2, 4, color.NRGBA{150, 150, 150, 255})
	case "factory":
		drawIsoBox(img, cx, cy+4, 28, 16, 16, color.NRGBA{r, g, b, 255})
		// Garage door (dark opening)
		for y := int(cy); y < int(cy)+10; y++ {
			for x := int(cx) - 6; x < int(cx)+6; x++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, color.NRGBA{15, 15, 15, 255})
				}
			}
		}
	}
	// Outer bevel
	drawBevelRectImg(img, 0, 0, size, size)
	return img
}

// === COMMAND ICONS ===

func genCommandIcon(size int, cmd string) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := size/2, size/2

	switch cmd {
	case "move":
		// Green arrow pointing right
		drawArrow(img, cx, cy, 10, color.NRGBA{40, 220, 80, 255})
	case "attack":
		// Red crosshair
		drawCrosshair(img, cx, cy, 11, color.NRGBA{220, 50, 40, 255})
	case "stop":
		// Red octagon outline
		drawCircleFill(img, cx, cy, 11, color.NRGBA{200, 40, 30, 255})
		drawCircleFill(img, cx, cy, 8, color.NRGBA{12, 14, 22, 255})
		// Horizontal bar
		for x := cx - 5; x <= cx+5; x++ {
			for y := cy - 2; y <= cy+2; y++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, color.NRGBA{220, 50, 40, 255})
				}
			}
		}
	case "guard":
		// Shield shape
		drawShield(img, cx, cy, 10, color.NRGBA{60, 120, 220, 255})
	case "deploy":
		// Down arrow into base
		drawArrowDown(img, cx, cy, 10, color.NRGBA{200, 180, 50, 255})
		// Base line
		for x := cx - 8; x <= cx+8; x++ {
			img.Set(x, cy+10, color.NRGBA{200, 180, 50, 255})
		}
	case "sell":
		// Dollar sign
		drawCircleFill(img, cx, cy, 11, color.NRGBA{40, 160, 50, 255})
		// S shape inside (simplified)
		drawCircleFill(img, cx, cy, 8, color.NRGBA{20, 100, 30, 255})
		drawCircleFill(img, cx, cy, 5, color.NRGBA{40, 160, 50, 255})
	case "repair":
		// Wrench shape
		drawCircleFill(img, cx-4, cy-4, 5, color.NRGBA{180, 180, 60, 255})
		drawCircleFill(img, cx+4, cy+4, 5, color.NRGBA{180, 180, 60, 255})
		// Handle between
		for i := 0; i < 10; i++ {
			px := cx - 4 + i
			py := cy - 4 + i
			if px >= 0 && px < size && py >= 0 && py < size {
				img.Set(px, py, color.NRGBA{160, 160, 50, 255})
				if py+1 < size {
					img.Set(px, py+1, color.NRGBA{160, 160, 50, 255})
				}
			}
		}
	case "rally":
		// Flag on pole
		// Pole
		for y := cy - 10; y <= cy+10; y++ {
			if y >= 0 && y < size {
				img.Set(cx-2, y, color.NRGBA{160, 160, 160, 255})
			}
		}
		// Flag triangle
		for dy := 0; dy < 8; dy++ {
			for dx := 0; dx < 8-dy; dx++ {
				px := cx - 1 + dx
				py := cy - 10 + dy
				if px >= 0 && px < size && py >= 0 && py < size {
					img.Set(px, py, color.NRGBA{220, 50, 50, 255})
				}
			}
		}
	}
	return img
}

// === RESOURCE ICONS ===

func genCreditsIcon(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := size/2, size/2
	// Gold coin
	drawCircleFill(img, cx, cy, size/2-1, color.NRGBA{255, 200, 0, 255})
	drawCircleFill(img, cx-1, cy-1, size/2-3, color.NRGBA{255, 230, 80, 255})
	// Border
	drawCircleOutline(img, cx, cy, size/2-1, color.NRGBA{180, 140, 0, 255})
	return img
}

func genPowerIcon(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx := size / 2
	// Lightning bolt
	pts := [][2]int{
		{cx + 2, 2}, {cx - 4, size/2 + 1}, {cx, size/2 - 1},
		{cx - 2, size - 2}, {cx + 4, size/2 - 1}, {cx, size/2 + 1},
	}
	for i := 0; i < len(pts)-1; i++ {
		drawLinePx(img, pts[i][0], pts[i][1], pts[i+1][0], pts[i+1][1], color.NRGBA{0, 220, 255, 255})
	}
	return img
}

// === BAR TEXTURES ===

func genBarTexture(w, h int, clr color.NRGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		t := float64(y) / float64(h)
		// Glossy gradient: lighter at top, darker at bottom, with highlight band
		var brightness float64
		if t < 0.3 {
			brightness = 1.2 - t*0.5
		} else if t < 0.5 {
			brightness = 1.05
		} else {
			brightness = 1.0 - (t-0.5)*0.4
		}
		for x := 0; x < w; x++ {
			r := clamp8(float64(clr.R) * brightness)
			g := clamp8(float64(clr.G) * brightness)
			b := clamp8(float64(clr.B) * brightness)
			img.Set(x, y, color.NRGBA{r, g, b, clr.A})
		}
	}
	return img
}

func genBarBG(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{15, 18, 28, 220})
		}
	}
	drawBevelRectImg(img, 0, 0, w, h)
	return img
}

func genPowerBarBG(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{12, 15, 25, 230})
		}
	}
	// Tick marks
	for x := 0; x < w; x += w / 10 {
		for y := 0; y < h; y++ {
			img.Set(x, y, color.NRGBA{40, 45, 55, 200})
		}
	}
	drawBevelRectImg(img, 0, 0, w, h)
	return img
}

// === CURSORS ===

func genCursor(size int, cursorType string) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	switch cursorType {
	case "default":
		// Arrow cursor
		for i := 0; i < 16; i++ {
			for j := 0; j <= i && j < 10; j++ {
				if j < size && i < size {
					img.Set(j, i, color.NRGBA{220, 230, 255, 255})
				}
			}
		}
		// Border
		for i := 0; i < 16 && i < size; i++ {
			img.Set(0, i, color.NRGBA{0, 0, 0, 255})
			if i < size {
				img.Set(i, i, color.NRGBA{0, 0, 0, 255})
			}
		}
	case "move":
		drawArrow(img, size/2, size/2, size/2-2, color.NRGBA{40, 220, 80, 255})
	case "attack":
		drawCrosshair(img, size/2, size/2, size/2-2, color.NRGBA{220, 50, 40, 255})
	case "deploy":
		drawArrowDown(img, size/2, size/2, size/2-2, color.NRGBA{200, 180, 50, 255})
	case "select":
		drawCircleOutline(img, size/2, size/2, size/2-2, color.NRGBA{0, 255, 100, 200})
	}
	return img
}

func genSelectionCorner(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	clr := color.NRGBA{0, 255, 100, 220}
	// L-shape corner
	for x := 0; x < size; x++ {
		img.Set(x, 0, clr)
		img.Set(x, 1, clr)
	}
	for y := 0; y < size; y++ {
		img.Set(0, y, clr)
		img.Set(1, y, clr)
	}
	return img
}

// === DRAWING HELPERS ===

func fill(img *image.RGBA, c color.NRGBA) {
	draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)
}

func clamp8(v float64) uint8 {
	if v < 0 { return 0 }
	if v > 255 { return 255 }
	return uint8(v)
}

func drawHLine(img *image.RGBA, x0, x1, y int, c color.NRGBA) {
	b := img.Bounds()
	for x := x0; x < x1 && x < b.Max.X; x++ {
		if y >= 0 && y < b.Max.Y {
			img.Set(x, y, c)
		}
	}
}

func drawGlowLineH(img *image.RGBA, x0, x1, y int, c color.NRGBA, thickness int) {
	b := img.Bounds()
	for dy := -thickness; dy <= thickness; dy++ {
		py := y + dy
		if py < 0 || py >= b.Max.Y { continue }
		dist := math.Abs(float64(dy)) / float64(thickness+1)
		alpha := uint8(float64(c.A) * (1.0 - dist*dist))
		for x := x0; x < x1 && x < b.Max.X; x++ {
			img.Set(x, py, color.NRGBA{c.R, c.G, c.B, alpha})
		}
	}
}

func drawRivet(img *image.RGBA, x, y, size int) {
	cx, cy := float64(x)+float64(size)/2, float64(y)+float64(size)/2
	r := float64(size) / 2
	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			px, py := x+dx, y+dy
			b := img.Bounds()
			if px < 0 || px >= b.Max.X || py < 0 || py >= b.Max.Y { continue }
			ddx := float64(px) - cx
			ddy := float64(py) - cy
			dist := math.Sqrt(ddx*ddx + ddy*ddy)
			if dist <= r {
				nx := ddx / r
				ny := ddy / r
				light := 0.5 + 0.5*(nx*0.3-ny*0.5)
				v := light * 140
				img.Set(px, py, color.NRGBA{clamp8(v), clamp8(v * 1.05), clamp8(v * 1.15), 255})
			}
		}
	}
}

func drawCircleFill(img *image.RGBA, cx, cy, r int, c color.NRGBA) {
	b := img.Bounds()
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				px, py := cx+x, cy+y
				if px >= 0 && px < b.Max.X && py >= 0 && py < b.Max.Y {
					img.Set(px, py, c)
				}
			}
		}
	}
}

func drawCircleOutline(img *image.RGBA, cx, cy, r int, c color.NRGBA) {
	for a := 0.0; a < math.Pi*2; a += 0.05 {
		px := cx + int(float64(r)*math.Cos(a))
		py := cy + int(float64(r)*math.Sin(a))
		b := img.Bounds()
		if px >= 0 && px < b.Max.X && py >= 0 && py < b.Max.Y {
			img.Set(px, py, c)
		}
	}
}

func drawIsoBox(img *image.RGBA, cx, cy float64, w, h, depth int, c color.NRGBA) {
	// Top face (lighter)
	topC := color.NRGBA{clamp8(float64(c.R) * 1.3), clamp8(float64(c.G) * 1.3), clamp8(float64(c.B) * 1.3), c.A}
	// Front face
	for dy := 0; dy < h; dy++ {
		for dx := -w / 2; dx < w/2; dx++ {
			px := int(cx) + dx
			py := int(cy) + dy - h/2
			b := img.Bounds()
			if px >= 0 && px < b.Max.X && py >= 0 && py < b.Max.Y {
				img.Set(px, py, c)
			}
		}
	}
	// Top face
	for dx := -w / 2; dx < w/2; dx++ {
		for dy := 0; dy < depth/2; dy++ {
			px := int(cx) + dx
			py := int(cy) - h/2 - dy
			b := img.Bounds()
			if px >= 0 && px < b.Max.X && py >= 0 && py < b.Max.Y {
				img.Set(px, py, topC)
			}
		}
	}
	// Right face (darker)
	sideC := color.NRGBA{clamp8(float64(c.R) * 0.7), clamp8(float64(c.G) * 0.7), clamp8(float64(c.B) * 0.7), c.A}
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < depth/3; dx++ {
			px := int(cx) + w/2 + dx
			py := int(cy) + dy - h/2 - dx
			b := img.Bounds()
			if px >= 0 && px < b.Max.X && py >= 0 && py < b.Max.Y {
				img.Set(px, py, sideC)
			}
		}
	}
}

func drawArrow(img *image.RGBA, cx, cy, size int, c color.NRGBA) {
	// Right-pointing arrow
	b := img.Bounds()
	for i := 0; i < size; i++ {
		px := cx - size/2 + i
		if px >= 0 && px < b.Max.X {
			for dy := -1; dy <= 1; dy++ {
				py := cy + dy
				if py >= 0 && py < b.Max.Y { img.Set(px, py, c) }
			}
		}
	}
	// Arrow head
	for i := 0; i < size/2; i++ {
		for dy := -i; dy <= i; dy++ {
			px := cx + size/2 - i
			py := cy + dy
			if px >= 0 && px < b.Max.X && py >= 0 && py < b.Max.Y {
				img.Set(px, py, c)
			}
		}
	}
}

func drawArrowDown(img *image.RGBA, cx, cy, size int, c color.NRGBA) {
	b := img.Bounds()
	for i := 0; i < size; i++ {
		py := cy - size/2 + i
		if py >= 0 && py < b.Max.Y {
			for dx := -1; dx <= 1; dx++ {
				px := cx + dx
				if px >= 0 && px < b.Max.X { img.Set(px, py, c) }
			}
		}
	}
	for i := 0; i < size/2; i++ {
		for dx := -i; dx <= i; dx++ {
			py := cy + size/2 - i
			px := cx + dx
			if px >= 0 && px < b.Max.X && py >= 0 && py < b.Max.Y {
				img.Set(px, py, c)
			}
		}
	}
}

func drawCrosshair(img *image.RGBA, cx, cy, size int, c color.NRGBA) {
	b := img.Bounds()
	// Circle
	drawCircleOutline(img, cx, cy, size, c)
	// Cross lines with gap in center
	for i := -size; i <= size; i++ {
		if abs(i) < 3 { continue }
		px := cx + i
		if px >= 0 && px < b.Max.X {
			if cy >= 0 && cy < b.Max.Y { img.Set(px, cy, c) }
		}
		py := cy + i
		if py >= 0 && py < b.Max.Y {
			if cx >= 0 && cx < b.Max.X { img.Set(cx, py, c) }
		}
	}
}

func drawShield(img *image.RGBA, cx, cy, size int, c color.NRGBA) {
	b := img.Bounds()
	for dy := -size; dy <= size; dy++ {
		halfW := size - abs(dy)/2
		if dy > size/2 { halfW = size - dy }
		for dx := -halfW; dx <= halfW; dx++ {
			px, py := cx+dx, cy+dy
			if px >= 0 && px < b.Max.X && py >= 0 && py < b.Max.Y {
				// Edge detection for outline
				edgeDist := halfW - abs(dx)
				if edgeDist < 2 || abs(dy) == size || (dy > size/2 && edgeDist < 3) {
					img.Set(px, py, c)
				} else {
					img.Set(px, py, color.NRGBA{c.R / 2, c.G / 2, c.B / 2, c.A / 2})
				}
			}
		}
	}
}

func drawLinePx(img *image.RGBA, x0, y0, x1, y1 int, c color.NRGBA) {
	b := img.Bounds()
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	steps := dx
	if dy > dx { steps = dy }
	if steps == 0 { return }
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		px := int(float64(x0) + t*float64(x1-x0))
		py := int(float64(y0) + t*float64(y1-y0))
		if px >= 0 && px < b.Max.X && py >= 0 && py < b.Max.Y {
			img.Set(px, py, c)
			if px+1 < b.Max.X { img.Set(px+1, py, c) }
		}
	}
}

func drawBevelRectImg(img *image.RGBA, x, y, w, h int) {
	b := img.Bounds()
	hi := color.NRGBA{70, 78, 95, 180}
	sh := color.NRGBA{10, 12, 18, 180}
	for i := x; i < x+w && i < b.Max.X; i++ {
		if y >= 0 && y < b.Max.Y { img.Set(i, y, hi) }
		if y+h-1 >= 0 && y+h-1 < b.Max.Y { img.Set(i, y+h-1, sh) }
	}
	for i := y; i < y+h && i < b.Max.Y; i++ {
		if x >= 0 && x < b.Max.X { img.Set(x, i, hi) }
		if x+w-1 >= 0 && x+w-1 < b.Max.X { img.Set(x+w-1, i, sh) }
	}
}

func abs(x int) int {
	if x < 0 { return -x }
	return x
}

func savePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("ERROR: %s: %v\n", path, err)
		return
	}
	defer f.Close()
	png.Encode(f, img)
	fmt.Printf("  → %s\n", path)
}

package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"

	"github.com/1siamBot/rts-engine/editor"
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
)

type EditorApp struct {
	editor   *editor.Editor
	renderer *render.IsoRenderer
	input    *input.InputState
	hoverX   int
	hoverY   int

	terrains []maplib.TerrainType
	selIdx   int
}

func NewEditorApp() *EditorApp {
	e := &EditorApp{
		editor:   editor.NewEditor(64, 64),
		renderer: render.NewIsoRenderer(ScreenWidth, ScreenHeight),
		input:    input.NewInputState(),
		terrains: []maplib.TerrainType{
			maplib.TerrainGrass, maplib.TerrainDirt, maplib.TerrainSand,
			maplib.TerrainWater, maplib.TerrainDeepWater, maplib.TerrainRock,
			maplib.TerrainCliff, maplib.TerrainRoad, maplib.TerrainBridge,
			maplib.TerrainOre, maplib.TerrainGem, maplib.TerrainSnow,
			maplib.TerrainUrban, maplib.TerrainForest,
		},
	}
	e.renderer.Camera.CenterOn(32, 32)

	// Load file from command line if provided
	if len(os.Args) > 1 {
		if err := e.editor.LoadMap(os.Args[1]); err != nil {
			log.Printf("Failed to load map: %v", err)
		}
	}

	return e
}

func (a *EditorApp) Update() error {
	a.input.Update()

	// Camera controls
	speed := a.renderer.Camera.Speed / 60.0
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		a.renderer.Camera.Pan(0, -speed)
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		a.renderer.Camera.Pan(0, speed)
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		a.renderer.Camera.Pan(-speed, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		a.renderer.Camera.Pan(speed, 0)
	}
	if a.input.ScrollY != 0 {
		a.renderer.Camera.ZoomAt(a.input.ScrollY*0.1, a.input.MouseX, a.input.MouseY)
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		a.renderer.Camera.Pan(float64(-a.input.MouseDX), float64(-a.input.MouseDY))
	}

	// Hover tile
	wx, wy := a.renderer.Camera.ScreenToWorld(a.input.MouseX, a.input.MouseY)
	a.hoverX = int(math.Floor(wx))
	a.hoverY = int(math.Floor(wy))

	// Terrain selection via number keys
	for i := 0; i < len(a.terrains) && i < 10; i++ {
		if a.input.IsKeyJustPressed(ebiten.Key0 + ebiten.Key(i)) {
			a.selIdx = i
			a.editor.Brush = a.terrains[i]
		}
	}

	// Tool selection
	if a.input.IsKeyJustPressed(ebiten.KeyP) {
		a.editor.Tool = editor.ToolPaint
	}
	if a.input.IsKeyJustPressed(ebiten.KeyH) {
		a.editor.Tool = editor.ToolHeight
	}

	// Brush size
	if a.input.IsKeyJustPressed(ebiten.KeyTab) {
		a.editor.BrushSize++
		if a.editor.BrushSize > 5 {
			a.editor.BrushSize = 1
		}
	}

	// Grid toggle
	if a.input.IsKeyJustPressed(ebiten.KeyG) {
		a.editor.ShowGrid = !a.editor.ShowGrid
	}

	// Paint with left click
	if a.input.LeftPressed && a.input.MouseX < ScreenWidth-200 {
		a.editor.Paint(a.hoverX, a.hoverY)
	}

	// Undo/Redo (Ctrl+Z / Ctrl+Shift+Z)
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	if ctrl && a.input.IsKeyJustPressed(ebiten.KeyZ) {
		if shift {
			a.editor.Redo()
		} else {
			a.editor.Undo()
		}
	}

	// Save (Ctrl+S)
	if ctrl && a.input.IsKeyJustPressed(ebiten.KeyS) {
		path := a.editor.FilePath
		if path == "" {
			path = "map.rtsmap"
		}
		if err := a.editor.SaveMap(path); err != nil {
			log.Printf("Save failed: %v", err)
		} else {
			log.Printf("Saved to %s", path)
		}
	}

	return nil
}

func (a *EditorApp) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 40, 255})

	a.renderer.DrawMap(screen, a.editor.TileMap)
	if a.editor.ShowGrid {
		a.renderer.DrawGrid(screen, a.editor.TileMap)
	}

	// Hover highlight
	if a.editor.TileMap.InBounds(a.hoverX, a.hoverY) {
		sx, sy := a.renderer.Camera.WorldToScreen(float64(a.hoverX), float64(a.hoverY))
		tw := float32(a.editor.TileMap.TileWidth)
		th := float32(a.editor.TileMap.TileHeight)
		hw := tw / 2
		hh := th / 2
		cx := float32(sx)
		cy := float32(sy) + hh
		hoverColor := color.RGBA{255, 255, 0, 150}
		vector.StrokeLine(screen, cx, cy-hh, cx+hw, cy, 2, hoverColor, false)
		vector.StrokeLine(screen, cx+hw, cy, cx, cy+hh, 2, hoverColor, false)
		vector.StrokeLine(screen, cx, cy+hh, cx-hw, cy, 2, hoverColor, false)
		vector.StrokeLine(screen, cx-hw, cy, cx, cy-hh, 2, hoverColor, false)
	}

	// Start positions
	for _, sp := range a.editor.TileMap.StartPositions {
		sx, sy := a.renderer.Camera.WorldToScreen(float64(sp.X)+0.5, float64(sp.Y)+0.5)
		clr := color.RGBA{255, 255, 0, 255}
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), 8, clr, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("P%d", sp.PlayerSlot), sx-5, sy-5)
	}

	// Sidebar
	a.drawSidebar(screen)

	// HUD info
	tile := a.editor.TileMap.At(a.hoverX, a.hoverY)
	tn := "OOB"
	ore := 0
	if tile != nil {
		tn = fmt.Sprintf("%d", tile.Terrain)
		ore = tile.OreAmount
	}
	info := fmt.Sprintf("Map Editor | Tile(%d,%d) %s Ore:%d | Brush:%d Size:%d | [WASD]Pan [Scroll]Zoom [G]Grid [Tab]Size [Ctrl+Z]Undo [Ctrl+S]Save",
		a.hoverX, a.hoverY, tn, ore, a.selIdx, a.editor.BrushSize)
	ebitenutil.DebugPrintAt(screen, info, 5, ScreenHeight-20)
}

func (a *EditorApp) drawSidebar(screen *ebiten.Image) {
	sx := float32(ScreenWidth - 200)
	vector.DrawFilledRect(screen, sx, 0, 200, float32(ScreenHeight), color.RGBA{20, 20, 40, 220}, false)

	y := 10
	ebitenutil.DebugPrintAt(screen, "=== TERRAIN ===", int(sx)+10, y)
	y += 20
	terrainNames := []string{
		"Grass", "Dirt", "Sand", "Water", "DeepWater",
		"Rock", "Cliff", "Road", "Bridge", "Ore",
		"Gem", "Snow", "Urban", "Forest",
	}
	for i, name := range terrainNames {
		clr := color.RGBA{50, 50, 80, 255}
		if i == a.selIdx {
			clr = color.RGBA{100, 100, 200, 255}
		}
		vector.DrawFilledRect(screen, sx+10, float32(y), 180, 20, clr, false)
		label := fmt.Sprintf("[%d] %s", i, name)
		if i >= 10 {
			label = fmt.Sprintf("    %s", name)
		}
		ebitenutil.DebugPrintAt(screen, label, int(sx)+15, y+3)
		y += 22
	}

	y += 10
	tools := []string{"[P] Paint", "[H] Height"}
	for _, t := range tools {
		ebitenutil.DebugPrintAt(screen, t, int(sx)+10, y)
		y += 18
	}

	if a.editor.Modified {
		ebitenutil.DebugPrintAt(screen, "* MODIFIED *", int(sx)+10, y+20)
	}
}

func (a *EditorApp) Layout(_, _ int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("🗺️ RTS Map Editor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	app := NewEditorApp()
	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}

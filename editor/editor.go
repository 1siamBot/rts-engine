package editor

import (
	"github.com/1siamBot/rts-engine/engine/maplib"
)

// Action represents an undoable editor action
type Action struct {
	X, Y     int
	OldTile  maplib.Tile
	NewTile  maplib.Tile
}

// Editor holds map editor state
type Editor struct {
	TileMap      *maplib.TileMap
	Brush        maplib.TerrainType
	BrushSize    int
	Tool         EditorTool
	UndoStack    [][]Action
	RedoStack    [][]Action
	FilePath     string
	Modified     bool
	ShowGrid     bool
	OreAmount    int
}

// EditorTool represents the current editor tool
type EditorTool int

const (
	ToolPaint EditorTool = iota
	ToolErase
	ToolOre
	ToolStartPos
	ToolHeight
)

// NewEditor creates a new map editor
func NewEditor(width, height int) *Editor {
	return &Editor{
		TileMap:   maplib.NewTileMap("Untitled", width, height),
		Brush:     maplib.TerrainGrass,
		BrushSize: 1,
		ShowGrid:  true,
		OreAmount: 1000,
	}
}

// LoadMap loads a map file
func (e *Editor) LoadMap(path string) error {
	tm, err := maplib.LoadJSON(path)
	if err != nil {
		return err
	}
	e.TileMap = tm
	e.FilePath = path
	e.Modified = false
	e.UndoStack = nil
	e.RedoStack = nil
	return nil
}

// SaveMap saves the current map
func (e *Editor) SaveMap(path string) error {
	if path == "" {
		path = e.FilePath
	}
	if path == "" {
		path = "untitled.rtsmap"
	}
	e.FilePath = path
	e.Modified = false
	return e.TileMap.SaveJSON(path)
}

// Paint applies the current brush at (cx, cy) with brush size
func (e *Editor) Paint(cx, cy int) {
	var actions []Action
	r := e.BrushSize / 2
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			x, y := cx+dx, cy+dy
			t := e.TileMap.At(x, y)
			if t == nil {
				continue
			}
			old := *t
			switch e.Tool {
			case ToolPaint:
				e.TileMap.SetTerrain(x, y, x, y, e.Brush)
			case ToolOre:
				e.TileMap.PlaceOre(x, y, e.OreAmount)
			case ToolErase:
				e.TileMap.SetTerrain(x, y, x, y, maplib.TerrainGrass)
				t.OreAmount = 0
			case ToolHeight:
				t.Height++
				if t.Height > 7 {
					t.Height = 0
				}
			}
			newTile := *e.TileMap.At(x, y)
			actions = append(actions, Action{X: x, Y: y, OldTile: old, NewTile: newTile})
		}
	}
	if len(actions) > 0 {
		e.UndoStack = append(e.UndoStack, actions)
		e.RedoStack = nil
		e.Modified = true
	}
}

// SetStartPos sets a player start position
func (e *Editor) SetStartPos(slot, x, y int) {
	for i := range e.TileMap.StartPositions {
		if e.TileMap.StartPositions[i].PlayerSlot == slot {
			e.TileMap.StartPositions[i].X = x
			e.TileMap.StartPositions[i].Y = y
			return
		}
	}
	e.TileMap.StartPositions = append(e.TileMap.StartPositions, maplib.StartPos{
		PlayerSlot: slot, X: x, Y: y,
	})
}

// Undo reverts the last action
func (e *Editor) Undo() {
	if len(e.UndoStack) == 0 {
		return
	}
	actions := e.UndoStack[len(e.UndoStack)-1]
	e.UndoStack = e.UndoStack[:len(e.UndoStack)-1]
	for _, a := range actions {
		t := e.TileMap.At(a.X, a.Y)
		if t != nil {
			*t = a.OldTile
		}
	}
	e.RedoStack = append(e.RedoStack, actions)
	e.Modified = true
}

// Redo re-applies the last undone action
func (e *Editor) Redo() {
	if len(e.RedoStack) == 0 {
		return
	}
	actions := e.RedoStack[len(e.RedoStack)-1]
	e.RedoStack = e.RedoStack[:len(e.RedoStack)-1]
	for _, a := range actions {
		t := e.TileMap.At(a.X, a.Y)
		if t != nil {
			*t = a.NewTile
		}
	}
	e.UndoStack = append(e.UndoStack, actions)
	e.Modified = true
}

// NewMap creates a fresh map
func (e *Editor) NewMap(name string, w, h int) {
	e.TileMap = maplib.NewTileMap(name, w, h)
	e.FilePath = ""
	e.Modified = false
	e.UndoStack = nil
	e.RedoStack = nil
}

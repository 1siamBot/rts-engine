package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// InputState tracks mouse and keyboard state per frame
type InputState struct {
	// Mouse
	MouseX, MouseY     int
	MouseDX, MouseDY   int     // delta since last frame
	prevMouseX         int
	prevMouseY         int
	LeftPressed        bool
	RightPressed       bool
	LeftJustPressed    bool
	RightJustPressed   bool
	LeftJustReleased   bool
	RightJustReleased  bool
	ScrollY            float64

	// Drag
	DragStartX, DragStartY int
	Dragging               bool
	DragThreshold          int

	// Keyboard
	KeysPressed map[ebiten.Key]bool
}

func NewInputState() *InputState {
	return &InputState{
		DragThreshold: 5,
		KeysPressed:   make(map[ebiten.Key]bool),
	}
}

// Update should be called every frame
func (s *InputState) Update() {
	// Mouse position
	s.prevMouseX = s.MouseX
	s.prevMouseY = s.MouseY
	s.MouseX, s.MouseY = ebiten.CursorPosition()
	s.MouseDX = s.MouseX - s.prevMouseX
	s.MouseDY = s.MouseY - s.prevMouseY

	// Mouse buttons
	leftDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	rightDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)

	s.LeftJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	s.RightJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
	s.LeftJustReleased = inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	s.RightJustReleased = inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight)
	s.LeftPressed = leftDown
	s.RightPressed = rightDown

	// Scroll
	_, scrollY := ebiten.Wheel()
	s.ScrollY = scrollY

	// Drag tracking
	if s.LeftJustPressed {
		s.DragStartX = s.MouseX
		s.DragStartY = s.MouseY
		s.Dragging = false
	}
	if leftDown && !s.Dragging {
		dx := s.MouseX - s.DragStartX
		dy := s.MouseY - s.DragStartY
		if dx*dx+dy*dy > s.DragThreshold*s.DragThreshold {
			s.Dragging = true
		}
	}
	if !leftDown {
		s.Dragging = false
	}

	// Common keys
	commonKeys := []ebiten.Key{
		ebiten.KeyW, ebiten.KeyA, ebiten.KeyS, ebiten.KeyD,
		ebiten.KeyUp, ebiten.KeyDown, ebiten.KeyLeft, ebiten.KeyRight,
		ebiten.KeySpace, ebiten.KeyEscape, ebiten.KeyEnter,
		ebiten.KeyShift, ebiten.KeyControl,
		ebiten.KeyDelete, ebiten.KeyBackspace,
		ebiten.KeyTab,
		ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5,
		ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9, ebiten.Key0,
		ebiten.KeyF1, ebiten.KeyF2, ebiten.KeyF3, ebiten.KeyF4, ebiten.KeyF5,
		ebiten.KeyH, ebiten.KeyG, ebiten.KeyP, ebiten.KeyM,
	}
	for _, k := range commonKeys {
		s.KeysPressed[k] = ebiten.IsKeyPressed(k)
	}
}

// IsKeyJustPressed returns true if key was just pressed this frame
func (s *InputState) IsKeyJustPressed(key ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(key)
}

// DragRect returns the selection rectangle if dragging
func (s *InputState) DragRect() (x1, y1, x2, y2 int, active bool) {
	if !s.Dragging {
		return 0, 0, 0, 0, false
	}
	return s.DragStartX, s.DragStartY, s.MouseX, s.MouseY, true
}

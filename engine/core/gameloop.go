package core

import "time"

// GameState represents the overall game state
type GameState uint8

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateLoading
)

// GameLoop manages the fixed-timestep game loop for deterministic simulation
type GameLoop struct {
	World       *World
	State       GameState
	TickRate    float64 // fixed ticks per second
	accumulator float64
	lastTime    time.Time
}

// NewGameLoop creates a game loop with fixed tick rate
func NewGameLoop(tickRate float64) *GameLoop {
	return &GameLoop{
		World:    NewWorld(tickRate),
		TickRate: tickRate,
		lastTime: time.Now(),
	}
}

// Update should be called every render frame. It runs the simulation
// at fixed timestep (important for deterministic multiplayer).
// Returns the interpolation alpha for smooth rendering.
func (gl *GameLoop) Update() float64 {
	now := time.Now()
	frameTime := now.Sub(gl.lastTime).Seconds()
	gl.lastTime = now

	// Cap frame time to avoid spiral of death
	if frameTime > 0.25 {
		frameTime = 0.25
	}

	dt := 1.0 / gl.TickRate
	gl.accumulator += frameTime

	for gl.accumulator >= dt {
		if gl.State == StatePlaying {
			gl.World.Tick(dt)
		}
		gl.accumulator -= dt
	}

	// Return interpolation alpha for smooth rendering
	return gl.accumulator / dt
}

// Play starts or resumes the game
func (gl *GameLoop) Play() {
	gl.State = StatePlaying
	gl.lastTime = time.Now()
}

// Pause pauses the game
func (gl *GameLoop) Pause() {
	gl.State = StatePaused
}

// CurrentTick returns the current simulation tick
func (gl *GameLoop) CurrentTick() uint64 {
	return gl.World.TickCount
}

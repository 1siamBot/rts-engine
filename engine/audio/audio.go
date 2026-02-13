package audio

import (
	"math"
)

// SoundID identifies a sound effect
type SoundID string

const (
	SndAttack    SoundID = "attack"
	SndExplosion SoundID = "explosion"
	SndSelect    SoundID = "select"
	SndMove      SoundID = "move"
	SndBuild     SoundID = "build"
	SndClick     SoundID = "click"
)

// AudioManager handles music and sound effects
// Uses Ebitengine's audio package internally
type AudioManager struct {
	MasterVolume float64
	MusicVolume  float64
	SFXVolume    float64
	MusicPlaying bool
	CameraX      float64
	CameraY      float64
}

func NewAudioManager() *AudioManager {
	return &AudioManager{
		MasterVolume: 1.0,
		MusicVolume:  0.5,
		SFXVolume:    0.8,
	}
}

// SetCameraPos updates the listener position for positional audio
func (am *AudioManager) SetCameraPos(x, y float64) {
	am.CameraX = x
	am.CameraY = y
}

// PlaySFX plays a sound effect at a world position
func (am *AudioManager) PlaySFX(id SoundID, worldX, worldY float64) {
	vol := am.calcVolume(worldX, worldY)
	_ = vol
	// In a real implementation, we'd load and play audio bytes via ebiten/audio
	// For now this is a stub that integrates into the architecture
}

// PlayMusic starts background music
func (am *AudioManager) PlayMusic(_ string) {
	am.MusicPlaying = true
	// Stub: would use ebiten/audio.Player
}

// StopMusic stops background music
func (am *AudioManager) StopMusic() {
	am.MusicPlaying = false
}

// calcVolume computes volume based on distance from camera
func (am *AudioManager) calcVolume(wx, wy float64) float64 {
	dx := wx - am.CameraX
	dy := wy - am.CameraY
	dist := math.Sqrt(dx*dx + dy*dy)
	maxDist := 30.0
	if dist >= maxDist {
		return 0
	}
	vol := (1.0 - dist/maxDist) * am.SFXVolume * am.MasterVolume
	return vol
}

// SetVolume sets master volume (0-1)
func (am *AudioManager) SetVolume(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	am.MasterVolume = v
}

package systems

import (
	"github.com/1siamBot/rts-engine/engine/core"
)

// AnimationSystem updates sprite animation frames
type AnimationSystem struct{}

func (s *AnimationSystem) Priority() int { return 60 }

func (s *AnimationSystem) Update(w *core.World, dt float64) {
	ids := w.Query(core.CompAnim, core.CompSprite)
	for _, id := range ids {
		anim := w.Get(id, core.CompAnim).(*core.AnimState)
		sprite := w.Get(id, core.CompSprite).(*core.Sprite)

		if anim.Finished || anim.Speed <= 0 {
			continue
		}

		anim.Timer += dt
		frameDur := 1.0 / anim.Speed
		if anim.Timer >= frameDur {
			anim.Timer -= frameDur
			anim.Frame++
			sprite.FrameX = anim.Frame

			// Simple loop/finish logic
			maxFrames := 8 // default max frames
			if anim.Frame >= maxFrames {
				if anim.Loop {
					anim.Frame = 0
					sprite.FrameX = 0
				} else {
					anim.Finished = true
					anim.Frame = maxFrames - 1
				}
			}
		}
	}
}

// VeterancySystem tracks unit kills and gives bonuses
type VeterancySystem struct{}

func (s *VeterancySystem) Priority() int { return 55 }

func (s *VeterancySystem) Update(_ *core.World, _ float64) {
	// Veterancy is tracked via events; this is a placeholder for tick-based checks
}

// GameOverSystem checks if any player has lost all buildings
type GameOverSystem struct {
	Players *core.PlayerManager
}

func (s *GameOverSystem) Priority() int { return 100 }

func (s *GameOverSystem) Update(w *core.World, _ float64) {
	buildings := w.Query(core.CompBuilding, core.CompOwner)
	hasBldg := make(map[int]bool)
	for _, id := range buildings {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		hasBldg[own.PlayerID] = true
	}
	for _, p := range s.Players.Players {
		if !p.Defeated && !hasBldg[p.ID] {
			// Check if they at least have units
			hasUnits := false
			for _, uid := range w.Query(core.CompOwner, core.CompMovable) {
				own := w.Get(uid, core.CompOwner).(*core.Owner)
				if own.PlayerID == p.ID {
					hasUnits = true
					break
				}
			}
			if !hasUnits {
				p.Defeated = true
			}
		}
	}
}

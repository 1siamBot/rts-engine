package systems

import (
	"math"

	"github.com/1siamBot/rts-engine/engine/core"
)

// ProjectileSystem moves projectiles and handles impact
type ProjectileSystem struct {
	EventBus *core.EventBus
}

func (s *ProjectileSystem) Priority() int { return 25 }

func (s *ProjectileSystem) Update(w *core.World, dt float64) {
	ids := w.Query(core.CompPosition, core.CompProjectile)
	for _, id := range ids {
		pos := w.Get(id, core.CompPosition).(*core.Position)
		proj := w.Get(id, core.CompProjectile).(*core.Projectile)

		// Update target position if target still alive
		if tpos := w.Get(proj.TargetID, core.CompPosition); tpos != nil {
			tp := tpos.(*core.Position)
			proj.TargetX = tp.X
			proj.TargetY = tp.Y
		}

		dx := proj.TargetX - pos.X
		dy := proj.TargetY - pos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < 0.3 {
			// Hit!
			if proj.Splash > 0 {
				// AoE damage
				allHP := w.Query(core.CompPosition, core.CompHealth)
				for _, tid := range allHP {
					if tid == id {
						continue
					}
					tp := w.Get(tid, core.CompPosition).(*core.Position)
					d := math.Sqrt(math.Pow(tp.X-pos.X, 2) + math.Pow(tp.Y-pos.Y, 2))
					if d <= proj.Splash {
						scale := 1.0 - d/proj.Splash
						dmg := int(float64(proj.Damage) * scale)
						if dmg < 1 {
							dmg = 1
						}
						ApplyDamage(w, tid, dmg, proj.DmgType, s.EventBus)
					}
				}
			} else {
				ApplyDamage(w, proj.TargetID, proj.Damage, proj.DmgType, s.EventBus)
			}
			if s.EventBus != nil {
				s.EventBus.Emit(core.Event{Type: core.EvtProjectileHit, Tick: w.TickCount})
			}
			w.Destroy(id)
			continue
		}

		// Move toward target
		speed := proj.Speed * dt
		pos.X += dx / dist * speed
		pos.Y += dy / dist * speed
		pos.Facing = math.Atan2(dy, dx)
	}
}

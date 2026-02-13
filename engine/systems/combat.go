package systems

import (
	"math"

	"github.com/1siamBot/rts-engine/engine/core"
)

// DamageMultiplier table: [DamageType][ArmorType] -> multiplier
var DamageMultiplier = [5][5]float64{
	// None   Light  Medium Heavy  Building
	{1.0, 1.0, 0.7, 0.4, 0.3},  // Kinetic
	{1.2, 0.8, 1.0, 1.2, 1.5},  // Explosive
	{1.5, 1.3, 0.9, 0.6, 0.8},  // Fire
	{1.0, 1.5, 1.2, 0.8, 0.5},  // Electric
	{1.3, 1.1, 1.1, 1.0, 1.0},  // Radiation
}

// CombatSystem processes weapon cooldowns and auto-attack
type CombatSystem struct {
	EventBus *core.EventBus
	Players  *core.PlayerManager
}

func (s *CombatSystem) Priority() int { return 20 }

func (s *CombatSystem) Update(w *core.World, dt float64) {
	attackers := w.Query(core.CompPosition, core.CompWeapon, core.CompOwner)
	targets := w.Query(core.CompPosition, core.CompHealth, core.CompOwner)

	for _, aid := range attackers {
		wep := w.Get(aid, core.CompWeapon).(*core.Weapon)
		// Cool down weapon
		if wep.CooldownNow > 0 {
			wep.CooldownNow -= dt
			continue
		}

		apos := w.Get(aid, core.CompPosition).(*core.Position)
		aown := w.Get(aid, core.CompOwner).(*core.Owner)

		// Find nearest enemy in range
		var bestID core.EntityID
		bestDist := math.MaxFloat64
		for _, tid := range targets {
			if tid == aid {
				continue
			}
			town := w.Get(tid, core.CompOwner).(*core.Owner)
			if s.Players.AreAllies(aown.PlayerID, town.PlayerID) {
				continue
			}
			tpos := w.Get(tid, core.CompPosition).(*core.Position)
			d := apos.DistanceTo(tpos)
			if d <= wep.Range && d < bestDist {
				bestDist = d
				bestID = tid
			}
		}
		if bestID == 0 {
			continue
		}

		// Fire
		wep.CooldownNow = wep.Cooldown
		tpos := w.Get(bestID, core.CompPosition).(*core.Position)

		if wep.Projectile != "" {
			// Spawn projectile entity
			pid := w.Spawn()
			w.Attach(pid, &core.Position{X: apos.X, Y: apos.Y})
			w.Attach(pid, &core.Projectile{
				SourceID: aid,
				TargetID: bestID,
				TargetX:  tpos.X,
				TargetY:  tpos.Y,
				Speed:    8.0,
				Damage:   wep.Damage,
				Splash:   wep.Splash,
				DmgType:  wep.DamageType,
				HitFX:    "explosion",
			})
		} else {
			// Hitscan: apply damage immediately
			ApplyDamage(w, bestID, wep.Damage, wep.DamageType, s.EventBus)
		}

		if s.EventBus != nil {
			s.EventBus.Emit(core.Event{Type: core.EvtUnitAttack, Tick: w.TickCount})
		}
	}
}

// ApplyDamage applies damage to an entity considering armor
func ApplyDamage(w *core.World, id core.EntityID, baseDamage int, dmgType core.DamageType, bus *core.EventBus) {
	hp := w.Get(id, core.CompHealth)
	if hp == nil {
		return
	}
	h := hp.(*core.Health)

	mult := 1.0
	if arm := w.Get(id, core.CompArmor); arm != nil {
		a := arm.(*core.Armor)
		if int(dmgType) < 5 && int(a.ArmorType) < 5 {
			mult = DamageMultiplier[dmgType][a.ArmorType]
		}
		baseDamage -= a.Value
		if baseDamage < 1 {
			baseDamage = 1
		}
	}

	finalDmg := int(float64(baseDamage) * mult)
	if finalDmg < 1 {
		finalDmg = 1
	}
	h.Current -= finalDmg

	if h.Current <= 0 {
		h.Current = 0
		w.Destroy(id)
		if bus != nil {
			bus.Emit(core.Event{Type: core.EvtUnitDestroyed, Tick: w.TickCount})
		}
	}
}

package render

import (
	"math"

	"github.com/1siamBot/rts-engine/engine/core"
	"github.com/hajimehoshi/ebiten/v2"
)

// DrawBuildingSprite draws a building using its sprite if available
// Returns true if a sprite was drawn, false to fall back to default rendering
func (r *IsoRenderer) DrawBuildingSprite(screen *ebiten.Image, w *core.World, id core.EntityID, sx, sy int) bool {
	bn := w.Get(id, core.CompBuildingName)
	if bn == nil {
		return false
	}
	key := bn.(*core.BuildingName).Key

	// Try faction-specific sprite first
	var sprite *ebiten.Image
	ownerComp := w.Get(id, core.CompOwner)
	if ownerComp != nil {
		owner := ownerComp.(*core.Owner)
		faction := owner.Faction
		if faction == "" {
			if owner.PlayerID == 0 {
				faction = "allied"
			} else {
				faction = "soviet"
			}
		}

		// Check for damaged state
		healthComp := w.Get(id, core.CompHealth)
		if healthComp != nil {
			h := healthComp.(*core.Health)
			if h.Ratio() < 0.5 {
				if dmgSprite, ok := r.Sprites.BuildingSprites[key+"_"+faction+"_damaged"]; ok {
					sprite = dmgSprite
				}
			}
		}

		// Check for construction state
		if sprite == nil {
			bcComp := w.Get(id, core.CompBuildingConstruction)
			if bcComp != nil {
				bc := bcComp.(*core.BuildingConstruction)
				if !bc.Complete {
					stage := int(bc.Progress * 3)
					if stage > 2 {
						stage = 2
					}
					buildKey := key + "_" + faction + "_build_" + string(rune('0'+stage))
					if buildSprite, ok := r.Sprites.BuildingSprites[buildKey]; ok {
						sprite = buildSprite
					}
				}
			}
		}

		// Regular faction sprite
		if sprite == nil {
			if factionSprite, ok := r.Sprites.BuildingSprites[key+"_"+faction]; ok {
				sprite = factionSprite
			}
		}
	}

	// Fallback to default sprite
	if sprite == nil {
		var ok bool
		sprite, ok = r.Sprites.BuildingSprites[key]
		if !ok {
			return false
		}
	}

	sw := sprite.Bounds().Dx()
	sh := sprite.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(sx-sw/2), float64(sy-sh/2))
	screen.DrawImage(sprite, op)
	return true
}

// facingToDirection converts a facing angle (radians) to 8-direction index
// 0=E, 1=SE, 2=S, 3=SW, 4=W, 5=NW, 6=N, 7=NE
func facingToDirection(facing float64) int {
	// Normalize to [0, 2π)
	for facing < 0 {
		facing += 2 * math.Pi
	}
	for facing >= 2*math.Pi {
		facing -= 2 * math.Pi
	}
	// Each direction covers π/4 radians
	dir := int(math.Round(facing/(math.Pi/4))) % 8
	return dir
}

// DrawUnitSprite draws a unit using its sprite if available
func (r *IsoRenderer) DrawUnitSprite(screen *ebiten.Image, w *core.World, id core.EntityID, sx, sy int, playerID int) bool {
	// Determine unit type
	var spriteKey string
	if w.Has(id, core.CompMCV) {
		spriteKey = "mcv"
	} else if w.Has(id, core.CompHarvester) {
		spriteKey = "harvester"
	} else if w.Has(id, core.CompWeapon) {
		spr := w.Get(id, core.CompSprite)
		if spr != nil {
			s := spr.(*core.Sprite)
			if s.Width > 26 {
				spriteKey = "tank"
			} else {
				spriteKey = "infantry"
			}
		} else {
			spriteKey = "infantry"
		}
	} else {
		spriteKey = "infantry"
	}

	// Get facing direction
	direction := 2 // default south
	animFrame := 0
	posComp := w.Get(id, core.CompPosition)
	if posComp != nil {
		pos := posComp.(*core.Position)
		direction = facingToDirection(pos.Facing)
	}

	// Get animation frame
	animComp := w.Get(id, core.CompAnim)
	if animComp != nil {
		anim := animComp.(*core.AnimState)
		animFrame = anim.Frame % 3
	}

	// Try directional sprite first
	sprite := r.Sprites.GetUnitDirectionalSprite(spriteKey, direction, animFrame)
	if sprite == nil {
		// Fallback to default
		var ok bool
		sprite, ok = r.Sprites.UnitSprites[spriteKey]
		if !ok {
			return false
		}
	}

	sw := sprite.Bounds().Dx()
	sh := sprite.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(sx-sw/2), float64(sy-sh/2))

	// Tint enemy units red
	if playerID != 0 {
		op.ColorScale.Scale(1.5, 0.6, 0.6, 1.0)
	}

	screen.DrawImage(sprite, op)
	return true
}

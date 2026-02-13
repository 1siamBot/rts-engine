package render

import (
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
	sprite, ok := r.Sprites.BuildingSprites[key]
	if !ok {
		return false
	}

	sw := sprite.Bounds().Dx()
	sh := sprite.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(sx-sw/2), float64(sy-sh/2))
	screen.DrawImage(sprite, op)
	return true
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
		// Has weapon but check sprite size to determine tank vs infantry
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

	sprite, ok := r.Sprites.UnitSprites[spriteKey]
	if !ok {
		return false
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

package ai

import (
	"math"
	"math/rand"

	"github.com/1siamBot/rts-engine/engine/core"
	"github.com/1siamBot/rts-engine/engine/pathfind"
	"github.com/1siamBot/rts-engine/engine/systems"
)

// Difficulty controls AI behavior
type Difficulty int

const (
	DiffEasy Difficulty = iota
	DiffMedium
	DiffHard
)

// AIController manages one AI player
type AIController struct {
	PlayerID   int
	Difficulty Difficulty
	TechTree   *systems.TechTree
	NavGrid    *pathfind.NavGrid

	tickTimer    float64
	thinkInterval float64
	attackTimer  float64
	builtBarracks bool
	builtFactory  bool
	waveCount    int
}

func NewAIController(playerID int, diff Difficulty, tt *systems.TechTree, ng *pathfind.NavGrid) *AIController {
	interval := 5.0
	switch diff {
	case DiffEasy:
		interval = 8.0
	case DiffHard:
		interval = 3.0
	}
	return &AIController{
		PlayerID:      playerID,
		Difficulty:    diff,
		TechTree:      tt,
		NavGrid:       ng,
		thinkInterval: interval,
	}
}

// AISystem runs all AI controllers
type AISystem struct {
	Controllers []*AIController
	Players     *core.PlayerManager
}

func (s *AISystem) Priority() int { return 50 }

func (s *AISystem) Update(w *core.World, dt float64) {
	for _, ai := range s.Controllers {
		ai.tickTimer += dt
		if ai.tickTimer >= ai.thinkInterval {
			ai.tickTimer = 0
			ai.Think(w, s.Players)
		}
		ai.attackTimer += dt
	}
}

// Think is the main AI decision loop
func (ai *AIController) Think(w *core.World, pm *core.PlayerManager) {
	player := pm.GetPlayer(ai.PlayerID)
	if player == nil || player.Defeated {
		return
	}

	// Count own buildings and units
	myBuildings := ai.countBuildings(w)
	myUnits := ai.countUnits(w)

	// Build order: power plant -> barracks -> refinery -> war factory -> produce units
	if myBuildings == 0 {
		return // no buildings, can't do anything
	}

	// Queue units from production buildings
	prodIDs := w.Query(core.CompProduction, core.CompOwner)
	for _, pid := range prodIDs {
		own := w.Get(pid, core.CompOwner).(*core.Owner)
		if own.PlayerID != ai.PlayerID {
			continue
		}
		prod := w.Get(pid, core.CompProduction).(*core.Production)
		if len(prod.Queue) > 0 {
			continue
		}

		// Pick a unit to build
		maxQueue := 2
		if ai.Difficulty == DiffHard {
			maxQueue = 3
		}
		if len(prod.Queue) >= maxQueue {
			continue
		}

		// Simple: build infantry if cheap, tanks if affordable
		unitType := "conscript"
		if player.Faction == "Allied" {
			unitType = "gi"
		}
		if player.Credits > 800 && myUnits > 3 {
			if player.Faction == "Allied" {
				unitType = "grizzly"
			} else {
				unitType = "rhino"
			}
		}
		if udef, ok := ai.TechTree.Units[unitType]; ok {
			if player.Credits >= udef.Cost {
				player.Credits -= udef.Cost
				prod.Queue = append(prod.Queue, unitType)
			}
		}
	}

	// Attack wave every 30-60 seconds depending on difficulty
	attackInterval := 60.0
	switch ai.Difficulty {
	case DiffMedium:
		attackInterval = 45.0
	case DiffHard:
		attackInterval = 30.0
	}

	if ai.attackTimer >= attackInterval && myUnits >= 3 {
		ai.attackTimer = 0
		ai.waveCount++
		ai.launchAttack(w, pm)
	}
}

func (ai *AIController) countBuildings(w *core.World) int {
	count := 0
	for _, id := range w.Query(core.CompBuilding, core.CompOwner) {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		if own.PlayerID == ai.PlayerID {
			count++
		}
	}
	return count
}

func (ai *AIController) countUnits(w *core.World) int {
	count := 0
	for _, id := range w.Query(core.CompMovable, core.CompOwner) {
		if !w.Has(id, core.CompBuilding) {
			own := w.Get(id, core.CompOwner).(*core.Owner)
			if own.PlayerID == ai.PlayerID {
				count++
			}
		}
	}
	return count
}

func (ai *AIController) launchAttack(w *core.World, pm *core.PlayerManager) {
	// Find enemy position (any enemy building or unit)
	var targetX, targetY float64
	found := false
	for _, id := range w.Query(core.CompPosition, core.CompOwner) {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		if own.PlayerID != ai.PlayerID && !pm.AreAllies(ai.PlayerID, own.PlayerID) {
			pos := w.Get(id, core.CompPosition).(*core.Position)
			targetX, targetY = pos.X, pos.Y
			found = true
			break
		}
	}
	if !found {
		return
	}

	// Send all combat units toward enemy
	gx, gy := int(targetX), int(targetY)
	for _, id := range w.Query(core.CompMovable, core.CompOwner, core.CompWeapon) {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		if own.PlayerID != ai.PlayerID {
			continue
		}
		// Add slight randomness to target
		ox := gx + rand.Intn(5) - 2
		oy := gy + rand.Intn(5) - 2
		systems.OrderMove(w, ai.NavGrid, id, ox, oy)
	}
}

// ThreatAssessment returns the total threat value of enemies near a position
func ThreatAssessment(w *core.World, pm *core.PlayerManager, playerID int, wx, wy, radius float64) float64 {
	threat := 0.0
	for _, id := range w.Query(core.CompPosition, core.CompWeapon, core.CompOwner) {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		if pm.AreAllies(playerID, own.PlayerID) {
			continue
		}
		pos := w.Get(id, core.CompPosition).(*core.Position)
		dx := pos.X - wx
		dy := pos.Y - wy
		d := math.Sqrt(dx*dx + dy*dy)
		if d <= radius {
			wep := w.Get(id, core.CompWeapon).(*core.Weapon)
			threat += float64(wep.Damage) * (1.0 - d/radius)
		}
	}
	return threat
}

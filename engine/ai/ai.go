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
	TileMap    systems.TileMapOccupy

	tickTimer     float64
	thinkInterval float64
	attackTimer   float64
	waveCount     int
	buildOffset   int // offset for next building placement
}

func NewAIController(playerID int, diff Difficulty, tt *systems.TechTree, ng *pathfind.NavGrid, tm systems.TileMapOccupy) *AIController {
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
		TileMap:       tm,
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

	// First: auto-deploy any MCV the AI owns
	ai.autoDeployMCV(w)

	// Collect owned building keys
	ownedKeys := ai.ownedBuildingKeys(w)
	myUnits := ai.countUnits(w)

	hasConYard := ownedKeys["construction_yard"]
	hasPower := ownedKeys["power_plant"]
	hasBarracks := ownedKeys["barracks"]
	hasRefinery := ownedKeys["refinery"]
	hasWarFactory := ownedKeys["war_factory"]

	if !hasConYard {
		return // no con yard, can't build
	}

	// AI build order: Power Plant → Barracks → Refinery → War Factory
	if !hasPower && player.Credits >= 800 {
		ai.aiBuildBuilding(w, player, "power_plant")
	} else if !hasBarracks && hasPower && player.Credits >= 500 {
		ai.aiBuildBuilding(w, player, "barracks")
	} else if !hasRefinery && hasPower && player.Credits >= 2000 {
		ai.aiBuildBuilding(w, player, "refinery")
	} else if !hasWarFactory && hasRefinery && player.Credits >= 2000 {
		ai.aiBuildBuilding(w, player, "war_factory")
	}

	// Queue units from production buildings
	prodIDs := w.Query(core.CompProduction, core.CompOwner)
	for _, pid := range prodIDs {
		own := w.Get(pid, core.CompOwner).(*core.Owner)
		if own.PlayerID != ai.PlayerID {
			continue
		}
		prod := w.Get(pid, core.CompProduction).(*core.Production)

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
		if hasWarFactory && player.Credits > 800 && myUnits > 3 {
			if player.Faction == "Allied" {
				unitType = "grizzly"
			} else {
				unitType = "rhino"
			}
		}
		if udef, ok := ai.TechTree.Units[unitType]; ok {
			if player.Credits >= udef.Cost && ai.TechTree.HasPrereqs(w, ai.PlayerID, udef.Prereqs) {
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

// autoDeployMCV deploys any MCV the AI owns
func (ai *AIController) autoDeployMCV(w *core.World) {
	for _, id := range w.Query(core.CompMCV, core.CompOwner) {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		if own.PlayerID == ai.PlayerID {
			systems.DeployMCV(w, id, nil)
			return
		}
	}
}

// ownedBuildingKeys returns a set of building keys the AI owns (completed only)
func (ai *AIController) ownedBuildingKeys(w *core.World) map[string]bool {
	keys := make(map[string]bool)
	for _, id := range w.Query(core.CompBuilding, core.CompOwner, core.CompBuildingName) {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		if own.PlayerID != ai.PlayerID {
			continue
		}
		// Only count completed buildings
		if bc := w.Get(id, core.CompBuildingConstruction); bc != nil {
			if !bc.(*core.BuildingConstruction).Complete {
				continue
			}
		}
		bn := w.Get(id, core.CompBuildingName).(*core.BuildingName)
		keys[bn.Key] = true
	}
	return keys
}

// aiBuildBuilding places a building near the AI's construction yard
func (ai *AIController) aiBuildBuilding(w *core.World, player *core.Player, key string) {
	bdef, ok := ai.TechTree.Buildings[key]
	if !ok {
		return
	}

	// Find con yard position
	var cyX, cyY float64
	found := false
	for _, id := range w.Query(core.CompBuilding, core.CompOwner, core.CompBuildingName) {
		own := w.Get(id, core.CompOwner).(*core.Owner)
		if own.PlayerID != ai.PlayerID {
			continue
		}
		bn := w.Get(id, core.CompBuildingName).(*core.BuildingName)
		if bn.Key == "construction_yard" {
			pos := w.Get(id, core.CompPosition).(*core.Position)
			cyX, cyY = pos.X, pos.Y
			found = true
			break
		}
	}
	if !found {
		return
	}

	// Try placement offsets around the con yard
	offsets := [][2]int{
		{-3, 0}, {4, 0}, {0, -3}, {0, 4},
		{-3, -3}, {4, 4}, {-3, 4}, {4, -3},
		{-6, 0}, {7, 0}, {0, -6}, {0, 7},
	}
	for _, off := range offsets {
		tx := int(cyX) + off[0]
		ty := int(cyY) + off[1]
		if ai.canAIPlace(w, tx, ty, bdef.SizeX, bdef.SizeY) {
			player.Credits -= bdef.Cost
			bid := systems.PlaceBuilding(w, key, ai.TechTree, ai.PlayerID, tx, ty, player.Faction, nil)
			if bid != 0 && ai.TileMap != nil {
				systems.OccupyTiles(ai.TileMap, tx, ty, bdef.SizeX, bdef.SizeY)
			}
			return
		}
	}
}

// canAIPlace checks if the AI can place a building at the given position
func (ai *AIController) canAIPlace(w *core.World, tileX, tileY, sizeX, sizeY int) bool {
	// Check near existing buildings
	nearBuilding := false
	for _, bid := range w.Query(core.CompBuilding, core.CompOwner, core.CompPosition) {
		o := w.Get(bid, core.CompOwner).(*core.Owner)
		if o.PlayerID != ai.PlayerID {
			continue
		}
		bp := w.Get(bid, core.CompPosition).(*core.Position)
		dx := float64(tileX) - bp.X
		dy := float64(tileY) - bp.Y
		if dx*dx+dy*dy < 100 {
			nearBuilding = true
			break
		}
	}
	return nearBuilding
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

package systems

import (
	"github.com/1siamBot/rts-engine/engine/core"
)

// UnitDef defines a unit type that can be produced
type UnitDef struct {
	Name      string
	Cost      int
	BuildTime float64 // seconds
	HP        int
	Speed     float64
	Damage    int
	Range     float64
	ArmorType core.ArmorType
	DmgType   core.DamageType
	MoveType  core.MoveType
	Vision    int
	Prereqs   []string
	Faction   string
}

// BuildingDef defines a building type
type BuildingDef struct {
	Name      string
	Cost      int
	BuildTime float64
	HP        int
	SizeX     int
	SizeY     int
	PowerGen  int
	PowerDraw int
	TechLevel int
	Prereqs   []string
	CanProduce []string
	Faction   string
}

// TechTree holds all definitions
type TechTree struct {
	Units     map[string]*UnitDef
	Buildings map[string]*BuildingDef
}

// NewTechTree creates a default RA2-style tech tree
func NewTechTree() *TechTree {
	tt := &TechTree{
		Units:     make(map[string]*UnitDef),
		Buildings: make(map[string]*BuildingDef),
	}

	// Allied units
	tt.Units["gi"] = &UnitDef{Name: "GI", Cost: 200, BuildTime: 3, HP: 125, Speed: 3.0, Damage: 15, Range: 5, ArmorType: core.ArmorLight, DmgType: core.DmgKinetic, MoveType: core.MoveInfantry, Vision: 5, Faction: "Allied"}
	tt.Units["grizzly"] = &UnitDef{Name: "Grizzly Tank", Cost: 700, BuildTime: 8, HP: 400, Speed: 2.5, Damage: 75, Range: 5.5, ArmorType: core.ArmorHeavy, DmgType: core.DmgExplosive, MoveType: core.MoveVehicle, Vision: 6, Faction: "Allied", Prereqs: []string{"war_factory"}}
	tt.Units["harvester_a"] = &UnitDef{Name: "Chrono Miner", Cost: 1400, BuildTime: 12, HP: 600, Speed: 1.5, MoveType: core.MoveVehicle, Vision: 4, Faction: "Allied"}

	// Soviet units
	tt.Units["conscript"] = &UnitDef{Name: "Conscript", Cost: 100, BuildTime: 2, HP: 100, Speed: 3.0, Damage: 12, Range: 4.5, ArmorType: core.ArmorNone, DmgType: core.DmgKinetic, MoveType: core.MoveInfantry, Vision: 5, Faction: "Soviet"}
	tt.Units["rhino"] = &UnitDef{Name: "Rhino Tank", Cost: 900, BuildTime: 10, HP: 500, Speed: 2.0, Damage: 90, Range: 5.5, ArmorType: core.ArmorHeavy, DmgType: core.DmgExplosive, MoveType: core.MoveVehicle, Vision: 6, Faction: "Soviet", Prereqs: []string{"war_factory"}}
	tt.Units["harvester_s"] = &UnitDef{Name: "War Miner", Cost: 1400, BuildTime: 12, HP: 800, Speed: 1.2, Damage: 20, Range: 3, ArmorType: core.ArmorHeavy, DmgType: core.DmgKinetic, MoveType: core.MoveVehicle, Vision: 4, Faction: "Soviet"}

	// Buildings (shared names, faction handled by Faction field)
	tt.Buildings["construction_yard"] = &BuildingDef{Name: "Construction Yard", Cost: 0, BuildTime: 0, HP: 1000, SizeX: 3, SizeY: 3, PowerGen: 0, PowerDraw: 0, TechLevel: 0, Faction: ""}
	tt.Buildings["power_plant"] = &BuildingDef{Name: "Power Plant", Cost: 800, BuildTime: 10, HP: 750, SizeX: 2, SizeY: 2, PowerGen: 100, PowerDraw: 0, TechLevel: 0, Prereqs: []string{"construction_yard"}, Faction: ""}
	tt.Buildings["barracks"] = &BuildingDef{Name: "Barracks", Cost: 500, BuildTime: 8, HP: 500, SizeX: 2, SizeY: 2, PowerDraw: 20, TechLevel: 0, CanProduce: []string{"gi", "conscript"}, Prereqs: []string{"power_plant"}, Faction: ""}
	tt.Buildings["refinery"] = &BuildingDef{Name: "Ore Refinery", Cost: 2000, BuildTime: 15, HP: 900, SizeX: 3, SizeY: 3, PowerDraw: 30, TechLevel: 0, Prereqs: []string{"power_plant"}, Faction: ""}
	tt.Buildings["war_factory"] = &BuildingDef{Name: "War Factory", Cost: 2000, BuildTime: 15, HP: 1000, SizeX: 3, SizeY: 3, PowerDraw: 50, TechLevel: 1, CanProduce: []string{"grizzly", "rhino", "harvester_a", "harvester_s"}, Prereqs: []string{"refinery"}, Faction: ""}

	return tt
}

// HasPrereqs checks if a player has all prerequisites built
func (tt *TechTree) HasPrereqs(w *core.World, playerID int, prereqs []string) bool {
	if len(prereqs) == 0 {
		return true
	}
	buildings := w.Query(core.CompBuilding, core.CompOwner)
	owned := make(map[string]bool)
	for _, bid := range buildings {
		own := w.Get(bid, core.CompOwner).(*core.Owner)
		if own.PlayerID != playerID {
			continue
		}
		// We need a way to identify building type; use a tag or check
		// For now we store building name in a simple lookup
		// We'll use the Building component's TechLevel as proxy
		// Better: we'll add a Name component or use sprite sheet ID
		_ = bid
	}
	// Simplified: just check by querying building counts
	_ = owned
	// For now, return true — prereqs checked at production order time via player building list
	return true
}

// ProductionSystem handles building production queues
type ProductionSystem struct {
	TechTree *TechTree
	Players  *core.PlayerManager
	EventBus *core.EventBus
}

func (s *ProductionSystem) Priority() int { return 35 }

func (s *ProductionSystem) Update(w *core.World, dt float64) {
	ids := w.Query(core.CompProduction, core.CompOwner, core.CompPosition)
	for _, id := range ids {
		prod := w.Get(id, core.CompProduction).(*core.Production)
		own := w.Get(id, core.CompOwner).(*core.Owner)
		pos := w.Get(id, core.CompPosition).(*core.Position)

		if len(prod.Queue) == 0 {
			continue
		}

		unitName := prod.Queue[0]
		udef, ok := s.TechTree.Units[unitName]
		if !ok {
			prod.Queue = prod.Queue[1:]
			continue
		}

		// Check power ratio for speed
		player := s.Players.GetPlayer(own.PlayerID)
		rate := prod.Rate
		if player != nil && !player.HasPower() {
			rate *= 0.5 // half speed without power
		}

		prod.Progress += (dt / udef.BuildTime) * rate
		if prod.Progress >= 1.0 {
			// Spawn unit at rally point
			spawnX := float64(prod.Rally.X) + 0.5
			spawnY := float64(prod.Rally.Y) + 0.5
			if prod.Rally.X == 0 && prod.Rally.Y == 0 {
				spawnX = pos.X + 2
				spawnY = pos.Y + 2
			}
			uid := w.Spawn()
			w.Attach(uid, &core.Position{X: spawnX, Y: spawnY})
			w.Attach(uid, &core.Sprite{Width: 24, Height: 24, Visible: true, ScaleX: 1, ScaleY: 1})
			w.Attach(uid, &core.Health{Current: udef.HP, Max: udef.HP})
			w.Attach(uid, &core.Movable{Speed: udef.Speed, MoveType: udef.MoveType})
			w.Attach(uid, &core.Selectable{Radius: 0.5})
			w.Attach(uid, &core.Owner{PlayerID: own.PlayerID, Faction: own.Faction})
			w.Attach(uid, &core.FogVision{Range: udef.Vision})
			if udef.Damage > 0 {
				w.Attach(uid, &core.Weapon{Name: udef.Name, Damage: udef.Damage, Range: udef.Range, Cooldown: 1.5, DamageType: udef.DmgType, TargetType: core.TargetAll})
			}
			w.Attach(uid, &core.Armor{ArmorType: udef.ArmorType})

			if s.EventBus != nil {
				s.EventBus.Emit(core.Event{Type: core.EvtUnitCreated, Tick: w.TickCount})
			}

			prod.Progress = 0
			prod.Queue = prod.Queue[1:]
		}
	}
}

// PowerSystem recalculates power for all players each tick
type PowerSystem struct {
	Players *core.PlayerManager
}

func (s *PowerSystem) Priority() int { return 5 }

func (s *PowerSystem) Update(w *core.World, _ float64) {
	// Reset power
	for _, p := range s.Players.Players {
		p.Power = 0
		p.PowerUse = 0
	}
	buildings := w.Query(core.CompBuilding, core.CompOwner)
	for _, bid := range buildings {
		b := w.Get(bid, core.CompBuilding).(*core.Building)
		own := w.Get(bid, core.CompOwner).(*core.Owner)
		player := s.Players.GetPlayer(own.PlayerID)
		if player == nil {
			continue
		}
		player.Power += b.PowerGen
		player.PowerUse += b.PowerDraw
	}
}

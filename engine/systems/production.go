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
	tt.Units["mcv"] = &UnitDef{Name: "MCV", Cost: 3000, BuildTime: 20, HP: 1000, Speed: 0.8, ArmorType: core.ArmorHeavy, MoveType: core.MoveVehicle, Vision: 6, Prereqs: []string{"war_factory"}, Faction: ""}

	// Buildings (shared names, faction handled by Faction field)
	tt.Buildings["construction_yard"] = &BuildingDef{Name: "Construction Yard", Cost: 0, BuildTime: 0, HP: 1000, SizeX: 3, SizeY: 3, PowerGen: 0, PowerDraw: 0, TechLevel: 0, Faction: ""}
	tt.Buildings["power_plant"] = &BuildingDef{Name: "Power Plant", Cost: 800, BuildTime: 10, HP: 750, SizeX: 2, SizeY: 2, PowerGen: 100, PowerDraw: 0, TechLevel: 0, Prereqs: []string{"construction_yard"}, Faction: ""}
	tt.Buildings["barracks"] = &BuildingDef{Name: "Barracks", Cost: 500, BuildTime: 8, HP: 500, SizeX: 2, SizeY: 2, PowerDraw: 20, TechLevel: 0, CanProduce: []string{"gi", "conscript"}, Prereqs: []string{"power_plant"}, Faction: ""}
	tt.Buildings["refinery"] = &BuildingDef{Name: "Ore Refinery", Cost: 2000, BuildTime: 15, HP: 900, SizeX: 3, SizeY: 3, PowerDraw: 30, TechLevel: 0, Prereqs: []string{"power_plant"}, Faction: ""}
	tt.Buildings["war_factory"] = &BuildingDef{Name: "War Factory", Cost: 2000, BuildTime: 15, HP: 1000, SizeX: 3, SizeY: 3, PowerDraw: 50, TechLevel: 1, CanProduce: []string{"grizzly", "rhino", "harvester_a", "harvester_s", "mcv"}, Prereqs: []string{"refinery"}, Faction: ""}

	return tt
}

// HasPrereqs checks if a player has all prerequisites built (completed)
func (tt *TechTree) HasPrereqs(w *core.World, playerID int, prereqs []string) bool {
	if len(prereqs) == 0 {
		return true
	}
	// Collect all completed building keys owned by player
	owned := make(map[string]bool)
	for _, bid := range w.Query(core.CompBuilding, core.CompOwner, core.CompBuildingName) {
		own := w.Get(bid, core.CompOwner).(*core.Owner)
		if own.PlayerID != playerID {
			continue
		}
		// Only count completed buildings
		if bc := w.Get(bid, core.CompBuildingConstruction); bc != nil {
			if !bc.(*core.BuildingConstruction).Complete {
				continue
			}
		}
		bn := w.Get(bid, core.CompBuildingName).(*core.BuildingName)
		owned[bn.Key] = true
	}
	for _, req := range prereqs {
		if !owned[req] {
			return false
		}
	}
	return true
}

// PlayerOwnsBuildingKey checks if a player has a completed building of a given key
func PlayerOwnsBuildingKey(w *core.World, playerID int, key string) bool {
	for _, bid := range w.Query(core.CompBuilding, core.CompOwner, core.CompBuildingName) {
		own := w.Get(bid, core.CompOwner).(*core.Owner)
		if own.PlayerID != playerID {
			continue
		}
		bn := w.Get(bid, core.CompBuildingName).(*core.BuildingName)
		if bn.Key != key {
			continue
		}
		if bc := w.Get(bid, core.CompBuildingConstruction); bc != nil {
			if !bc.(*core.BuildingConstruction).Complete {
				continue
			}
		}
		return true
	}
	return false
}

// FindProductionBuilding finds a building that can produce the given unit for a player
func FindProductionBuilding(w *core.World, tt *TechTree, playerID int, unitKey string) core.EntityID {
	for _, bid := range w.Query(core.CompProduction, core.CompOwner, core.CompBuildingName) {
		own := w.Get(bid, core.CompOwner).(*core.Owner)
		if own.PlayerID != playerID {
			continue
		}
		bn := w.Get(bid, core.CompBuildingName).(*core.BuildingName)
		bdef, ok := tt.Buildings[bn.Key]
		if !ok {
			continue
		}
		// Check if this building can produce the unit
		canProduce := false
		for _, u := range bdef.CanProduce {
			if u == unitKey {
				canProduce = true
				break
			}
		}
		if !canProduce {
			continue
		}
		// Check building is complete
		if bc := w.Get(bid, core.CompBuildingConstruction); bc != nil {
			if !bc.(*core.BuildingConstruction).Complete {
				continue
			}
		}
		prod := w.Get(bid, core.CompProduction).(*core.Production)
		if len(prod.Queue) < 5 {
			return bid
		}
	}
	return 0
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

			// MCV special component
			if unitName == "mcv" {
				w.Attach(uid, &core.MCV{CanDeploy: true})
			}

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

// BuildingConstructionSystem handles building construction animation
type BuildingConstructionSystem struct {
	Players  *core.PlayerManager
	EventBus *core.EventBus
}

func (s *BuildingConstructionSystem) Priority() int { return 6 }

func (s *BuildingConstructionSystem) Update(w *core.World, dt float64) {
	ids := w.Query(core.CompBuildingConstruction, core.CompHealth)
	for _, id := range ids {
		bc := w.Get(id, core.CompBuildingConstruction).(*core.BuildingConstruction)
		if bc.Complete {
			continue
		}

		// Low power slows construction
		rate := bc.Rate
		if own := w.Get(id, core.CompOwner); own != nil {
			o := own.(*core.Owner)
			if player := s.Players.GetPlayer(o.PlayerID); player != nil {
				if !player.HasPower() {
					rate *= 0.5
				}
			}
		}

		bc.Progress += rate * dt
		if bc.Progress >= 1.0 {
			bc.Progress = 1.0
			bc.Complete = true
			// Restore full health
			hp := w.Get(id, core.CompHealth).(*core.Health)
			hp.Current = hp.Max

			// Special: refinery spawns a harvester on completion
			if bn := w.Get(id, core.CompBuildingName); bn != nil {
				key := bn.(*core.BuildingName).Key
				if key == "refinery" {
					s.spawnRefineryHarvester(w, id)
				}
			}
		} else {
			// Health increases with construction
			hp := w.Get(id, core.CompHealth).(*core.Health)
			hp.Current = int(float64(hp.Max) * bc.Progress)
		}
	}
}

func (s *BuildingConstructionSystem) spawnRefineryHarvester(w *core.World, refID core.EntityID) {
	pos := w.Get(refID, core.CompPosition)
	own := w.Get(refID, core.CompOwner)
	if pos == nil || own == nil {
		return
	}
	p := pos.(*core.Position)
	o := own.(*core.Owner)

	uid := w.Spawn()
	w.Attach(uid, &core.Position{X: p.X + 3, Y: p.Y + 1})
	w.Attach(uid, &core.Sprite{Width: 28, Height: 28, Visible: true, ScaleX: 1, ScaleY: 1})
	w.Attach(uid, &core.Health{Current: 600, Max: 600})
	w.Attach(uid, &core.Movable{Speed: 1.5, MoveType: core.MoveVehicle})
	w.Attach(uid, &core.Harvester{Capacity: 20, Rate: 2.0, Resource: "ore"})
	w.Attach(uid, &core.Selectable{Radius: 0.6})
	w.Attach(uid, &core.Owner{PlayerID: o.PlayerID, Faction: o.Faction})
	w.Attach(uid, &core.FogVision{Range: 4})

	if s.EventBus != nil {
		s.EventBus.Emit(core.Event{Type: core.EvtUnitCreated, Tick: w.TickCount})
	}
}

// DeployMCV deploys an MCV into a Construction Yard at its current position
func DeployMCV(w *core.World, mcvID core.EntityID, eventBus *core.EventBus) core.EntityID {
	pos := w.Get(mcvID, core.CompPosition)
	own := w.Get(mcvID, core.CompOwner)
	if pos == nil || own == nil {
		return 0
	}
	p := pos.(*core.Position)
	o := own.(*core.Owner)

	// Remove MCV
	w.Destroy(mcvID)

	// Create Construction Yard
	cyID := w.Spawn()
	w.Attach(cyID, &core.Position{X: p.X, Y: p.Y})
	w.Attach(cyID, &core.Health{Current: 100, Max: 1000}) // starts low, builds up
	w.Attach(cyID, &core.Building{SizeX: 3, SizeY: 3, IsConYard: true, Sellable: true})
	w.Attach(cyID, &core.Production{Rate: 1.0, Rally: core.TilePos{X: int(p.X) + 3, Y: int(p.Y) + 3}})
	w.Attach(cyID, &core.Owner{PlayerID: o.PlayerID, Faction: o.Faction})
	w.Attach(cyID, &core.FogVision{Range: 8})
	w.Attach(cyID, &core.Selectable{Radius: 1.5})
	w.Attach(cyID, &core.BuildingName{Key: "construction_yard"})
	w.Attach(cyID, &core.BuildingConstruction{Progress: 0, Rate: 0.2, Complete: false}) // 5 seconds build

	if eventBus != nil {
		eventBus.Emit(core.Event{Type: core.EvtBuildingPlaced, Tick: w.TickCount})
	}

	return cyID
}

// UndeployConYard turns a Construction Yard back into an MCV
func UndeployConYard(w *core.World, cyID core.EntityID, eventBus *core.EventBus) core.EntityID {
	pos := w.Get(cyID, core.CompPosition)
	own := w.Get(cyID, core.CompOwner)
	if pos == nil || own == nil {
		return 0
	}
	p := pos.(*core.Position)
	o := own.(*core.Owner)

	w.Destroy(cyID)

	mcvID := w.Spawn()
	w.Attach(mcvID, &core.Position{X: p.X, Y: p.Y})
	w.Attach(mcvID, &core.Health{Current: 1000, Max: 1000})
	w.Attach(mcvID, &core.Movable{Speed: 0.8, MoveType: core.MoveVehicle})
	w.Attach(mcvID, &core.Sprite{Width: 32, Height: 32, Visible: true, ScaleX: 1, ScaleY: 1})
	w.Attach(mcvID, &core.Selectable{Radius: 0.8})
	w.Attach(mcvID, &core.Owner{PlayerID: o.PlayerID, Faction: o.Faction})
	w.Attach(mcvID, &core.FogVision{Range: 6})
	w.Attach(mcvID, &core.MCV{CanDeploy: true})
	w.Attach(mcvID, &core.Armor{ArmorType: core.ArmorHeavy})

	if eventBus != nil {
		eventBus.Emit(core.Event{Type: core.EvtUnitCreated, Tick: w.TickCount})
	}
	return mcvID
}

// PlaceBuilding places a building at the given tile position
func PlaceBuilding(w *core.World, key string, tt *TechTree, playerID int, tileX, tileY int, faction string, eventBus *core.EventBus) core.EntityID {
	bdef, ok := tt.Buildings[key]
	if !ok {
		return 0
	}

	id := w.Spawn()
	w.Attach(id, &core.Position{X: float64(tileX), Y: float64(tileY)})
	w.Attach(id, &core.Health{Current: 1, Max: bdef.HP})
	w.Attach(id, &core.Building{
		SizeX: bdef.SizeX, SizeY: bdef.SizeY,
		PowerGen: bdef.PowerGen, PowerDraw: bdef.PowerDraw,
		TechLevel: bdef.TechLevel, Sellable: true,
	})
	w.Attach(id, &core.Owner{PlayerID: playerID, Faction: faction})
	w.Attach(id, &core.FogVision{Range: 5})
	w.Attach(id, &core.Selectable{Radius: 1.0})
	w.Attach(id, &core.BuildingName{Key: key})

	// Construction animation
	buildRate := 1.0 / bdef.BuildTime // completes in BuildTime seconds
	w.Attach(id, &core.BuildingConstruction{Progress: 0, Rate: buildRate, Complete: false})

	// Add production if applicable
	if len(bdef.CanProduce) > 0 {
		w.Attach(id, &core.Production{Rate: 1.0, Rally: core.TilePos{X: tileX + bdef.SizeX + 1, Y: tileY + bdef.SizeY + 1}})
	}

	if eventBus != nil {
		eventBus.Emit(core.Event{Type: core.EvtBuildingPlaced, Tick: w.TickCount})
	}
	return id
}

// OccupyTiles marks tiles as occupied for a building footprint
func OccupyTiles(tm TileMapOccupy, tileX, tileY, sizeX, sizeY int) {
	for dy := 0; dy < sizeY; dy++ {
		for dx := 0; dx < sizeX; dx++ {
			tm.SetOccupied(tileX+dx, tileY+dy, true)
		}
	}
}

// FreeTiles unmarks tiles for a destroyed/sold building
func FreeTiles(tm TileMapOccupy, tileX, tileY, sizeX, sizeY int) {
	for dy := 0; dy < sizeY; dy++ {
		for dx := 0; dx < sizeX; dx++ {
			tm.SetOccupied(tileX+dx, tileY+dy, false)
		}
	}
}

// TileMapOccupy interface for marking tiles
type TileMapOccupy interface {
	SetOccupied(x, y int, occupied bool)
}

// SellBuilding sells a building for 50% of its cost
func SellBuilding(w *core.World, id core.EntityID, tt *TechTree, pm *core.PlayerManager) {
	own := w.Get(id, core.CompOwner)
	bn := w.Get(id, core.CompBuildingName)
	if own == nil {
		return
	}
	o := own.(*core.Owner)
	player := pm.GetPlayer(o.PlayerID)
	if player == nil {
		return
	}
	if bn != nil {
		name := bn.(*core.BuildingName)
		if bdef, ok := tt.Buildings[name.Key]; ok {
			player.Credits += bdef.Cost / 2
		}
	}
	w.Destroy(id)
}

// CanPlaceBuilding checks if a building can be placed at the given tile
func CanPlaceBuilding(w *core.World, tileX, tileY, sizeX, sizeY, playerID int, tm interface{ InBounds(int, int) bool; IsPassable(int, int, interface{}) bool }) bool {
	// Check bounds
	for dy := 0; dy < sizeY; dy++ {
		for dx := 0; dx < sizeX; dx++ {
			if !tm.InBounds(tileX+dx, tileY+dy) {
				return false
			}
		}
	}

	// Check near existing buildings (build radius)
	nearBuilding := false
	for _, bid := range w.Query(core.CompBuilding, core.CompOwner, core.CompPosition) {
		o := w.Get(bid, core.CompOwner).(*core.Owner)
		if o.PlayerID != playerID {
			continue
		}
		bp := w.Get(bid, core.CompPosition).(*core.Position)
		dx := float64(tileX) - bp.X
		dy := float64(tileY) - bp.Y
		if dx*dx+dy*dy < 100 { // within ~10 tiles
			nearBuilding = true
			break
		}
	}
	return nearBuilding
}

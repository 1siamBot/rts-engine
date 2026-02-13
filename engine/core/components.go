package core

import "math"

// ---- Position & Transform ----

// Position represents a world position in isometric space
type Position struct {
	X, Y   float64 // world position (tile coords, fractional)
	Z      float64 // height (for elevation, flying units)
	Facing float64 // direction in radians (0 = east, π/2 = south)
}

func (p *Position) Type() ComponentType { return CompPosition }

// DistanceTo returns euclidean distance to another position
func (p *Position) DistanceTo(other *Position) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// AngleTo returns the angle from this position to another
func (p *Position) AngleTo(other *Position) float64 {
	return math.Atan2(other.Y-p.Y, other.X-p.X)
}

// ---- Sprite & Animation ----

// Sprite represents rendering info
type Sprite struct {
	SheetID    string  // sprite sheet identifier
	FrameX     int     // current frame column
	FrameY     int     // current frame row (direction)
	Width      int     // frame width in pixels
	Height     int     // frame height in pixels
	OffsetX    int     // render offset X
	OffsetY    int     // render offset Y
	ScaleX     float64 // horizontal scale
	ScaleY     float64 // vertical scale
	Visible    bool
	ZOrder     int // rendering layer order
}

func (s *Sprite) Type() ComponentType { return CompSprite }

// AnimState represents animation state
type AnimState struct {
	CurrentAnim string  // animation name
	Frame       int     // current frame index
	Timer       float64 // time accumulator
	Speed       float64 // frames per second
	Loop        bool
	Finished    bool
}

func (a *AnimState) Type() ComponentType { return CompAnim }

// ---- Health & Combat ----

// Health represents hit points
type Health struct {
	Current int
	Max     int
}

func (h *Health) Type() ComponentType { return CompHealth }

func (h *Health) Ratio() float64 {
	if h.Max <= 0 {
		return 0
	}
	return float64(h.Current) / float64(h.Max)
}

// Weapon represents attack capability
type Weapon struct {
	Name        string
	Damage      int
	Range       float64 // in tile units
	Cooldown    float64 // seconds between shots
	CooldownNow float64
	Projectile  string  // projectile type (or "" for hitscan)
	Splash      float64 // AoE radius (0 = single target)
	DamageType  DamageType
	TargetType  TargetMask // what can this weapon target
}

func (w *Weapon) Type() ComponentType { return CompWeapon }

type DamageType uint8

const (
	DmgKinetic DamageType = iota
	DmgExplosive
	DmgFire
	DmgElectric
	DmgRadiation
)

type TargetMask uint8

const (
	TargetGround TargetMask = 1 << iota
	TargetAir
	TargetNaval
	TargetBuilding
	TargetAll TargetMask = 0xFF
)

// Armor represents defensive stats
type Armor struct {
	ArmorType ArmorType
	Value     int
}

func (a *Armor) Type() ComponentType { return CompArmor }

type ArmorType uint8

const (
	ArmorNone ArmorType = iota
	ArmorLight
	ArmorMedium
	ArmorHeavy
	ArmorBuilding
)

// ---- Movement ----

// Movable represents movement capability
type Movable struct {
	Speed    float64   // tiles per second
	TurnRate float64   // radians per second
	Path     []TilePos // current path
	PathIdx  int       // current position in path
	MoveType MoveType
}

func (m *Movable) Type() ComponentType { return CompMovable }

type MoveType uint8

const (
	MoveInfantry MoveType = iota
	MoveVehicle
	MoveNaval
	MoveAmphibious
	MoveAir
)

// TilePos represents integer tile coordinates
type TilePos struct {
	X, Y int
}

// ---- Selection ----

// Selectable marks an entity as selectable by player
type Selectable struct {
	Selected bool
	Radius   float64 // selection hitbox radius
	Group    int     // control group (0 = none, 1-9)
}

func (s *Selectable) Type() ComponentType { return CompSelectable }

// ---- Ownership ----

// Owner identifies which player owns this entity
type Owner struct {
	PlayerID int
	TeamID   int
	Faction  string
}

func (o *Owner) Type() ComponentType { return CompOwner }

// ---- Production ----

// Production represents a building that can produce units
type Production struct {
	Queue    []string // unit type names in queue
	Progress float64  // 0.0 to 1.0
	Rate     float64  // production speed multiplier
	Rally    TilePos  // rally point
}

func (p *Production) Type() ComponentType { return CompProduction }

// ---- Building ----

// Building represents a structure
type Building struct {
	SizeX, SizeY int    // footprint in tiles
	BuildTime    float64
	Powered      bool   // needs power?
	PowerDraw    int    // power consumption
	PowerGen     int    // power generation
	TechLevel    int
	Prereqs      []string // required buildings
	IsConYard    bool     // is this a Construction Yard?
	Sellable     bool     // can be sold for 50% refund
}

func (b *Building) Type() ComponentType { return CompBuilding }

// ---- MCV (Mobile Construction Vehicle) ----

// MCV marks a unit as deployable into a Construction Yard
type MCV struct {
	CanDeploy bool
}

func (m *MCV) Type() ComponentType { return CompMCV }

// ---- Building Construction Progress ----

// BuildingConstruction tracks construction animation progress
type BuildingConstruction struct {
	Progress float64 // 0.0 to 1.0
	Rate     float64 // progress per second
	Complete bool
}

func (bc *BuildingConstruction) Type() ComponentType { return CompBuildingConstruction }

// ---- Building Name Tag ----

// BuildingName stores the tech-tree key for a building
type BuildingName struct {
	Key string
}

func (bn *BuildingName) Type() ComponentType { return CompBuildingName }

// ---- Harvester ----

// Harvester represents a resource-gathering unit
type Harvester struct {
	Capacity int
	Current  int
	Rate     float64 // harvest speed
	Resource string  // "ore" or "gem"
	State    HarvesterState
}

func (h *Harvester) Type() ComponentType { return CompHarvester }

type HarvesterState uint8

const (
	HarvIdle HarvesterState = iota
	HarvMovingToOre
	HarvHarvesting
	HarvReturning
	HarvUnloading
)

// ---- Projectile ----

// Projectile represents a moving bullet/missile
type Projectile struct {
	SourceID EntityID
	TargetID EntityID
	TargetX  float64
	TargetY  float64
	Speed    float64
	Damage   int
	Splash   float64
	DmgType  DamageType
	TrailFX  string
	HitFX    string
}

func (p *Projectile) Type() ComponentType { return CompProjectile }

// ---- Fog of War Vision ----

// FogVision represents sight range
type FogVision struct {
	Range   int  // sight range in tiles
	Stealth bool // can this unit cloak?
	Detect  bool // can detect stealth?
}

func (f *FogVision) Type() ComponentType { return CompFogVision }

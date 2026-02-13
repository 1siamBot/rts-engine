package core

// Player represents a game player
type Player struct {
	ID       int
	Name     string
	TeamID   int
	Faction  string
	Color    uint32 // RGBA
	Credits  int    // money
	Power    int    // current power generation
	PowerUse int    // current power consumption
	IsAI     bool
	Defeated bool
}

// PowerRatio returns the power ratio (>= 1.0 means enough power)
func (p *Player) PowerRatio() float64 {
	if p.PowerUse <= 0 {
		return 1.0
	}
	return float64(p.Power) / float64(p.PowerUse)
}

// HasPower returns true if power is sufficient
func (p *Player) HasPower() bool {
	return p.Power >= p.PowerUse
}

// PlayerManager manages all players in a game
type PlayerManager struct {
	Players []*Player
}

func NewPlayerManager() *PlayerManager {
	return &PlayerManager{}
}

func (pm *PlayerManager) AddPlayer(p *Player) {
	pm.Players = append(pm.Players, p)
}

func (pm *PlayerManager) GetPlayer(id int) *Player {
	for _, p := range pm.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// AreAllies checks if two players are allied
func (pm *PlayerManager) AreAllies(a, b int) bool {
	pa := pm.GetPlayer(a)
	pb := pm.GetPlayer(b)
	if pa == nil || pb == nil {
		return false
	}
	return pa.TeamID == pb.TeamID
}

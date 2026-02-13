package core

import "sync/atomic"

// EntityID is a unique identifier for game entities
type EntityID uint64

var entityCounter uint64

// NewEntityID generates a unique entity ID
func NewEntityID() EntityID {
	return EntityID(atomic.AddUint64(&entityCounter, 1))
}

// Component is a marker interface for all components
type Component interface {
	Type() ComponentType
}

// ComponentType identifies the type of component
type ComponentType uint32

const (
	CompPosition ComponentType = iota
	CompSprite
	CompHealth
	CompMovable
	CompSelectable
	CompWeapon
	CompArmor
	CompProduction
	CompHarvester
	CompBuilding
	CompProjectile
	CompOwner
	CompAI
	CompFogVision
	CompAnim
	CompAudio
	CompMax
)

// World holds all entities and their components
type World struct {
	entities   map[EntityID]map[ComponentType]Component
	systems    []System
	toRemove   []EntityID
	TickCount  uint64
	TickRate   float64 // ticks per second (for deterministic lockstep)
}

// System processes entities each tick
type System interface {
	Update(w *World, dt float64)
	Priority() int
}

// NewWorld creates a new ECS world
func NewWorld(tickRate float64) *World {
	return &World{
		entities: make(map[EntityID]map[ComponentType]Component),
		TickRate: tickRate,
	}
}

// Spawn creates a new entity and returns its ID
func (w *World) Spawn() EntityID {
	id := NewEntityID()
	w.entities[id] = make(map[ComponentType]Component)
	return id
}

// Attach adds a component to an entity
func (w *World) Attach(id EntityID, c Component) {
	if comps, ok := w.entities[id]; ok {
		comps[c.Type()] = c
	}
}

// Detach removes a component from an entity
func (w *World) Detach(id EntityID, ct ComponentType) {
	if comps, ok := w.entities[id]; ok {
		delete(comps, ct)
	}
}

// Get returns a component for an entity, or nil
func (w *World) Get(id EntityID, ct ComponentType) Component {
	if comps, ok := w.entities[id]; ok {
		return comps[ct]
	}
	return nil
}

// Has checks if an entity has a component
func (w *World) Has(id EntityID, ct ComponentType) bool {
	if comps, ok := w.entities[id]; ok {
		_, exists := comps[ct]
		return exists
	}
	return false
}

// Destroy marks an entity for removal
func (w *World) Destroy(id EntityID) {
	w.toRemove = append(w.toRemove, id)
}

// Query returns all entity IDs that have ALL specified component types
func (w *World) Query(types ...ComponentType) []EntityID {
	var result []EntityID
	for id, comps := range w.entities {
		match := true
		for _, t := range types {
			if _, ok := comps[t]; !ok {
				match = false
				break
			}
		}
		if match {
			result = append(result, id)
		}
	}
	return result
}

// AddSystem registers a system
func (w *World) AddSystem(s System) {
	w.systems = append(w.systems, s)
	// Sort by priority (simple insertion)
	for i := len(w.systems) - 1; i > 0; i-- {
		if w.systems[i].Priority() < w.systems[i-1].Priority() {
			w.systems[i], w.systems[i-1] = w.systems[i-1], w.systems[i]
		}
	}
}

// Tick runs all systems once
func (w *World) Tick(dt float64) {
	for _, s := range w.systems {
		s.Update(w, dt)
	}
	// Clean up destroyed entities
	for _, id := range w.toRemove {
		delete(w.entities, id)
	}
	w.toRemove = w.toRemove[:0]
	w.TickCount++
}

// EntityCount returns the number of alive entities
func (w *World) EntityCount() int {
	return len(w.entities)
}

package core

// Event represents a game event
type Event struct {
	Type    EventType
	Tick    uint64
	Payload interface{}
}

type EventType uint16

const (
	EvtUnitCreated EventType = iota
	EvtUnitDestroyed
	EvtBuildingPlaced
	EvtBuildingDestroyed
	EvtBuildingComplete
	EvtUnitAttack
	EvtUnitDamaged
	EvtUnitMoveOrder
	EvtProjectileFired
	EvtProjectileHit
	EvtResourceHarvested
	EvtResourceSpent
	EvtTechUnlocked
	EvtPlayerDefeated
	EvtPlayerAlliance
	EvtChatMessage
	EvtGameStart
	EvtGameEnd
)

// EventBus dispatches events to listeners
type EventBus struct {
	listeners map[EventType][]EventHandler
	queue     []Event
}

type EventHandler func(e Event)

func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[EventType][]EventHandler),
	}
}

// On registers a handler for an event type
func (eb *EventBus) On(t EventType, h EventHandler) {
	eb.listeners[t] = append(eb.listeners[t], h)
}

// Emit queues an event for dispatch
func (eb *EventBus) Emit(e Event) {
	eb.queue = append(eb.queue, e)
}

// Dispatch processes all queued events
func (eb *EventBus) Dispatch() {
	for _, e := range eb.queue {
		if handlers, ok := eb.listeners[e.Type]; ok {
			for _, h := range handlers {
				h(e)
			}
		}
	}
	eb.queue = eb.queue[:0]
}

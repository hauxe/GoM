package event

// Event event type
type Event interface {
	ID() string
	Name() string
	Cond() bool
}

// Emitter event emitter
type Emitter interface {
	Emit(evt Event)
}

// Handler event handler
type Handler interface {
	IsEmitted(Event) bool
	On(evt Event, handler func(Event))
	Off(evt Event)
}

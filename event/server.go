package event

// Server source of event
type Server struct {
}

// NewEmitter returns event emitter
func NewEmitter() Emitter {
	return &Server{}
}

// Emit emit the event
func (s *Server) Emit(evt Event) {
	event := initEvent(evt)
	event.Broadcast()
}

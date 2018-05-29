package broadcast

import (
	"sync"

	"github.com/pkg/errors"
)

type broadcast struct {
	c        chan broadcast
	v        interface{}
	shutdown chan struct{}
}

// Broadcaster defines broadcaster properties
type Broadcaster struct {
	listenc  chan chan (chan broadcast)
	sendc    chan<- interface{}
	shutdown chan struct{}
	isClosed bool
	mux      sync.RWMutex
}

// Receiver defines receiver struct
type Receiver struct {
	c        chan broadcast
	shutdown chan struct{}
}

// BroadcasterClosed broadcaster closed
type BroadcasterClosed error

// NewBroadcaster create a new broadcaster object.
func NewBroadcaster() *Broadcaster {
	listenc := make(chan (chan (chan broadcast)))
	sendc := make(chan interface{})
	shutdown := make(chan struct{})
	broadcaster := &Broadcaster{
		listenc:  listenc,
		sendc:    sendc,
		shutdown: shutdown,
	}
	go func() {
		currc := make(chan broadcast, 1)
		for {
			select {
			case v := <-sendc:
				c := make(chan broadcast, 1)
				b := broadcast{c: c, v: v}
				currc <- b
				currc = c
			case r := <-listenc:
				r <- currc
			case <-shutdown:
				// shutdown broadcaster
				broadcaster.mux.Lock()
				defer broadcaster.mux.Unlock()
				broadcaster.isClosed = true
				close(broadcaster.listenc)
				close(broadcaster.sendc)
				return
			}
		}
	}()
	return broadcaster
}

// Close closes broadcaster
func (b *Broadcaster) Close() {
	close(b.shutdown)
}

// Listen start listening to the broadcasts.
func (b *Broadcaster) Listen() (*Receiver, error) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if b.isClosed && b.listenc != nil && b.sendc != nil {
		return nil, errors.New("broadcaster is closed or channel is not initialized")
	}
	c := make(chan chan broadcast, 0)
	b.listenc <- c
	return &Receiver{c: <-c, shutdown: b.shutdown}, nil
}

// Write broadcast a value to all listeners.
func (b *Broadcaster) Write(v interface{}) error {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if b.isClosed && b.listenc != nil && b.sendc != nil {
		return errors.New("broadcaster is closed or channel is not initialized")
	}
	b.sendc <- v
	return nil
}

// Read read a value that has been broadcast,
// waiting until one is available if necessary.
func (r *Receiver) Read(handlers ...func(interface{}) (interface{}, error)) (interface{}, error) {
	select {
	case b := <-r.c:
		v := b.v
		for _, handler := range handlers {
			val, err := handler(v)
			if err != nil {
				return nil, err
			}
			b.v = val
		}
		r.c <- b
		r.c = b.c
		return v, nil
	case <-r.shutdown:
		r.c = nil
		return nil, BroadcasterClosed(errors.New("broadcaster is closed"))
	}
}

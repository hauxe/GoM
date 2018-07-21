package event

import (
	"sync"
)

// Events map of events
var Events map[string]*sync.Cond
var mux sync.Mutex

func init() {
	Events = make(map[string]*sync.Cond)
}

// Client defines event client structure
type Client struct {
	mux     sync.Mutex
	emitted bool
	off     chan struct{}
}

// New returns new event handler
func New() Handler {
	return &Client{off: make(chan struct{}, 1)}
}

// On subscriber to an event with synchronous callback
func (c *Client) On(evt Event, handler func(Event)) {
	event := initEvent(evt)
	wait := new(sync.WaitGroup)
	wait.Add(1)
	go func() {
		wait.Done()
		for {
			select {
			case <-c.off:
				return
			default:
				event.L.Lock()
				for evt.Cond() == false {
					event.Wait()
					c.mux.Lock()
					c.emitted = true
					c.mux.Unlock()
				}
				handler(evt)
				c.mux.Lock()
				c.emitted = false
				c.mux.Unlock()
				event.L.Unlock()
			}
		}
	}()
	wait.Wait()
}

// Off turn of event listener
func (c *Client) Off(evt Event) {
	c.off <- struct{}{}
}

// IsEmitted check if event is emitted or not
func (c *Client) IsEmitted(evt Event) bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.emitted
}

func initEvent(evt Event) *sync.Cond {
	mux.Lock()
	defer mux.Unlock()
	id := evt.ID()
	if Events[id] == nil {
		Events[id] = sync.NewCond(&sync.Mutex{})
	}
	return Events[id]
}

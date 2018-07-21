package event

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type event struct {
	id      string
	name    string
	counter int
	limit   int
	handler Handler
	wg      *sync.WaitGroup
}

func (e *event) ID() string {
	return e.id
}
func (e *event) Name() string {
	return e.name
}
func (e *event) Cond() bool {
	defer e.wg.Done()
	if !e.handler.IsEmitted(e) {
		return false
	}
	e.counter++
	return e.counter >= e.limit
}

func TestEvent(t *testing.T) {
	t.Parallel()
	limit := 2
	client := New()
	e := &event{
		id:      "test",
		name:    "test event",
		limit:   limit,
		handler: client,
		wg:      new(sync.WaitGroup),
	}
	mux := new(sync.Mutex)
	called := false
	e.wg.Add(1)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	client.On(e, func(evt Event) {
		defer wg.Done()
		_, ok := evt.(*event)
		require.True(t, ok)
		mux.Lock()
		defer mux.Unlock()
		called = true
	})
	e.wg.Wait()
	server := NewEmitter()
	e.wg.Add(1)
	server.Emit(e)
	e.wg.Wait()
	mux.Lock()
	require.False(t, called)
	mux.Unlock()
	time.Sleep(time.Second)
	e.wg.Add(2)
	server.Emit(e)
	e.wg.Wait()
	wg.Wait()
	require.True(t, called)
	require.Equal(t, limit, e.counter)
}

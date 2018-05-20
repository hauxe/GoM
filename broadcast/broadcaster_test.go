package broadcast

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBroadcaster(t *testing.T) {
	t.Parallel()
	b := NewBroadcaster()
	var val int32
	num := 100
	factor := int32(1)
	for i := 0; i < num; i++ {
		r, err := b.Listen()
		require.Nil(t, err)
		go func(r *Receiver, index int) {
			for {
				v, err := r.Read()
				if err != nil {
					atomic.AddInt32(&val, -1*factor)
					return
				}
				atomic.AddInt32(&val, v.(int32))
			}
		}(r, i)
	}
	// broadcast some value
	b.Write(factor)
	for v := atomic.LoadInt32(&val); v != int32(num)*factor; v = atomic.LoadInt32(&val) {
		time.Sleep(time.Millisecond * 100)
	}
	// stop broadcast listen
	b.Close()
	for v := atomic.LoadInt32(&val); v != 0; v = atomic.LoadInt32(&val) {
		time.Sleep(time.Millisecond * 100)
	}
	// after stop can't listen
	r, err := b.Listen()
	require.Error(t, err)
	require.Nil(t, r)
}

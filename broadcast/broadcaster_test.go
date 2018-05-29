package broadcast

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBroadcaster(t *testing.T) {
	t.Parallel()
	b := NewBroadcaster()
	var val int32
	num := 100
	factor := int32(1)
	wg := new(sync.WaitGroup)
	wg.Add(num)
	for i := 0; i < num; i++ {
		r, err := b.Listen()
		require.Nil(t, err)
		go func(r *Receiver, index int) {
			for {
				v, err := r.Read()
				if err != nil {
					atomic.AddInt32(&val, -1*factor)
					wg.Done()
					return
				}
				atomic.AddInt32(&val, v.(int32))
				wg.Done()
			}
		}(r, i)
	}
	// broadcast some value
	b.Write(factor)
	wg.Wait()
	require.Equal(t, int32(num)*factor, atomic.LoadInt32(&val))
	// stop broadcast listen
	wg.Add(num)
	b.Close()
	wg.Wait()
	require.Equal(t, int32(0), atomic.LoadInt32(&val))
	// after stop can't listen
	r, err := b.Listen()
	require.Error(t, err)
	require.Nil(t, r)
}

func TestBroadcasterHandlers(t *testing.T) {
	t.Parallel()
	t.Run("error", func(t *testing.T) {
		t.Parallel()
		b := NewBroadcaster()
		r, err := b.Listen()
		require.Nil(t, err)
		go func(r *Receiver) {
			for {
				_, err := r.Read(func(_ interface{}) (interface{}, error) {
					return nil, errors.New("test error")
				})
				require.Error(t, err)
			}
		}(r)
		// broadcast some value
		b.Write("any")
	})
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		b := NewBroadcaster()
		var val int32
		num := 100
		factor := int32(1)
		wg := new(sync.WaitGroup)
		wg.Add(num)
		for i := 0; i < num; i++ {
			r, err := b.Listen()
			require.Nil(t, err)
			go func(r *Receiver, index int) {
				for {
					v, err := r.Read(func(i interface{}) (interface{}, error) {
						return i.(int32) + 1, nil
					})
					if err != nil {
						return
					}
					atomic.AddInt32(&val, v.(int32))
					wg.Done()
				}
			}(r, i)
		}
		// broadcast some value
		b.Write(factor)
		wg.Wait()
		require.Equal(t, int32(num)*factor+int32(num*(num-1)/2), atomic.LoadInt32(&val))
		// stop broadcast listen
		b.Close()
	})
}

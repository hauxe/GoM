package retry

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"
)

func TestRetry(t *testing.T) {
	t.Parallel()
	t.Run("cancel by context", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		client, err := CreateClient(UseConstantRetry(time.Millisecond * 100))
		require.Nil(t, err)
		require.NotNil(t, client)
		var counter int32
		factor := int32(1)
		f := func() error {
			atomic.AddInt32(&counter, factor)
			return errors.New("test error")
		}
		err = client.Init(client.SetContextOption(ctx), client.SetMaxRetriesOption(10))
		require.Nil(t, err)
		ch := client.Do(f)
		time.Sleep(time.Millisecond * 250)
		cancel()
		result := <-ch
		require.False(t, result)
		require.Equal(t, int32(4), atomic.LoadInt32(&counter))
	})

	t.Run("reach maximum retries", func(t *testing.T) {
		t.Parallel()
		client, err := CreateClient(UseConstantRetry(time.Millisecond * 100))
		require.Nil(t, err)
		require.NotNil(t, client)
		var counter int32
		factor := int32(1)
		f := func() error {
			atomic.AddInt32(&counter, factor)
			return errors.New("test error")
		}
		err = client.Init(client.SetMaxRetriesOption(10))
		require.Nil(t, err)
		ch := client.Do(f)
		result := <-ch
		require.False(t, result)
		require.Equal(t, int32(11), atomic.LoadInt32(&counter))
	})

	t.Run("success after several retries", func(t *testing.T) {
		t.Parallel()
		client, err := CreateClient(UseConstantRetry(time.Millisecond * 100))
		require.Nil(t, err)
		require.NotNil(t, client)
		var counter int32
		factor := int32(1)
		f := func() error {
			atomic.AddInt32(&counter, factor)
			if atomic.LoadInt32(&counter) == int32(5) {
				return nil
			}
			return errors.New("test error")
		}
		err = client.Init(client.SetMaxRetriesOption(10))
		require.Nil(t, err)
		ch := client.Do(f)
		result := <-ch
		require.True(t, result)
		require.Equal(t, int32(5), atomic.LoadInt32(&counter))
	})
}

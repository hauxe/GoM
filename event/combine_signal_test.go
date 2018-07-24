package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOR(t *testing.T) {
	t.Parallel()
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}
	start := time.Now()
	<-OR(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	require.WithinDuration(t, time.Now(), start, 2*time.Second)
}

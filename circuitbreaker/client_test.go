package circuitbreaker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	t.Parallel()
	t.Run("max_execution_time", func(t *testing.T) {
		t.Parallel()
		client, err := CreateClient()
		require.Nil(t, err)
		require.NotNil(t, client)
		timeout := 100 * time.Microsecond
		requestThreshold := int64(20)
		errorThreshold := int64(50)
		rollingDuration := 5 * time.Second
		rollingBuckets := 5
		sleepWindow := 100 * time.Microsecond
		halfOpenAttempts := int64(10)
		requireSuccess := int64(2)
		require.Nil(t, client.Start(t.Name(),
			client.SetExecuteTimeoutOption(timeout),
			client.SetOpenerRequestThreshold(requestThreshold),
			client.SetOpenerErrorThreshold(errorThreshold),
			client.SetOpenerNumBuckets(rollingBuckets),
			client.SetOpenerRollingDuration(rollingDuration),
			client.SetFallbackDisable(true),
			client.SetCloserSleepWindow(sleepWindow),
			client.SetCloserAttempts(halfOpenAttempts),
			client.SetCloserRequiredSuccessful(requireSuccess)))
		wg := new(sync.WaitGroup)
		wg.Add(int(requestThreshold))
		for i := int64(0); i < requestThreshold; i++ {
			go func(i int64) {
				err := client.Execute(context.Background(), t.Name(),
					func(_ context.Context) error {
						if i%2 == 0 {
							time.Sleep(2 * timeout)
						}
						return nil
					}, func(_ context.Context, _ error) error {
						t.Error("can not reach fallback here")
						return nil
					})
				require.Nil(t, err)
				wg.Done()
			}(i)
		}
		wg.Wait()
		err = client.Execute(context.Background(), t.Name(),
			func(_ context.Context) error {
				time.Sleep(2 * timeout)
				return nil
			}, func(_ context.Context, _ error) error {
				t.Error("can not reach fallback here")
				return nil
			})
		require.NotNil(t, err)
		// sleep 100 microsecond and then do success request to open circuit again
		time.Sleep(sleepWindow + time.Microsecond*100)
		for i := int64(0); i < requireSuccess+1; i++ {
			err = client.Execute(context.Background(), t.Name(),
				func(_ context.Context) error {
					return nil
				}, func(_ context.Context, _ error) error {
					t.Error("can not reach fallback here")
					return nil
				})
			require.Nil(t, err)
		}
		for i := int64(0); i < halfOpenAttempts+requireSuccess; i++ {
			err = client.Execute(context.Background(), t.Name(),
				func(_ context.Context) error {
					return nil
				}, func(_ context.Context, _ error) error {
					t.Error("can not reach fallback here")
					return nil
				})
			require.Nil(t, err)
		}
	})
	t.Run("throttle", func(t *testing.T) {
		t.Parallel()
		client, err := CreateClient()
		require.Nil(t, err)
		require.NotNil(t, client)
		throttle := int64(100)
		require.Nil(t, client.Start(t.Name(),
			client.SetExecuteMaxConcurrentRequestsOption(throttle),
			client.SetFallbackDisable(true)))
		wg := new(sync.WaitGroup)
		wg.Add(int(throttle))
		for i := int64(0); i < throttle; i++ {
			go func(i int64) {
				// done before execute to test throttle
				wg.Done()
				err = client.Execute(context.Background(), t.Name(),
					func(_ context.Context) error {
						time.Sleep(2 * time.Second)
						return nil
					}, func(_ context.Context, _ error) error {
						t.Error("can not reach fallback here")
						return nil
					})
				require.Nil(t, err)
			}(i)
		}
		wg.Wait()
		// reach max councurrent request
		err = client.Execute(context.Background(), t.Name(),
			func(_ context.Context) error {
				return nil
			}, func(_ context.Context, _ error) error {
				t.Error("can not reach fallback here")
				return nil
			})
		require.NotNil(t, err)
	})
}

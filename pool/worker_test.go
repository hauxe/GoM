package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hauxe/gom/environment"
	lib "github.com/hauxe/gom/library"
	"github.com/pkg/errors"
)

func TestCreateWorker(t *testing.T) {
	t.Parallel()
	t.Run("error create env", func(t *testing.T) {
		t.Parallel()
		worker, err := CreateWorker(func(_ *environment.ENVConfig) error {
			return errors.New("test env error")
		})
		require.Error(t, err)
		require.Nil(t, worker)
	})

	t.Run("succcess", func(t *testing.T) {
		t.Parallel()
		worker, err := CreateWorker()
		require.Nil(t, err)
		require.NotNil(t, worker)
	})
}

type job struct {
	name string
	f    func()
}

// Name get job name
func (j *job) Name() string {
	return j.name
}

// GetContext get http context
func (j *job) GetContext() context.Context {
	return context.Background()
}

// Execute handle job
func (j *job) Execute() error {
	j.f()
	return nil
}

func TestStartWorker(t *testing.T) {
	t.Parallel()

	t.Run("error empty config", func(t *testing.T) {
		t.Parallel()
		worker := Worker{}
		err := worker.StartServer()
		require.Error(t, err)
	})

	t.Run("error option", func(t *testing.T) {
		t.Parallel()
		worker, err := CreateWorker()
		require.Nil(t, err)
		require.NotNil(t, worker)
		err = worker.StartServer(func() error {
			return errors.New("test option error")
		})
		require.Error(t, err)
	})

	t.Run("success drop timeout jobs", func(t *testing.T) {
		t.Parallel()
		numJob := 10
		numWorker := 10
		worker, err := CreateWorker()
		require.Nil(t, err)
		require.NotNil(t, worker)
		err = worker.StartServer(worker.SetMaxWorkersOption(numWorker))
		require.Nil(t, err)
		var counter int32
		factor := int32(1)
		var wg1 sync.WaitGroup
		var wg2 sync.WaitGroup
		f := func() {
			atomic.AddInt32(&counter, factor)
			wg1.Done()
			// sleep 500 millisecond before counting more
			time.Sleep(time.Millisecond * 500)
			atomic.AddInt32(&counter, factor)
			wg2.Done()
		}
		// add 20 job to worker
		for i := 0; i < numJob; i++ {
			j := job{
				name: "job: " + lib.ToString(i),
				f:    f,
			}
			if i < numWorker {
				wg1.Add(1)
				wg2.Add(1)
			}
			err := worker.QueueJob(&j, 100*time.Millisecond)
			if i < numWorker {
				require.Nil(t, err)
			} else {
				require.Error(t, err)
			}
		}
		wg1.Wait()
		require.Equal(t, int32(numWorker)*factor, atomic.LoadInt32(&counter))
		wg2.Wait()
		// after timeout, workers continue increase counter
		require.Equal(t, int32(numWorker*2)*factor, atomic.LoadInt32(&counter))
		worker.StopServer()
	})

	t.Run("success continue after timeout jobs", func(t *testing.T) {
		t.Parallel()
		numJob := 20
		numWorker := 10
		worker, err := CreateWorker()
		require.Nil(t, err)
		require.NotNil(t, worker)
		err = worker.StartServer(worker.SetMaxWorkersOption(numWorker))
		require.Nil(t, err)
		var counter int32
		factor := int32(1)
		var wg sync.WaitGroup
		f := func() {
			atomic.AddInt32(&counter, factor)
			// sleep 500 millisecond before counting more
			time.Sleep(time.Millisecond * 100)
			atomic.AddInt32(&counter, factor)
			wg.Done()
		}
		// add 20 job to worker
		for i := 0; i < numJob; i++ {
			j := job{
				name: "job: " + lib.ToString(i),
				f:    f,
			}
			wg.Add(1)
			err := worker.QueueJob(&j, 500*time.Millisecond)
			require.Nil(t, err)
		}
		wg.Wait()
		require.Equal(t, int32(numJob*2)*factor, atomic.LoadInt32(&counter))
		worker.StopServer()
	})
}

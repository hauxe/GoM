package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJobName(t *testing.T) {
	t.Parallel()
	name := "name"
	job := &JobHandler{
		name: name,
	}
	require.Equal(t, name, job.Name())
}

func TestJobContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	type key string
	k := key("test_key")
	v := "test_value"
	ctx = context.WithValue(ctx, k, v)
	r := &http.Request{}
	job := &JobHandler{
		r: r.WithContext(ctx),
	}
	cont := job.GetContext()
	require.NotNil(t, cont)
	require.Equal(t, v, cont.Value(k).(string))
}

func TestExecute(t *testing.T) {
	t.Parallel()
	t.Run("error empty handler", func(t *testing.T) {
		t.Parallel()
		job := JobHandler{}
		require.Error(t, job.Execute())
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		f := func(_ http.ResponseWriter, _ *http.Request) {

		}
		job := JobHandler{handler: f}
		require.Nil(t, job.Execute())
	})
}

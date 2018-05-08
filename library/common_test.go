package library

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"
)

func TestGetURL(t *testing.T) {
	t.Parallel()
	host := "localhost"
	port := 1234
	require.Equal(t, "localhost:1234", GetURL(host, port))
}

func TestStringTags(t *testing.T) {
	t.Parallel()
	t.Run("empty tags", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "", StringTags())
	})
	t.Run("not empty tags", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "[field1][field2][field3][field4]", StringTags("field1",
			"field2", "field3", "field4"))
	})
}

func TestRunOptionalFunc(t *testing.T) {
	t.Parallel()
	t.Run("error", func(t *testing.T) {
		t.Parallel()
		a := 0
		f1 := func() error {
			a++
			return nil
		}
		f2 := func() error {
			a++
			return nil
		}
		f3 := func() error {
			return errors.New("error")
		}
		err := RunOptionalFunc(f1, f2, f3)
		require.Equal(t, 2, a)
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		a := 0
		f1 := func() error {
			a++
			return nil
		}
		f2 := func() error {
			a++
			return nil
		}
		f3 := func() error {
			a++
			return nil
		}
		err := RunOptionalFunc(f1, f2, f3)
		require.Equal(t, 3, a)
		require.Nil(t, err)
	})
}

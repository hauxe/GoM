package library

import (
	"testing"

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

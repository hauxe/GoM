package library

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeJSON(t *testing.T) {
	t.Parallel()
	a := time.Now()
	ti := TimeRFC3339(a)
	var tii TimeRFC3339
	// json marshal
	mTi, err := json.Marshal(&ti)
	require.Nil(t, err)
	require.Equal(t, "\""+a.Format(time.RFC3339)+"\"", string(mTi))
	// json unmarshal
	err = json.Unmarshal(mTi, &tii)
	require.Nil(t, err)
	require.Equal(t, time.Time(ti).Format(time.RFC3339), time.Time(tii).Format(time.RFC3339))
}

func TestTimeIsSet(t *testing.T) {
	t.Parallel()
	t.Run("not set", func(t *testing.T) {
		t.Parallel()
		var ti TimeRFC3339
		require.False(t, ti.IsSet())
	})

	t.Run("set", func(t *testing.T) {
		t.Parallel()
		ti := TimeRFC3339(time.Now())
		require.True(t, ti.IsSet())
	})
}

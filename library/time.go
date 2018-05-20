package library

import (
	"fmt"
	"strings"
	"time"
)

var nilTime = (time.Time{}).UnixNano()

//TimeRFC3339 custom time type use RFC3339 format
type TimeRFC3339 time.Time

//UnmarshalJSON parse from json string
func (t *TimeRFC3339) UnmarshalJSON(b []byte) (err error) {
	stamp := strings.Trim(string(b), "\"")
	ti, err := time.Parse(time.RFC3339, stamp)
	if err != nil {
		return err
	}
	*t = TimeRFC3339(ti)
	return nil
}

//MarshalJSON format to json string
func (t *TimeRFC3339) MarshalJSON() ([]byte, error) {
	ti := time.Time(*t)
	if ti.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	stamp := fmt.Sprintf("\"%s\"", ti.Format(time.RFC3339))
	return []byte(stamp), nil
}

//IsSet check time is set
func (t *TimeRFC3339) IsSet() bool {
	ti := time.Time(*t)
	return ti.UnixNano() != nilTime
}

package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringSlice is slice of strings
type StringSlice []string

// Scan sqlx JSON scan method
func (s *StringSlice) Scan(val interface{}) error {
	*s = StringSlice{}
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, &s)
	case string:
		return json.Unmarshal([]byte(v), &s)
	default:
		return fmt.Errorf("Unsupported type: %T", v)
	}
}

// Value sqlx JSON value method
func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

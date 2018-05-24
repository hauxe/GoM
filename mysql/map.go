package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringMap defines map string-string for mysql
type StringMap map[string]string

// Scan sqlx JSON scan method
func (m *StringMap) Scan(val interface{}) error {
	m = &StringMap{}
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, &m)
	case string:
		return json.Unmarshal([]byte(v), &m)
	default:
		return fmt.Errorf("Unsupported type: %T", v)
	}
}

// Value sqlx JSON value method
func (m StringMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

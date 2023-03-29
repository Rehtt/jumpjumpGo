package database

import (
	"database/sql/driver"
	"encoding/json"
)

type JSON[T any] struct {
	Data T
}

func (JSON[T]) GormDataType() string {
	return "json"
}
func (m *JSON[T]) Scan(value any) error {
	return json.Unmarshal([]byte(value.(string)), &m.Data)
}
func (m JSON[T]) Value() (value driver.Value, err error) {
	src, err := json.Marshal(m.Data)
	return string(src), err
}

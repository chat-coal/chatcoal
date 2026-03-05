package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/bwmarrin/snowflake"
)

// Snowflake is a 64-bit ID type that serializes as a JSON string
// to avoid JavaScript precision loss (Number.MAX_SAFE_INTEGER = 2^53-1).
type Snowflake int64

var snowflakeNode *snowflake.Node

// InitSnowflake sets the custom epoch (2024-01-01) and creates a node.
func InitSnowflake(nodeID int64) {
	snowflake.Epoch = 1704067200000 // 2024-01-01 00:00:00 UTC in ms
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		panic(fmt.Sprintf("failed to create snowflake node: %v", err))
	}
	snowflakeNode = node
}

// GenerateID returns a new unique Snowflake ID.
func GenerateID() Snowflake {
	return Snowflake(snowflakeNode.Generate().Int64())
}

// MarshalJSON serializes the Snowflake as a JSON string.
func (s Snowflake) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(s), 10))
}

// UnmarshalJSON accepts both string ("123") and number (123) JSON values.
func (s *Snowflake) UnmarshalJSON(data []byte) error {
	// Try string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		v, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid snowflake string: %w", err)
		}
		*s = Snowflake(v)
		return nil
	}
	// Fall back to number
	var num float64
	if err := json.Unmarshal(data, &num); err != nil {
		return fmt.Errorf("invalid snowflake value: %s", string(data))
	}
	*s = Snowflake(int64(num))
	return nil
}

// Scan implements sql.Scanner for reading bigint from the database.
func (s *Snowflake) Scan(value interface{}) error {
	if value == nil {
		*s = 0
		return nil
	}
	switch v := value.(type) {
	case int64:
		*s = Snowflake(v)
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		*s = Snowflake(n)
	default:
		return fmt.Errorf("cannot scan %T into Snowflake", value)
	}
	return nil
}

// Value implements driver.Valuer for writing bigint to the database.
func (s Snowflake) Value() (driver.Value, error) {
	return int64(s), nil
}

// GormDataType tells GORM to use bigint for this column.
func (s Snowflake) GormDataType() string {
	return "bigint"
}

// Int64 returns the underlying int64 value.
func (s Snowflake) Int64() int64 {
	return int64(s)
}

// String returns the string representation.
func (s Snowflake) String() string {
	return strconv.FormatInt(int64(s), 10)
}

package types

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// FlexTime is a time.Time wrapper that accepts both RFC3339 and YYYY-MM-DD formats in JSON.
type FlexTime struct {
	time.Time
}

func (ft *FlexTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		ft.Time = time.Time{}
		return nil
	}

	// Try RFC3339 first
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		ft.Time = t
		return nil
	}

	// Try date-only format
	t, err = time.Parse("2006-01-02", s)
	if err == nil {
		ft.Time = t
		return nil
	}

	return fmt.Errorf("flextime: cannot parse %q as time", s)
}

func (ft FlexTime) MarshalJSON() ([]byte, error) {
	if ft.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + ft.Time.Format(time.RFC3339) + `"`), nil
}

// Value implements driver.Valuer for GORM.
func (ft FlexTime) Value() (driver.Value, error) {
	if ft.Time.IsZero() {
		return nil, nil
	}
	return ft.Time, nil
}

// Scan implements sql.Scanner for GORM.
func (ft *FlexTime) Scan(value interface{}) error {
	if value == nil {
		ft.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		ft.Time = v
		return nil
	}
	return fmt.Errorf("flextime: cannot scan type %T", value)
}

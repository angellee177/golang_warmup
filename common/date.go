package common

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Date time.Time

const DateFormat = "2006-01-02"

// UnmarshalJSON parses the YYYY-MM-DD string into the Date type
func (d *Date) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		return nil
	}
	// Remove quotes from the JSON string
	t, err := time.Parse(`"`+DateFormat+`"`, s)
	if err != nil {
		return err
	}
	*d = Date(t)
	return nil
}

// MarshalJSON converts the Date back to a JSON string
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, time.Time(d).Format(DateFormat))), nil
}

// GORM support: Value converts Date to database format
func (d Date) Value() (driver.Value, error) {
	return time.Time(d), nil
}

// GORM support: Scan converts database format to Date
func (d *Date) Scan(value interface{}) error {
	if t, ok := value.(time.Time); ok {
		*d = Date(t)
		return nil
	}
	return fmt.Errorf("failed to scan Date: %v", value)
}

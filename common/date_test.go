package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDate_UnmarshalJSON(t *testing.T) {
	// Happy Path: Correct format
	t.Run("Success - Valid Date String", func(t *testing.T) {
		var d Date
		input := []byte(`"2026-03-25"`)

		err := d.UnmarshalJSON(input)

		require.NoError(t, err)
		assert.Equal(t, 2026, time.Time(d).Year())
		assert.Equal(t, time.March, time.Time(d).Month())
		assert.Equal(t, 25, time.Time(d).Day())
	})

	// Happy Path: Null handling
	t.Run("Success - Null Input", func(t *testing.T) {
		var d Date
		err := d.UnmarshalJSON([]byte("null"))

		require.NoError(t, err)
	})

	// Negative Path: Wrong format (includes timestamp)
	t.Run("Failure - Invalid Format", func(t *testing.T) {
		var d Date
		input := []byte(`"2026-03-25T15:00:00Z"`)

		err := d.UnmarshalJSON(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parsing time")
	})

	// Negative Path: Random string
	t.Run("Failure - Not a Date", func(t *testing.T) {
		var d Date
		err := d.UnmarshalJSON([]byte(`"hello-world"`))

		assert.Error(t, err)
	})
}

func TestDate_MarshalJSON(t *testing.T) {
	t.Run("Success - Marshal to String", func(t *testing.T) {
		now := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		d := Date(now)

		output, err := d.MarshalJSON()

		require.NoError(t, err)
		assert.Equal(t, []byte(`"2026-12-31"`), output)
	})
}

func TestDate_Scan(t *testing.T) {
	// Happy Path: Database returns time.Time
	t.Run("Success - Scan from time.Time", func(t *testing.T) {
		var d Date
		dbTime := time.Now()

		err := d.Scan(dbTime)

		require.NoError(t, err)
		assert.True(t, time.Time(d).Equal(dbTime))
	})

	// Negative Path: Database returns a string or int
	t.Run("Failure - Scan from invalid type", func(t *testing.T) {
		var d Date
		err := d.Scan("2026-01-01") // Scan logic expects time.Time object

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan Date")
	})
}

func TestDate_Value(t *testing.T) {
	t.Run("Success - Convert Date to driver.Value", func(t *testing.T) {
		// Arrange: Create a specific date
		expectedTime := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
		d := Date(expectedTime)

		// Act: Call Value()
		val, err := d.Value()

		// Assert
		require.NoError(t, err)

		// GORM/SQL drivers expect a time.Time object for date columns
		actualTime, ok := val.(time.Time)
		assert.True(t, ok, "Value() should return a time.Time object")
		assert.True(t, actualTime.Equal(expectedTime))
	})

	t.Run("Success - Zero Date Value", func(t *testing.T) {
		var d Date // Zero value
		val, err := d.Value()

		require.NoError(t, err)
		assert.True(t, val.(time.Time).IsZero())
	})
}

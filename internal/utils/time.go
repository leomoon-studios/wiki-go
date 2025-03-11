package utils

import (
	"log"
	"time"
	_ "time/tzdata"
)

// FormatTimeInTimezone formats a time.Time value using the specified timezone
// If the timezone is invalid, it falls back to UTC
func FormatTimeInTimezone(t time.Time, timezone string, format string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Error loading timezone %s: %v, falling back to UTC", timezone, err)
		loc = time.UTC
	}

	return t.In(loc).Format(format)
}

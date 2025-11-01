// Package timely provides utilities to parse time strings into time.Time
// values. This package is intended to be imported by other packages.
package timeutils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// validTimeFormat matches accepted time strings:
	// - 1 to 4 digits, or
	// - 1-2 digits, colon, 2 digits (H:MM or HH:MM)
	validTimeFormat = regexp.MustCompile(`^(\d{1,4}|\d{1,2}:\d{2})$`)
)

// ParseTime parses common short time formats into a time.Time value. The
// returned time uses the local date (the Zero date is not preserved) but the
// hour and minute fields are set from the parsed value.
//
// Accepted input examples:
//   - "01", "1" -> 01:00
//   - "14", "1400", "14:00" -> 14:00
//   - "730", "7:30", "0730" -> 07:30
//
// The input may contain only digits and an optional single ":" separator.
// An error is returned for invalid formats or out-of-range hour/minute values.
func ParseTime(timeStr string) (time.Time, error) {
	if !validTimeFormat.MatchString(timeStr) {
		return time.Time{}, fmt.Errorf("%s is not a supported time format: ", timeStr)
	}

	// Normalize by removing colon
	timeStr = strings.ReplaceAll(timeStr, ":", "")

	switch len(timeStr) {
	case 1, 2:
		// hour only: append minutes = 00
		timeStr = fmt.Sprintf("%02s00", timeStr)
	case 3:
		// e.g. 730 -> 0730
		timeStr = "0" + timeStr
	case 4:
		// already HHMM
	default:
		return time.Time{}, fmt.Errorf("unsupported time format length: %d", len(timeStr))
	}

	hours, err := strconv.Atoi(timeStr[:2])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid hours: %w", err)
	}
	minutes, err := strconv.Atoi(timeStr[2:])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid minutes: %w", err)
	}

	if hours < 0 || hours > 23 {
		return time.Time{}, fmt.Errorf("hours out of range (0-23): %d", hours)
	}
	if minutes < 0 || minutes > 59 {
		return time.Time{}, fmt.Errorf("minutes out of range (0-50): %d", minutes)
	}

	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hours, minutes, 0, 0, time.Local), nil
}

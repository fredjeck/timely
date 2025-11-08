// Package timeutils provides utilities for working with time.Time values and durations.
// It includes functions for calculating durations between paired time values and
// handling time-based calculations in a deterministic way.
package timeutils

import (
	"fmt"
	"sort"
	"time"
)

// FormatDuration formats a time.Duration into a string in "HH:MM" format.
// It handles negative durations by prefixing the result with a minus sign.
func FormatDuration(d time.Duration) string {
	if d < 0 {
		return "-" + FormatDuration(-d)
	}
	h := int(d / time.Hour)
	m := int((d % time.Hour) / time.Minute)
	return fmt.Sprintf("%02d:%02d", h, m)
}

// FormatTime formats a time.Duration into a string in "HH:MM" format.
// It handles negative durations by prefixing the result with a minus sign.
func FormatTime(d time.Time) string {
	return d.Format("15:04")
}

// Durations represents an ordered collection of time.Time values.
// The collection maintains chronological order (ascending) when elements
// are added or removed.
type Durations []time.Time

// Last returns the last time.Time value in the Durations collection.
// If the collection is empty, it returns the zero value of time.Time.
func (durations Durations) Last() time.Time {
	if len(durations) == 0 {
		return time.Time{}
	}
	return durations[len(durations)-1]
}

// sortTimesAscending sorts a slice of time.Time values in ascending order.
// This is an internal helper used to maintain chronological order of Durations.
func sortTimesAscending(times []time.Time) {
	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
}

// Append adds a new time to the Durations collection and maintains chronological order.
// Returns a new Durations slice with the added time in sorted position.
func (durations Durations) Append(t time.Time) Durations {
	values := append(durations, t)
	sortTimesAscending(values)
	return values
}

// RemoveItem removes the time at the specified index from the Durations collection.
// If the index is out of bounds, returns the unchanged collection.
// The resulting collection maintains chronological order.
func (duration Durations) RemoveItem(index int) Durations {
	if index < 0 || index >= len(duration) {
		return duration
	}
	values := append(duration[:index], duration[index+1:]...)
	sortTimesAscending(values)
	return values
}

// StringSlice converts the Durations collection to a slice of formatted time strings.
// Each time is formatted using the 24-hour format "HH:MM".
func (duration Durations) StringSlice() []string {
	strs := make([]string, len(duration))
	for i, d := range duration {
		strs[i] = d.Format("15:04")
	}
	return strs
}

// SumPairedDurations calculates the total duration between pairs of times in the Durations collection.
// Times are already maintained in ascending order by the Durations type, and durations
// are calculated between consecutive pairs (times[0]->times[1], times[2]->times[3], etc.).
//
// If the collection has an odd number of elements, time.Now() is appended to complete
// the final pair. For deterministic behavior in tests, use SumPairedDurationsWithNow
// to provide an explicit "now" value.
//
// Example usage:
//
//	times := Durations{
//	    time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC),  // 08:00
//	    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC), // 12:00
//	    time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC), // 13:00
//	    time.Date(2025, 1, 1, 17, 0, 0, 0, time.UTC), // 17:00
//	}
//	total := times.SumPairedDurations() // Returns 8 hours (4h + 4h)
//
// Special cases:
//   - Empty collection returns duration 0
//   - If a later time in a pair is before its earlier time, that pair contributes 0
//   - Times are already sorted, so order of addition doesn't affect the result
func SumPairedDurations(times Durations) time.Duration {
	return SumPairedDurationsWithNow(times, time.Now())
}

// SumPairedDurationsWithNow is like SumPairedDurations but accepts an explicit time
// to use when balancing an odd-length collection. This makes the function's behavior
// deterministic, which is especially useful in tests.
//
// The function works as follows:
//  1. Copy collection to avoid modifying the original
//  2. If odd number of times, append the provided 'now' value
//  3. Sum durations between consecutive pairs
//  4. Skip (contribute 0) any pair where end time <= start time
//
// Example with odd number of times:
//
//	times := Durations{
//	    time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC),  // 08:00
//	    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC), // 12:00
//	    time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC), // 13:00
//	}
//	now := time.Date(2025, 1, 1, 17, 0, 0, 0, time.UTC)
//	total := times.SumPairedDurationsWithNow(now) // Returns 8 hours
//
// The function is particularly useful when you need to:
//   - Calculate total working time from clock-in/clock-out pairs
//   - Sum durations between event start/end times
//   - Measure accumulated time spans in a deterministic way
func SumPairedDurationsWithNow(times Durations, now time.Time) time.Duration {
	if len(times) == 0 {
		return 0
	}

	tlist := make([]time.Time, len(times))
	copy(tlist, times)

	if len(tlist)%2 == 1 && !now.IsZero() {
		tlist = append(tlist, now)
	}

	// Sort the times ascending so pairing always takes the earlier time first
	// and later time second. This makes the pairing deterministic even when
	// the input order is arbitrary.
	sort.Slice(tlist, func(i, j int) bool { return tlist[i].Before(tlist[j]) })

	var total time.Duration
	for i := 0; i < len(tlist); i += 2 {
		start := tlist[i]
		if (i + 1) >= len(tlist) {
			break
		}
		end := tlist[i+1]
		d := end.Sub(start)
		if d > 0 {
			total += d
		}
	}
	return total
}

package timeutils

import (
	"reflect"
	"testing"
	"time"
)

var (
	t8am  = time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)
	t10am = time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	t12pm = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	t4pm  = time.Date(2025, 1, 1, 16, 0, 0, 0, time.UTC)
)

func TestDurations_Append(t *testing.T) {
	tests := []struct {
		name     string
		initial  Durations
		toAdd    time.Time
		expected Durations
	}{
		{
			name:     "append to empty",
			initial:  Durations{},
			toAdd:    t12pm,
			expected: Durations{t12pm},
		},
		{
			name:     "append later time",
			initial:  Durations{t8am, t10am},
			toAdd:    t12pm,
			expected: Durations{t8am, t10am, t12pm},
		},
		{
			name:     "append earlier time",
			initial:  Durations{t10am, t12pm},
			toAdd:    t8am,
			expected: Durations{t8am, t10am, t12pm},
		},
		{
			name:     "append middle time",
			initial:  Durations{t8am, t12pm},
			toAdd:    t10am,
			expected: Durations{t8am, t10am, t12pm},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initial.Append(tt.toAdd)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Append() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDurations_RemoveItem(t *testing.T) {
	tests := []struct {
		name     string
		initial  Durations
		index    int
		expected Durations
	}{
		{
			name:     "remove from empty",
			initial:  Durations{},
			index:    0,
			expected: Durations{},
		},
		{
			name:     "remove first",
			initial:  Durations{t8am, t10am, t12pm},
			index:    0,
			expected: Durations{t10am, t12pm},
		},
		{
			name:     "remove middle",
			initial:  Durations{t8am, t10am, t12pm},
			index:    1,
			expected: Durations{t8am, t12pm},
		},
		{
			name:     "remove last",
			initial:  Durations{t8am, t10am, t12pm},
			index:    2,
			expected: Durations{t8am, t10am},
		},
		{
			name:     "remove invalid negative",
			initial:  Durations{t8am, t10am},
			index:    -1,
			expected: Durations{t8am, t10am},
		},
		{
			name:     "remove invalid too large",
			initial:  Durations{t8am, t10am},
			index:    2,
			expected: Durations{t8am, t10am},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initial.RemoveItem(tt.index)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("RemoveItem(%d) = %v, want %v", tt.index, result, tt.expected)
			}
		})
	}
}

func TestDurations_StringSlice(t *testing.T) {
	tests := []struct {
		name     string
		times    Durations
		expected []string
	}{
		{
			name:     "empty durations",
			times:    Durations{},
			expected: []string{},
		},
		{
			name:     "single time",
			times:    Durations{t8am},
			expected: []string{"08:00"},
		},
		{
			name:     "multiple times",
			times:    Durations{t8am, t12pm, t4pm},
			expected: []string{"08:00", "12:00", "16:00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.times.StringSlice()
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("StringSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSumPairedDurationsWithNow_EvenPairs(t *testing.T) {
	loc := time.UTC
	t0 := time.Date(2025, 1, 1, 8, 0, 0, 0, loc)
	t1 := time.Date(2025, 1, 1, 12, 0, 0, 0, loc)
	t2 := time.Date(2025, 1, 1, 13, 0, 0, 0, loc)
	t3 := time.Date(2025, 1, 1, 17, 0, 0, 0, loc)

	got := SumPairedDurationsWithNow([]time.Time{t0, t1, t2, t3}, time.Now())
	want := 8 * time.Hour
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestSumPairedDurationsWithNow_OddAppendsNow(t *testing.T) {
	loc := time.UTC
	t0 := time.Date(2025, 1, 1, 8, 0, 0, 0, loc)
	t1 := time.Date(2025, 1, 1, 12, 0, 0, 0, loc)
	t2 := time.Date(2025, 1, 1, 13, 0, 0, 0, loc)
	now := time.Date(2025, 1, 1, 17, 0, 0, 0, loc)

	got := SumPairedDurationsWithNow([]time.Time{t0, t1, t2}, now)
	want := 8 * time.Hour
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestSumPairedDurationsWithNow_OddZeroNow(t *testing.T) {
	loc := time.UTC
	t0 := time.Date(2025, 1, 1, 8, 0, 0, 0, loc)
	t1 := time.Date(2025, 1, 1, 12, 0, 0, 0, loc)
	t2 := time.Date(2025, 1, 1, 13, 0, 0, 0, loc)
	now := time.Time{}

	got := SumPairedDurationsWithNow([]time.Time{t0, t1, t2}, now)
	want := 4 * time.Hour
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestSumPairedDurationsWithNow_Unordered(t *testing.T) {
	loc := time.UTC
	t0 := time.Date(2025, 1, 1, 15, 0, 0, 0, loc)
	t1 := time.Date(2025, 1, 1, 10, 0, 0, 0, loc)
	t2 := time.Date(2025, 1, 1, 7, 0, 0, 0, loc)
	now := time.Date(2025, 1, 1, 16, 0, 0, 0, loc)

	got := SumPairedDurationsWithNow([]time.Time{t0, t1, t2}, now)
	want := 4 * time.Hour
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

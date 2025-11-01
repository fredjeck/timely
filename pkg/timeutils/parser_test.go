package timeutils

import (
	"testing"
)

func TestParseTime_ValidExamples(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1", "01:00"},
		{"01", "01:00"},
		{"14", "14:00"},
		{"1400", "14:00"},
		{"14:00", "14:00"},
		{"730", "07:30"},
		{"7:30", "07:30"},
		{"0730", "07:30"},
	}

	for _, tt := range tests {
		got, err := ParseTime(tt.input)
		if err != nil {
			t.Fatalf("ParseTime(%q) returned error: %v", tt.input, err)
		}
		if got.Format("15:04") != tt.want {
			t.Fatalf("ParseTime(%q) = %s, want %s", tt.input, got.Format("15:04"), tt.want)
		}
	}
}

func TestParseTime_Invalid(t *testing.T) {
	invalid := []string{"14a00", "25:00", "14:60", ""}
	for _, s := range invalid {
		if _, err := ParseTime(s); err == nil {
			t.Fatalf("expected error for %q", s)
		}
	}
}

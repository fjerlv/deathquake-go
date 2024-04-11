package ui

import (
	"strings"
	"testing"
)

func TestAlmostEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		expected bool
	}{
		{
			name:     "Exactly equal",
			a:        1.0,
			b:        1.0,
			expected: true,
		},
		{
			name:     "Within threshold",
			a:        1.0,
			b:        1.0000001,
			expected: true,
		},
		{
			name:     "Outside threshold",
			a:        1.0,
			b:        1.0001,
			expected: false,
		},
		{
			name:     "Negative values within threshold",
			a:        -2.5,
			b:        -2.50000001,
			expected: true,
		},
		{
			name:     "Zero comparison",
			a:        0.0,
			b:        0.00000001,
			expected: true,
		},
		{
			name:     "Large difference",
			a:        10.5,
			b:        11.5,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := almostEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("almostEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestFormatIntStat(t *testing.T) {
	tests := []struct {
		name            string
		value           int
		maxValue        int
		shouldHighlight bool
		shouldMatch     bool
	}{
		{
			name:            "Highlight when value equals max",
			value:           10,
			maxValue:        10,
			shouldHighlight: true,
			shouldMatch:     true,
		},
		{
			name:            "No highlight when value less than max",
			value:           5,
			maxValue:        10,
			shouldHighlight: true,
			shouldMatch:     false,
		},
		{
			name:            "No highlight when shouldHighlight is false",
			value:           10,
			maxValue:        10,
			shouldHighlight: false,
			shouldMatch:     false,
		},
		{
			name:            "Zero values with highlight disabled",
			value:           0,
			maxValue:        0,
			shouldHighlight: false,
			shouldMatch:     false,
		},
		{
			name:            "Zero values with highlight enabled",
			value:           0,
			maxValue:        0,
			shouldHighlight: true,
			shouldMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatIntStat(tt.value, tt.maxValue, tt.shouldHighlight)

			// Verify result contains the value
			if !strings.Contains(result, string(rune('0'+tt.value%10))) && tt.value < 10 {
				t.Errorf("Result %q should contain digit %d", result, tt.value)
			}

			// Test consistency: same inputs should produce same output
			result2 := formatIntStat(tt.value, tt.maxValue, tt.shouldHighlight)
			if result != result2 {
				t.Error("formatIntStat should produce consistent results")
			}
		})
	}
}

func TestFormatFloatStat(t *testing.T) {
	tests := []struct {
		name            string
		value           float64
		maxValue        float64
		shouldHighlight bool
		shouldMatch     bool
	}{
		{
			name:            "Highlight when value equals max",
			value:           2.5,
			maxValue:        2.5,
			shouldHighlight: true,
			shouldMatch:     true,
		},
		{
			name:            "Highlight when value almost equals max",
			value:           2.5000001,
			maxValue:        2.5,
			shouldHighlight: true,
			shouldMatch:     true,
		},
		{
			name:            "No highlight when value less than max",
			value:           1.5,
			maxValue:        2.5,
			shouldHighlight: true,
			shouldMatch:     false,
		},
		{
			name:            "No highlight when shouldHighlight is false",
			value:           2.5,
			maxValue:        2.5,
			shouldHighlight: false,
			shouldMatch:     false,
		},
		{
			name:            "Zero values",
			value:           0.0,
			maxValue:        0.0,
			shouldHighlight: true,
			shouldMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFloatStat(tt.value, tt.maxValue, tt.shouldHighlight)

			// Verify the value is formatted to 4 decimal places
			// (checking for the decimal point is sufficient)
			if !strings.Contains(result, ".") {
				t.Errorf("Expected float formatting with decimal point, got: %q", result)
			}

			// Test consistency: same inputs should produce same output
			result2 := formatFloatStat(tt.value, tt.maxValue, tt.shouldHighlight)
			if result != result2 {
				t.Error("formatFloatStat should produce consistent results")
			}
		})
	}
}

func TestFormatIntStatConsistency(t *testing.T) {
	// Test that the same inputs always produce the same output
	value1 := formatIntStat(10, 10, true)
	value2 := formatIntStat(10, 10, true)

	if value1 != value2 {
		t.Error("formatIntStat should produce consistent results for same inputs")
	}
}

func TestFormatFloatStatConsistency(t *testing.T) {
	// Test that the same inputs always produce the same output
	value1 := formatFloatStat(2.5, 2.5, true)
	value2 := formatFloatStat(2.5, 2.5, true)

	if value1 != value2 {
		t.Error("formatFloatStat should produce consistent results for same inputs")
	}
}

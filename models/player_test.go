package models

import (
	"testing"
)

func TestCalculateScore14(t *testing.T) {
	tests := []struct {
		name          string
		score         float64
		drinkingCider bool
		expected      string
	}{
		// Cider mode tests
		{
			name:          "cider mode: zero score",
			score:         0.0,
			drinkingCider: true,
			expected:      "0.00 cider",
		},
		{
			name:          "cider mode: small positive score",
			score:         0.5,
			drinkingCider: true,
			expected:      "0.33 cider",
		},
		{
			name:          "cider mode: at 1 cider boundary",
			score:         1.515,
			drinkingCider: true,
			expected:      "1.00 cider",
		},
		{
			name:          "cider mode: just over 1 cider (plural)",
			score:         1.53,
			drinkingCider: true,
			expected:      "1.01 ciders",
		},
		{
			name:          "cider mode: medium score",
			score:         5.0,
			drinkingCider: true,
			expected:      "3.30 ciders",
		},
		{
			name:          "cider mode: high score",
			score:         17.0,
			drinkingCider: true,
			expected:      "11.22 ciders",
		},
		{
			name:          "cider mode: negative score",
			score:         -1.0,
			drinkingCider: true,
			expected:      "-0.66 cider",
		},
		{
			name:          "cider mode: fractional score",
			score:         3.75,
			drinkingCider: true,
			expected:      "2.48 ciders",
		},

		// Beer mode tests - Zero and empty results
		{
			name:          "beer mode: zero score",
			score:         0.0,
			drinkingCider: false,
			expected:      "",
		},
		{
			name:          "beer mode: very small score rounds to 0",
			score:         0.035,
			drinkingCider: false,
			expected:      "",
		},

		// Beer mode tests - Only sips (score < 1)
		{
			name:          "beer mode: exactly 1 sip",
			score:         1.0 / 14.0,
			drinkingCider: false,
			expected:      "1 sip",
		},
		{
			name:          "beer mode: 2 sips",
			score:         2.0 / 14.0,
			drinkingCider: false,
			expected:      "2 sips",
		},
		{
			name:          "beer mode: 7 sips",
			score:         0.5,
			drinkingCider: false,
			expected:      "7 sips",
		},
		{
			name:          "beer mode: 13 sips",
			score:         13.0 / 14.0,
			drinkingCider: false,
			expected:      "13 sips",
		},

		// Beer mode tests - Exactly 14 sips should round to 1 beer
		{
			name:          "beer mode: 14 sips rounds to 1 beer",
			score:         0.9999,
			drinkingCider: false,
			expected:      "1 beer",
		},

		// Beer mode tests - Only beers (whole numbers)
		{
			name:          "beer mode: exactly 1 beer",
			score:         1.0,
			drinkingCider: false,
			expected:      "1 beer",
		},
		{
			name:          "beer mode: exactly 2 beers",
			score:         2.0,
			drinkingCider: false,
			expected:      "2 beers",
		},
		{
			name:          "beer mode: exactly 10 beers",
			score:         10.0,
			drinkingCider: false,
			expected:      "10 beers",
		},
		{
			name:          "beer mode: exactly 17 beers",
			score:         17.0,
			drinkingCider: false,
			expected:      "17 beers",
		},

		// Beer mode tests - Beers and sips combinations
		{
			name:          "beer mode: 1 beer & 1 sip",
			score:         1.0 + 1.0/14.0,
			drinkingCider: false,
			expected:      "1 beer & 1 sip",
		},
		{
			name:          "beer mode: 1 beer & 2 sips",
			score:         1.0 + 2.0/14.0,
			drinkingCider: false,
			expected:      "1 beer & 2 sips",
		},
		{
			name:          "beer mode: 1 beer & 7 sips",
			score:         1.5,
			drinkingCider: false,
			expected:      "1 beer & 7 sips",
		},
		{
			name:          "beer mode: 2 beers & 1 sip",
			score:         2.0 + 1.0/14.0,
			drinkingCider: false,
			expected:      "2 beers & 1 sip",
		},
		{
			name:          "beer mode: 2 beers & 7 sips",
			score:         2.5,
			drinkingCider: false,
			expected:      "2 beers & 7 sips",
		},
		{
			name:          "beer mode: 5 beers & 3 sips",
			score:         5.0 + 3.0/14.0,
			drinkingCider: false,
			expected:      "5 beers & 3 sips",
		},
		{
			name:          "beer mode: 10 beers & 10 sips",
			score:         10.0 + 10.0/14.0,
			drinkingCider: false,
			expected:      "10 beers & 10 sips",
		},

		// Beer mode tests - Edge case where sips round to 14
		{
			name:          "beer mode: 1 beer & 13.5 sips rounds to 13 sips",
			score:         1.0 + 13.5/14.0,
			drinkingCider: false,
			expected:      "1 beer & 13 sips",
		},
		{
			name:          "beer mode: 5 beers & rounds to 14 sips",
			score:         5.9999,
			drinkingCider: false,
			expected:      "6 beers",
		},

		// Beer mode tests - Negative scores
		{
			name:          "beer mode: -1 score",
			score:         -1.0,
			drinkingCider: false,
			expected:      "",
		},
		{
			name:          "beer mode: -0.5 score",
			score:         -0.5,
			drinkingCider: false,
			expected:      "",
		},

		// Beer mode tests - Various fractional beers
		{
			name:          "beer mode: 0.25 beers (3-4 sips)",
			score:         0.25,
			drinkingCider: false,
			expected:      "4 sips",
		},
		{
			name:          "beer mode: 0.75 beers (10-11 sips)",
			score:         0.75,
			drinkingCider: false,
			expected:      "11 sips",
		},
		{
			name:          "beer mode: 3.14159 beers",
			score:         3.14159,
			drinkingCider: false,
			expected:      "3 beers & 2 sips",
		},
		{
			name:          "beer mode: 8.9 beers",
			score:         8.9,
			drinkingCider: false,
			expected:      "8 beers & 13 sips",
		},
		{
			name:          "beer mode: 16.5 beers",
			score:         16.5,
			drinkingCider: false,
			expected:      "16 beers & 7 sips",
		},
		{
			name:          "beer mode: 16.99 beers rounds to 17",
			score:         16.99,
			drinkingCider: false,
			expected:      "17 beers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateScore14(tt.score, tt.drinkingCider)
			if result != tt.expected {
				t.Errorf("calculateScore14(%v, %v) = %q, want %q",
					tt.score, tt.drinkingCider, result, tt.expected)
			}
		})
	}
}

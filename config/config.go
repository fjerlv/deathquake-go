package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds configuration for the game parser
type Config struct {
	// IgnoredPlayers is a list of player names that should be ignored
	// (not displayed in the UI or counted in statistics)
	IgnoredPlayers []string `json:"ignored_players"`

	// DrinkingCiderPlayers is a list of player names that have the
	// special "drinking cider" attribute
	DrinkingCiderPlayers []string `json:"drinking_cider_players"`

	// SkipGames is a list of game identifiers to skip
	SkipGames []string `json:"skip_games"`
}

// LoadFromFile loads configuration from a JSON file
func LoadFromFile(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &cfg, nil
}

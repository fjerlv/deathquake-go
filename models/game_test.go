package models

import (
	"io"
	"log"
	"testing"

	"github.com/fjerlv/deathquake-go/config"
)

func TestNewGame(t *testing.T) {
	cfg := &config.Config{
		IgnoredPlayers:       []string{"<world>"},
		DrinkingCiderPlayers: []string{"TestPlayer"},
	}
	logger := log.New(io.Discard, "", 0)
	game := NewGame(cfg, logger)

	// Verify Players map is initialized
	if game.Players == nil {
		t.Error("Expected Players map to be initialized, got nil")
	}

	// Verify Config is set
	if game.Config == nil {
		t.Error("Expected Config to be set, got nil")
	}
	if game.Config != cfg {
		t.Error("Expected Config to match the provided config")
	}

	// Verify initial warmup state
	if !game.IsWarmup {
		t.Error("Expected IsWarmup to be true for new game")
	}

	// Verify Players map is empty
	if len(game.Players) != 0 {
		t.Errorf("Expected Players map to be empty, got %d players", len(game.Players))
	}

	// Verify other fields have zero values
	if game.CurrentMapName != "" {
		t.Errorf("Expected CurrentMapName to be empty, got %s", game.CurrentMapName)
	}
	if game.MapChanges != 0 {
		t.Errorf("Expected MapChanges to be 0, got %d", game.MapChanges)
	}
	if game.CurrentRoundId != "" {
		t.Errorf("Expected CurrentRoundId to be empty, got %s", game.CurrentRoundId)
	}
}

func TestChangeMap(t *testing.T) {
	tests := []struct {
		name              string
		initialMap        string
		initialMapChanges int
		initialIsWarmup   bool
		newMap            string
		expectedMap       string
		expectedChanges   int
		expectedWarmup    bool
	}{
		{
			name:              "First map change",
			initialMap:        "",
			initialMapChanges: 0,
			initialIsWarmup:   true,
			newMap:            "q3dm17",
			expectedMap:       "q3dm17",
			expectedChanges:   1,
			expectedWarmup:    true,
		},
		{
			name:              "Second map change - ends warmup",
			initialMap:        "q3dm17",
			initialMapChanges: 1,
			initialIsWarmup:   true,
			newMap:            "q3dm6",
			expectedMap:       "q3dm6",
			expectedChanges:   2,
			expectedWarmup:    false,
		},
		{
			name:              "Same map - no change",
			initialMap:        "q3dm17",
			initialMapChanges: 1,
			initialIsWarmup:   true,
			newMap:            "q3dm17",
			expectedMap:       "q3dm17",
			expectedChanges:   1,
			expectedWarmup:    true,
		},
		{
			name:              "Third map change - warmup stays false",
			initialMap:        "q3dm6",
			initialMapChanges: 2,
			initialIsWarmup:   false,
			newMap:            "q3tourney2",
			expectedMap:       "q3tourney2",
			expectedChanges:   3,
			expectedWarmup:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard, "", 0)
			game := &Game{
				CurrentMapName: tt.initialMap,
				MapChanges:     tt.initialMapChanges,
				IsWarmup:       tt.initialIsWarmup,
				Logger:         logger,
			}

			game.NewMap(tt.newMap, "2024-12-07 14:30:00")

			if game.CurrentMapName != tt.expectedMap {
				t.Errorf("CurrentMapName: expected %s, got %s", tt.expectedMap, game.CurrentMapName)
			}
			if game.MapChanges != tt.expectedChanges {
				t.Errorf("MapChanges: expected %d, got %d", tt.expectedChanges, game.MapChanges)
			}
			if game.IsWarmup != tt.expectedWarmup {
				t.Errorf("IsWarmup: expected %v, got %v", tt.expectedWarmup, game.IsWarmup)
			}
		})
	}
}

func TestSetIsWarmup(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	game := &Game{IsWarmup: false, Logger: logger}

	game.IsWarmup = true
	if !game.IsWarmup {
		t.Error("Expected IsWarmup to be true")
	}

	game.IsWarmup = false
	if game.IsWarmup {
		t.Error("Expected IsWarmup to be false")
	}
}

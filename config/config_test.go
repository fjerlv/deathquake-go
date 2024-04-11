package config

import (
	"os"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	// Test loading from a valid JSON file
	testConfig := `{
  "ignored_players": ["<world>", "TestBot"],
  "drinking_cider_players": ["Player1", "Player2"],
  "skip_games": ["q3dm1", "q3dm2"]
}`

	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(testConfig)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := LoadFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if len(cfg.IgnoredPlayers) != 2 {
		t.Errorf("Expected 2 ignored players, got %d", len(cfg.IgnoredPlayers))
	}

	if cfg.IgnoredPlayers[0] != "<world>" || cfg.IgnoredPlayers[1] != "TestBot" {
		t.Errorf("Ignored players not loaded correctly: %v", cfg.IgnoredPlayers)
	}

	if len(cfg.DrinkingCiderPlayers) != 2 {
		t.Errorf("Expected 2 drinking cider players, got %d", len(cfg.DrinkingCiderPlayers))
	}

	if len(cfg.SkipGames) != 2 {
		t.Errorf("Expected 2 skip games, got %d", len(cfg.SkipGames))
	}

	if cfg.SkipGames[0] != "q3dm1" || cfg.SkipGames[1] != "q3dm2" {
		t.Errorf("Skip games not loaded correctly: %v", cfg.SkipGames)
	}
}

func TestLoadFromFile_NonExistent(t *testing.T) {
	// Test loading from a non-existent file returns error
	_, err := LoadFromFile("/non/existent/path.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	// Test loading from invalid JSON file
	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte("invalid json{}")); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	_, err = LoadFromFile(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

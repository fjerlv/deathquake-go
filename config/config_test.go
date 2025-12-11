package config

import (
	"os"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	// Test loading from a valid JSON file
	testConfig := `{
  "ignored_players": ["TestBot"],
  "drinking_cider_players": ["Player1", "Player2"],
  "ignored_rounds": ["5d41402abc4b2a76b9719d911017c592", "7d793037a0760186574b0282f2f435e7"]
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

	// <world> is automatically appended after the config is loaded
	if cfg.IgnoredPlayers[0] != "TestBot" || cfg.IgnoredPlayers[1] != "<world>" {
		t.Errorf("Ignored players not loaded correctly: %v", cfg.IgnoredPlayers)
	}

	if len(cfg.DrinkingCiderPlayers) != 2 {
		t.Errorf("Expected 2 drinking cider players, got %d", len(cfg.DrinkingCiderPlayers))
	}

	if len(cfg.IgnoredRounds) != 2 {
		t.Errorf("Expected 2 ignored rounds, got %d", len(cfg.IgnoredRounds))
	}

	if cfg.IgnoredRounds[0] != "5d41402abc4b2a76b9719d911017c592" || cfg.IgnoredRounds[1] != "7d793037a0760186574b0282f2f435e7" {
		t.Errorf("Ignored rounds not loaded correctly: %v", cfg.IgnoredRounds)
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

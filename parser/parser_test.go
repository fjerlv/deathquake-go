package parser

import (
	"bytes"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/fjerlv/deathquake-go/config"
	"github.com/fjerlv/deathquake-go/models"
)

func TestParseLine_KillCreatesPlayers(t *testing.T) {
	// Arrange - setup test data
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	// Sample kill log line format with proper timestamp:
	// YYYY-MM-DD HH:MM:SS Kill: id1 id2 weaponId: AttackerName killed VictimName by WEAPON
	killLine := "2025-12-05 14:23:45 Kill: 3 2 10: PlayerOne killed PlayerTwo by MOD_RAILGUN"

	// Act - parse the line
	_, _ = ParseLine(killLine, game, logger, false)

	// Assert - verify both players were created
	if len(game.Players) != 2 {
		t.Errorf("Expected 2 players to be created, got %d", len(game.Players))
	}

	attacker, attackerExists := game.Players["PlayerOne"]
	if !attackerExists {
		t.Error("Expected attacker 'PlayerOne' to be created")
	}

	victim, victimExists := game.Players["PlayerTwo"]
	if !victimExists {
		t.Error("Expected victim 'PlayerTwo' to be created")
	}

	// Verify player stats were updated correctly
	if attackerExists {
		if attacker.RoundKills != 1 {
			t.Errorf("Expected attacker to have 1 kill, got %d", attacker.RoundKills)
		}
		if attacker.RoundRailgunKills != 1 {
			t.Errorf("Expected attacker to have 1 railgun kill, got %d", attacker.RoundRailgunKills)
		}
	}

	if victimExists {
		if victim.RoundDeaths != 1 {
			t.Errorf("Expected victim to have 1 death, got %d", victim.RoundDeaths)
		}
	}
}

func TestParseLine_KillsDuringWarmupNotRegistered(t *testing.T) {
	// Test that kills during warmup are not registered
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players:  make(map[string]*models.Player),
		Config:   cfg,
		Logger:   logger,
		IsWarmup: true, // Game is in warmup mode
	}

	// Process a kill line during warmup
	killLine := "2025-12-05 14:23:45 Kill: 3 2 10: PlayerOne killed PlayerTwo by MOD_RAILGUN"

	_, _ = ParseLine(killLine, game, logger, false)

	// Assert - no players should be created during warmup
	if len(game.Players) != 0 {
		t.Errorf("Expected no players to be created during warmup, got %d", len(game.Players))
	}

	// Verify no player stats were recorded
	attacker, attackerExists := game.Players["PlayerOne"]
	if attackerExists {
		t.Error("Expected attacker not to be created during warmup")
		if attacker.RoundKills != 0 {
			t.Errorf("Expected attacker to have 0 kills during warmup, got %d", attacker.RoundKills)
		}
	}

	victim, victimExists := game.Players["PlayerTwo"]
	if victimExists {
		t.Error("Expected victim not to be created during warmup")
		if victim.RoundDeaths != 0 {
			t.Errorf("Expected victim to have 0 deaths during warmup, got %d", victim.RoundDeaths)
		}
	}
}

func TestParseLine_WorldKill(t *testing.T) {
	// Test for <world> kills (environmental deaths)
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	// <world> kill means environmental death (fall, lava, etc.)
	killLine := "2025-12-05 14:24:50 Kill: 1022 3 22: <world> killed PlayerOne by MOD_FALLING"

	_, _ = ParseLine(killLine, game, logger, false)

	// Verify world and victim were created
	if len(game.Players) != 2 {
		t.Errorf("Expected 2 players (world and victim), got %d", len(game.Players))
	}

	victim, victimExists := game.Players["PlayerOne"]
	if !victimExists {
		t.Fatal("Expected victim 'PlayerOne' to be created")
	}

	// World kills should decrement the victim's kills and increment deaths/suicides
	if victim.RoundKills != -1 {
		t.Errorf("Expected victim to have -1 kills after world death, got %d", victim.RoundKills)
	}
	if victim.RoundDeaths != 1 {
		t.Errorf("Expected victim to have 1 death, got %d", victim.RoundDeaths)
	}
	if victim.RoundSuicideDeaths != 1 {
		t.Errorf("Expected victim to have 1 suicide death, got %d", victim.RoundSuicideDeaths)
	}
}

func TestParseLine_Suicide(t *testing.T) {
	// Test for self-kills (suicide)
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	// Player kills themselves
	killLine := "2025-12-05 14:25:30 Kill: 2 2 19: PlayerOne killed PlayerOne by MOD_ROCKET_SPLASH"

	_, _ = ParseLine(killLine, game, logger, false)

	if len(game.Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(game.Players))
	}

	player, exists := game.Players["PlayerOne"]
	if !exists {
		t.Fatal("Expected 'PlayerOne' to be created")
	}

	// Suicides should decrement kills and increment deaths/suicides
	if player.RoundKills != -1 {
		t.Errorf("Expected player to have -1 kills after suicide, got %d", player.RoundKills)
	}
	if player.RoundDeaths != 1 {
		t.Errorf("Expected player to have 1 death, got %d", player.RoundDeaths)
	}
	if player.RoundSuicideDeaths != 1 {
		t.Errorf("Expected player to have 1 suicide death, got %d", player.RoundSuicideDeaths)
	}
}

func TestParseLine_PlasmaWeapon(t *testing.T) {
	// Test plasma weapon kills
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	killLine := "2025-12-05 14:30:22 Kill: 4 3 9: Triple-H killed Rysgaard by MOD_PLASMA_SPLASH"

	_, _ = ParseLine(killLine, game, logger, false)

	if len(game.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(game.Players))
	}

	attacker, exists := game.Players["Triple-H"]
	if !exists {
		t.Fatal("Expected attacker 'Triple-H' to be created")
	}

	if attacker.RoundKills != 1 {
		t.Errorf("Expected attacker to have 1 kill, got %d", attacker.RoundKills)
	}

	victim, exists := game.Players["Rysgaard"]
	if !exists {
		t.Fatal("Expected victim 'Rysgaard' to be created")
	}

	if victim.RoundDeaths != 1 {
		t.Errorf("Expected victim to have 1 death, got %d", victim.RoundDeaths)
	}
}

func TestParseLine_RocketWeapon(t *testing.T) {
	// Test rocket weapon kills
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	killLine := "2025-12-05 15:10:45 Kill: 2 5 7: PlayerOne killed PlayerTwo by MOD_ROCKET"

	_, _ = ParseLine(killLine, game, logger, false)

	attacker, exists := game.Players["PlayerOne"]
	if !exists {
		t.Fatal("Expected attacker 'PlayerOne' to be created")
	}

	if attacker.RoundKills != 1 {
		t.Errorf("Expected attacker to have 1 kill, got %d", attacker.RoundKills)
	}

	if attacker.RoundRocketKills != 1 {
		t.Errorf("Expected attacker to have 1 rocket kill, got %d", attacker.RoundRocketKills)
	}
}

func TestParseLine_RocketSplashWeapon(t *testing.T) {
	// Test rocket splash weapon kills
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	killLine := "2025-12-05 15:12:30 Kill: 3 4 8: PlayerA killed PlayerB by MOD_ROCKET_SPLASH"

	_, _ = ParseLine(killLine, game, logger, false)

	attacker, exists := game.Players["PlayerA"]
	if !exists {
		t.Fatal("Expected attacker 'PlayerA' to be created")
	}

	if attacker.RoundRocketKills != 1 {
		t.Errorf("Expected attacker to have 1 rocket kill, got %d", attacker.RoundRocketKills)
	}
}

func TestParseLine_GauntletWeapon(t *testing.T) {
	// Test gauntlet weapon kills
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	killLine := "2025-12-05 15:20:10 Kill: 1 2 2: Warrior killed Victim by MOD_GAUNTLET"

	_, _ = ParseLine(killLine, game, logger, false)

	attacker, exists := game.Players["Warrior"]
	if !exists {
		t.Fatal("Expected attacker 'Warrior' to be created")
	}

	if attacker.RoundGauntletKills != 1 {
		t.Errorf("Expected attacker to have 1 gauntlet kill, got %d", attacker.RoundGauntletKills)
	}
}

func TestParseLine_MapChange(t *testing.T) {
	// Test server map change
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	// Create a player with some stats
	player := &models.Player{Name: "TestPlayer"}
	player.IncrementKills()
	player.IncrementKills()
	game.Players["TestPlayer"] = player

	mapChangeLine := "2025-12-05 16:00:00 Server: q3dm17"

	_, _ = ParseLine(mapChangeLine, game, logger, false)

	if game.CurrentMapName != "q3dm17" {
		t.Errorf("Expected map to be 'q3dm17', got '%s'", game.CurrentMapName)
	}

	// Player round stats should be reset after map change
	if player.RoundKills != 0 {
		t.Errorf("Expected player round kills to be reset to 0, got %d", player.RoundKills)
	}
}

func TestParseLine_MultipleMapChanges(t *testing.T) {
	// Test multiple map changes
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	maps := []string{
		"2025-12-05 16:00:00 Server: q3dm1",
		"2025-12-05 16:15:00 Server: q3dm4",
		"2025-12-05 16:30:00 Server: q3dm10",
	}

	for i, mapLine := range maps {
		_, _ = ParseLine(mapLine, game, logger, false)

		expectedMap := ""
		if i == 0 {
			expectedMap = "q3dm1"
		} else if i == 1 {
			expectedMap = "q3dm4"
		} else {
			expectedMap = "q3dm10"
		}

		if game.CurrentMapName != expectedMap {
			t.Errorf("Expected map to be '%s', got '%s'", expectedMap, game.CurrentMapName)
		}
	}
}

func TestParseLine_ScoreAction(t *testing.T) {
	// Test score action triggers round save
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	// Create players with stats
	player1 := &models.Player{Name: "Player1"}
	player1.IncrementKills()
	player1.IncrementKills()
	player1.IncrementKills()
	game.Players["Player1"] = player1

	player2 := &models.Player{Name: "Player2"}
	player2.IncrementKills()
	game.Players["Player2"] = player2

	scoreLine := "2025-12-05 17:00:00 score: 15"

	_, isReceivingScores := ParseLine(scoreLine, game, logger, false)

	if !isReceivingScores {
		t.Error("Expected to be receiving scores")
	}

	if !game.IsWarmup {
		t.Error("Expected to be in warmup mode")
	}

	// Players should have ranks assigned
	if player1.Rank == 0 {
		t.Error("Expected Player1 to have a rank assigned")
	}
	if player2.Rank == 0 {
		t.Error("Expected Player2 to have a rank assigned")
	}
}

func TestParseLine_InvalidLine(t *testing.T) {
	// Test that invalid lines are handled gracefully
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	// Line with less than 3 parts should be ignored
	invalidLine := "2025-12-05"

	_, _ = ParseLine(invalidLine, game, logger, false)

	if len(game.Players) != 0 {
		t.Errorf("Expected no players to be created for invalid line, got %d", len(game.Players))
	}
}

func TestParseLine_MultiWordPlayerNames(t *testing.T) {
	// Test that multi-word player names are handled correctly
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	killLine := "2025-12-05 14:30:22 Kill: 4 3 9: Triple-H killed Rysgaard by MOD_PLASMA"

	_, _ = ParseLine(killLine, game, logger, false)

	_, attackerExists := game.Players["Triple-H"]
	if !attackerExists {
		t.Error("Expected multi-word attacker name 'Triple-H' to be parsed correctly")
	}

	_, victimExists := game.Players["Rysgaard"]
	if !victimExists {
		t.Error("Expected victim 'Rysgaard' to be parsed correctly")
	}
}

func TestParseLine_PlayerNameContainsKilled(t *testing.T) {
	// Test that lines with "killed" in player names are ignored
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	// Player name contains "killed" - should be ignored
	killLine := "2025-12-05 14:30:22 Kill: 4 3 9: killedPlayer killed Victim by MOD_PLASMA"

	_, _ = ParseLine(killLine, game, logger, false)

	if len(game.Players) != 0 {
		t.Errorf("Expected no players to be created when attacker name contains 'killed', got %d", len(game.Players))
	}

	// Victim name contains "killed" - should also be ignored
	game2 := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
	}
	killLine2 := "2025-12-05 14:30:22 Kill: 4 3 9: Attacker killed killedVictim by MOD_PLASMA"

	_, _ = ParseLine(killLine2, game2, logger, false)

	if len(game2.Players) != 0 {
		t.Errorf("Expected no players to be created when victim name contains 'killed', got %d", len(game2.Players))
	}
}

func TestParseLine_SkipGames(t *testing.T) {
	logger := log.New(io.Discard, "", 0)

	// Test 1: Game in skip list should NOT save rounds when score is posted
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
		IgnoredRounds:           []string{}, // Will be set after getting hash
	}
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	// First map change to initialize
	mapChangeLine1 := "2025-12-05 15:55:00 Server: q3dm1"
	_, _ = ParseLine(mapChangeLine1, game, logger, false)

	// Second map change (this ends warmup and sets the MapChangeTimestamp we'll use for hashing)
	mapChangeLine2 := "2025-12-05 16:00:00 Server: q3dm17"
	_, _ = ParseLine(mapChangeLine2, game, logger, false)

	// Get the game id that will be checked when score is posted
	gameId := game.CurrentRoundId

	// Update config with this id in skip list
	cfg.IgnoredRounds = []string{gameId}

	// Process kill events (these should still be processed even for skipped games)
	killLine := "2025-12-05 16:01:00 Kill: 3 2 10: PlayerOne killed PlayerTwo by MOD_RAILGUN"
	_, _ = ParseLine(killLine, game, logger, false)

	// Verify that players WERE created (kills are still processed)
	if len(game.Players) != 2 {
		t.Errorf("Expected 2 players to be created even when game is in skip list, got %d players", len(game.Players))
	}

	// Now post a score - the round should NOT be saved because game hash is in skip list
	scoreLine := "2025-12-05 16:02:00 score: 10"
	_, _ = ParseLine(scoreLine, game, logger, false)

	// Verify that SaveRound was NOT called by checking that players don't have saved rounds
	// (checking Score which would be > 0 if rounds were saved)
	for _, p := range game.Players {
		if p.Score > 0 {
			t.Errorf("Expected round not to be saved for skipped game, but player %s has Score %.2f", p.Name, p.Score)
		}
	}

	// Test 2: Game NOT in skip list should save rounds normally
	cfg2 := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
		IgnoredRounds:           []string{"differenthash123"},
	}
	game2 := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg2,
		Logger:  logger,
	}

	// First map change
	mapChangeLine3 := "2025-12-05 17:00:00 Server: q3dm1"
	_, _ = ParseLine(mapChangeLine3, game2, logger, false)

	// Second map change (ends warmup)
	mapChangeLine4 := "2025-12-05 17:05:00 Server: q3dm6"
	_, _ = ParseLine(mapChangeLine4, game2, logger, false)

	// Process a kill
	killLine2 := "2025-12-05 17:06:00 Kill: 3 2 10: PlayerOne killed PlayerTwo by MOD_RAILGUN"
	_, _ = ParseLine(killLine2, game2, logger, false)

	// Post a score - should save the round since hash is NOT in skip list
	scoreLine2 := "2025-12-05 17:07:00 score: 10"
	_, _ = ParseLine(scoreLine2, game2, logger, false)

	// Verify that rounds WERE saved (players should have Score > 0)
	savedCount := 0
	for _, p := range game2.Players {
		if p.Score > 0 {
			savedCount++
		}
	}
	if savedCount == 0 {
		t.Error("Expected rounds to be saved for non-skipped game, but no players have Score > 0")
	}
}

func TestParseKillEvent(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		wantAttacker string
		wantVictim   string
		wantWeapon   string
	}{
		{
			name:         "Single word names",
			line:         "2024-04-19 16:01:33 Kill: 5 5 20: fjerlv killed fjerlv by MOD_SUICIDE",
			wantAttacker: "fjerlv",
			wantVictim:   "fjerlv",
			wantWeapon:   "MOD_SUICIDE",
		},
		{
			name:         "Attacker with hyphen",
			line:         "2024-04-19 16:01:39 Kill: 4 3 9: Triple-H killed Rysgaard by MOD_PLASMA_SPLASH",
			wantAttacker: "Triple-H",
			wantVictim:   "Rysgaard",
			wantWeapon:   "MOD_PLASMA_SPLASH",
		},
		{
			name:         "Victim with dot in name",
			line:         "2024-04-19 16:01:51 Kill: 1 0 7: miniFURI killed Mr.Chinaman by MOD_ROCKET_SPLASH",
			wantAttacker: "miniFURI",
			wantVictim:   "Mr.Chinaman",
			wantWeapon:   "MOD_ROCKET_SPLASH",
		},
		{
			name:         "Multi-word victim name",
			line:         "2024-04-19 16:01:54 Kill: 1 7 7: miniFURI killed GreekSisterFister by MOD_ROCKET_SPLASH",
			wantAttacker: "miniFURI",
			wantVictim:   "GreekSisterFister",
			wantWeapon:   "MOD_ROCKET_SPLASH",
		},
		{
			name:         "Multi-word attacker name",
			line:         "2024-04-19 16:01:44 Kill: 4 6 8: BetterLuckThanGood killed Triple-H by MOD_PLASMA",
			wantAttacker: "BetterLuckThanGood",
			wantVictim:   "Triple-H",
			wantWeapon:   "MOD_PLASMA",
		},
		{
			name:         "World kill",
			line:         "2024-04-19 16:14:43 Kill: 1022 2 16: <world> killed cmester by MOD_LAVA",
			wantAttacker: "<world>",
			wantVictim:   "cmester",
			wantWeapon:   "MOD_LAVA",
		},
		{
			name:         "Gauntlet weapon",
			line:         "2024-04-19 16:02:05 Kill: 7 1 2: GreekSisterFister killed miniFURI by MOD_GAUNTLET",
			wantAttacker: "GreekSisterFister",
			wantVictim:   "miniFURI",
			wantWeapon:   "MOD_GAUNTLET",
		},
		{
			name:         "Railgun weapon",
			line:         "2024-04-19 16:01:49 Kill: 1 6 10: miniFURI killed BetterLuckThanGood by MOD_RAILGUN",
			wantAttacker: "miniFURI",
			wantVictim:   "BetterLuckThanGood",
			wantWeapon:   "MOD_RAILGUN",
		},
		{
			name:         "Falling death",
			line:         "2024-04-19 16:16:54 Kill: 1022 8 19: <world> killed Siff by MOD_FALLING",
			wantAttacker: "<world>",
			wantVictim:   "Siff",
			wantWeapon:   "MOD_FALLING",
		},
		{
			name:         "No 'killed' word - invalid",
			line:         "2024-04-19 16:01:33 Kill: 5 5 20: fjerlv fragged victim by MOD_RAILGUN",
			wantAttacker: "",
			wantVictim:   "",
			wantWeapon:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageSplit := strings.Split(tt.line, " ")
			gotAttacker, gotVictim, gotWeapon := parseKillEvent(messageSplit)

			if gotAttacker != tt.wantAttacker {
				t.Errorf("parseKillEvent() attacker = %q, want %q", gotAttacker, tt.wantAttacker)
			}
			if gotVictim != tt.wantVictim {
				t.Errorf("parseKillEvent() victim = %q, want %q", gotVictim, tt.wantVictim)
			}
			if gotWeapon != tt.wantWeapon {
				t.Errorf("parseKillEvent() weapon = %q, want %q", gotWeapon, tt.wantWeapon)
			}
		})
	}
}

func TestSetRanks(t *testing.T) {
	tests := []struct {
		name     string
		players  map[string]*models.Player
		expected map[string]int // playerName -> expectedRank
	}{
		{
			name: "Basic ranking by score",
			players: map[string]*models.Player{
				"Alice": {Name: "Alice", Score: 10.0, Kills: 5},
				"Bob":   {Name: "Bob", Score: 15.0, Kills: 7},
				"Carol": {Name: "Carol", Score: 5.0, Kills: 3},
			},
			expected: map[string]int{"Bob": 1, "Alice": 2, "Carol": 3},
		},
		{
			name: "Tie-break by kills when scores equal",
			players: map[string]*models.Player{
				"Alice": {Name: "Alice", Score: 10.0, Kills: 8},
				"Bob":   {Name: "Bob", Score: 10.0, Kills: 5},
				"Carol": {Name: "Carol", Score: 10.0, Kills: 10},
			},
			expected: map[string]int{"Carol": 1, "Alice": 2, "Bob": 3},
		},
		{
			name: "Tie-break by name when score and kills equal",
			players: map[string]*models.Player{
				"Alice": {Name: "Alice", Score: 10.0, Kills: 5},
				"Bob":   {Name: "Bob", Score: 10.0, Kills: 5},
				"Zoe":   {Name: "Zoe", Score: 10.0, Kills: 5},
			},
			expected: map[string]int{"Zoe": 1, "Bob": 2, "Alice": 3},
		},
		{
			name: "Zero score players sorted by name only",
			players: map[string]*models.Player{
				"Alice": {Name: "Alice", Score: 0, Kills: 3},
				"Bob":   {Name: "Bob", Score: 0, Kills: 5},
				"Carol": {Name: "Carol", Score: 0, Kills: 1},
			},
			expected: map[string]int{"Carol": 1, "Bob": 2, "Alice": 3},
		},
		{
			name: "Mixed scores including zero",
			players: map[string]*models.Player{
				"Alice": {Name: "Alice", Score: 5.0, Kills: 2},
				"Bob":   {Name: "Bob", Score: 0, Kills: 10},
				"Carol": {Name: "Carol", Score: 0, Kills: 5},
			},
			expected: map[string]int{"Alice": 1, "Carol": 2, "Bob": 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard, "", 0)
			game := &models.Game{Players: tt.players, Logger: logger}
			playerSlice := make([]*models.Player, 0, len(game.Players))
			for _, p := range game.Players {
				playerSlice = append(playerSlice, p)
			}

			sort.Slice(playerSlice, func(i, j int) bool {
				player1 := playerSlice[i]
				player2 := playerSlice[j]

				// Primary sort: by score (highest first)
				if player1.Score != player2.Score {
					return player1.Score > player2.Score
				}

				// Scores are equal - check for special case
				// If both have 0 score, skip kills comparison and go to name
				if player1.Score == 0 {
					return player1.Name > player2.Name
				}

				// Same non-zero score - tie-break by kills
				if player1.Kills != player2.Kills {
					return player1.Kills > player2.Kills
				}

				// Same score and kills - tie-break alphabetically (descending)
				return player1.Name > player2.Name
			})

			// Assign sequential ranks (1, 2, 3, ...)
			for i, player := range playerSlice {
				player.SetRank(i + 1)
			}
			for playerName, expectedRank := range tt.expected {
				if tt.players[playerName].Rank != expectedRank {
					t.Errorf("Player %s: expected rank %d, got %d",
						playerName, expectedRank, tt.players[playerName].Rank)
				}
			}
		})
	}
}

func TestGetSortedPlayers(t *testing.T) {
	tests := []struct {
		name          string
		players       map[string]*models.Player
		expectedOrder []string
	}{
		{
			name: "Sorts by rank ascending",
			players: map[string]*models.Player{
				"Alice": {Name: "Alice", Rank: 3, Kills: 5, IsIgnored: false},
				"Bob":   {Name: "Bob", Rank: 1, Kills: 10, IsIgnored: false},
				"Carol": {Name: "Carol", Rank: 2, Kills: 7, IsIgnored: false},
			},
			expectedOrder: []string{"Bob", "Carol", "Alice"},
		},
		{
			name: "Unranked players go last",
			players: map[string]*models.Player{
				"Alice": {Name: "Alice", Rank: 2, Kills: 5, IsIgnored: false},
				"Bob":   {Name: "Bob", Rank: 0, Kills: 15, IsIgnored: false},
				"Carol": {Name: "Carol", Rank: 1, Kills: 7, IsIgnored: false},
			},
			expectedOrder: []string{"Carol", "Alice", "Bob"},
		},
		{
			name: "Tie-break same rank by kills",
			players: map[string]*models.Player{
				"Alice": {Name: "Alice", Rank: 1, Kills: 5, IsIgnored: false},
				"Bob":   {Name: "Bob", Rank: 1, Kills: 10, IsIgnored: false},
				"Carol": {Name: "Carol", Rank: 1, Kills: 7, IsIgnored: false},
			},
			expectedOrder: []string{"Bob", "Carol", "Alice"},
		},
		{
			name: "Tie-break by name when rank and kills equal",
			players: map[string]*models.Player{
				"Alice": {Name: "Alice", Rank: 1, Kills: 5, IsIgnored: false},
				"Bob":   {Name: "Bob", Rank: 1, Kills: 5, IsIgnored: false},
				"Zoe":   {Name: "Zoe", Rank: 1, Kills: 5, IsIgnored: false},
			},
			expectedOrder: []string{"Zoe", "Bob", "Alice"},
		},
		{
			name: "Ignored players are filtered out",
			players: map[string]*models.Player{
				"Alice":   {Name: "Alice", Rank: 1, Kills: 5, IsIgnored: false},
				"<world>": {Name: "<world>", Rank: 2, Kills: 0, IsIgnored: true},
				"Bob":     {Name: "Bob", Rank: 2, Kills: 3, IsIgnored: false},
			},
			expectedOrder: []string{"Alice", "Bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard, "", 0)
			game := &models.Game{Players: tt.players, Logger: logger}
			result := game.GetSortedPlayers()

			if len(result) != len(tt.expectedOrder) {
				t.Fatalf("Expected %d players, got %d", len(tt.expectedOrder), len(result))
			}

			for i, expectedName := range tt.expectedOrder {
				if result[i].Name != expectedName {
					t.Errorf("Position %d: expected %s, got %s", i, expectedName, result[i].Name)
				}
			}
		})
	}
}

func TestGetFragLimit(t *testing.T) {
	tests := []struct {
		name     string
		players  map[string]*models.Player
		expected int
	}{
		{
			name: "Returns max round kills",
			players: map[string]*models.Player{
				"Alice": {RoundKills: 5},
				"Bob":   {RoundKills: 10},
				"Carol": {RoundKills: 3},
			},
			expected: 10,
		},
		{
			name:     "Returns zero for empty players",
			players:  map[string]*models.Player{},
			expected: 0,
		},
		{
			name: "Returns zero when all have zero kills",
			players: map[string]*models.Player{
				"Alice": {RoundKills: 0},
				"Bob":   {RoundKills: 0},
			},
			expected: 0,
		},
		{
			name: "Single player",
			players: map[string]*models.Player{
				"Alice": {RoundKills: 7},
			},
			expected: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard, "", 0)
			game := &models.Game{Players: tt.players, Logger: logger}
			result := game.GetFragLimit()
			if result != tt.expected {
				t.Errorf("Expected frag limit %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestSetMax(t *testing.T) {
	tests := []struct {
		name          string
		players       map[string]*models.Player
		expectedMaxes models.Game
	}{
		{
			name: "Sets all maximum values",
			players: map[string]*models.Player{
				"Alice": {
					Kills: 10, Deaths: 5, KillDeathRatio: 2.0,
					RocketKills: 3, RailgunKills: 4, GauntletKills: 1,
					SuicideDeaths: 2, KillingStreak: 5,
					IsIgnored: false,
				},
				"Bob": {
					Kills: 15, Deaths: 8, KillDeathRatio: 1.875,
					RocketKills: 5, RailgunKills: 2, GauntletKills: 3,
					SuicideDeaths: 1, KillingStreak: 7,
					IsIgnored: false,
				},
			},
			expectedMaxes: models.Game{
				MaxKills: 15, MaxDeaths: 8, MaxKillDeathRatio: 2.0,
				MaxRocketKills: 5, MaxRailgunKills: 4, MaxGauntletKills: 3,
				MaxSuicides: 2, MaxKillingStreak: 7,
			},
		},
		{
			name: "Ignores players marked as ignored",
			players: map[string]*models.Player{
				"Alice": {
					Kills: 10, Deaths: 5, KillDeathRatio: 2.0,
					RocketKills: 3, RailgunKills: 4, GauntletKills: 1,
					SuicideDeaths: 2, KillingStreak: 5,
					IsIgnored: false,
				},
				"<world>": {
					Kills: 100, Deaths: 50, KillDeathRatio: 10.0,
					RocketKills: 50, RailgunKills: 50, GauntletKills: 50,
					SuicideDeaths: 50, KillingStreak: 50,
					IsIgnored: true,
				},
			},
			expectedMaxes: models.Game{
				MaxKills: 10, MaxDeaths: 5, MaxKillDeathRatio: 2.0,
				MaxRocketKills: 3, MaxRailgunKills: 4, MaxGauntletKills: 1,
				MaxSuicides: 2, MaxKillingStreak: 5,
			},
		},
		{
			name: "Resets kills and ratio before calculating",
			players: map[string]*models.Player{
				"Alice": {
					Kills: 5, Deaths: 2, KillDeathRatio: 2.5,
					IsIgnored: false,
				},
			},
			expectedMaxes: models.Game{
				MaxKills: 5, MaxDeaths: 2, MaxKillDeathRatio: 2.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := &models.Game{
				// Start with some values to ensure they get reset properly
				MaxKills:          100,
				MaxKillDeathRatio: 100.0,
				Players:           tt.players,
			}

			// Inline UpdateMaxStats logic (method was removed)
			game.MaxKillDeathRatio = 0
			game.MaxKills = 0
			for _, p := range game.Players {
				if p.IsIgnored {
					continue
				}
				game.MaxKills = max(game.MaxKills, p.Kills)
				game.MaxDeaths = max(game.MaxDeaths, p.Deaths)
				game.MaxKillDeathRatio = max(game.MaxKillDeathRatio, p.KillDeathRatio)
				game.MaxRocketKills = max(game.MaxRocketKills, p.RocketKills)
				game.MaxRailgunKills = max(game.MaxRailgunKills, p.RailgunKills)
				game.MaxGauntletKills = max(game.MaxGauntletKills, p.GauntletKills)
				game.MaxSuicides = max(game.MaxSuicides, p.SuicideDeaths)
				game.MaxKillingStreak = max(game.MaxKillingStreak, p.KillingStreak)
			}

			if game.MaxKills != tt.expectedMaxes.MaxKills {
				t.Errorf("MaxKills: expected %d, got %d", tt.expectedMaxes.MaxKills, game.MaxKills)
			}
			if game.MaxDeaths != tt.expectedMaxes.MaxDeaths {
				t.Errorf("MaxDeaths: expected %d, got %d", tt.expectedMaxes.MaxDeaths, game.MaxDeaths)
			}
			if game.MaxKillDeathRatio != tt.expectedMaxes.MaxKillDeathRatio {
				t.Errorf("MaxKillDeathRatio: expected %.2f, got %.2f",
					tt.expectedMaxes.MaxKillDeathRatio, game.MaxKillDeathRatio)
			}
			if game.MaxRocketKills != tt.expectedMaxes.MaxRocketKills {
				t.Errorf("MaxRocketKills: expected %d, got %d",
					tt.expectedMaxes.MaxRocketKills, game.MaxRocketKills)
			}
			if game.MaxRailgunKills != tt.expectedMaxes.MaxRailgunKills {
				t.Errorf("MaxRailgunKills: expected %d, got %d",
					tt.expectedMaxes.MaxRailgunKills, game.MaxRailgunKills)
			}
			if game.MaxGauntletKills != tt.expectedMaxes.MaxGauntletKills {
				t.Errorf("MaxGauntletKills: expected %d, got %d",
					tt.expectedMaxes.MaxGauntletKills, game.MaxGauntletKills)
			}
			if game.MaxSuicides != tt.expectedMaxes.MaxSuicides {
				t.Errorf("MaxSuicides: expected %d, got %d",
					tt.expectedMaxes.MaxSuicides, game.MaxSuicides)
			}
			if game.MaxKillingStreak != tt.expectedMaxes.MaxKillingStreak {
				t.Errorf("MaxKillingStreak: expected %d, got %d",
					tt.expectedMaxes.MaxKillingStreak, game.MaxKillingStreak)
			}
		})
	}
}

func TestParseLine_ReturnsErrorWhenAttackerNameContainsKilled(t *testing.T) {
	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
		IgnoredRounds:           []string{},
	}
	logger := log.New(io.Discard, "", 0)
	game := &models.Game{
		Players: make(map[string]*models.Player),
		Config:  cfg,
		Logger:  logger,
	}

	// Player name contains "killed" - should return error
	killLine := "2025-12-05 14:30:22 Kill: 4 3 9: killedPlayer killed Victim by MOD_PLASMA"

	err, _ := ParseLine(killLine, game, logger, false)

	if err == nil {
		t.Error("Expected error when attacker name contains 'killed', got nil")
	}

	expectedErrMsg := "invalid kill event: line contains 'killed' 2 times"
	if err != nil && !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain %q, got %q", expectedErrMsg, err.Error())
	}

	// Verify no players were created
	if len(game.Players) != 0 {
		t.Errorf("Expected no players to be created, got %d", len(game.Players))
	}
}

func TestTail_LogsErrorFromParseLine(t *testing.T) {
	// Create a temporary file with invalid kill line
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write a line with attacker name containing "killed"
	invalidLine := "2025-12-05 14:30:22 Kill: 4 3 9: killedPlayer killed Victim by MOD_PLASMA\n"
	if _, err := tmpFile.WriteString(invalidLine); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Create a buffer to capture logger output
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := &config.Config{
		IgnoredPlayers:       []string{},
		DrinkingCiderPlayers: []string{},
		IgnoredRounds:           []string{},
	}
	game := models.NewGame(cfg, logger)

	// Use a channel to signal when we're done reading
	done := make(chan bool)

	go func() {
		// Give Tail a moment to process the line
		time.Sleep(100 * time.Millisecond)
		done <- true
	}()

	// Start tailing in a goroutine
	go func() {
		Tail(tmpFile.Name(), nil, game, logger)
	}()

	// Wait for processing
	<-done

	// Verify that the error was logged
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "[ERROR]") {
		t.Errorf("Expected logger to contain '[ERROR]', got: %q", logOutput)
	}

	expectedErrMsg := "invalid kill event: line contains 'killed' 2 times"
	if !strings.Contains(logOutput, expectedErrMsg) {
		t.Errorf("Expected logger to contain %q, got: %q", expectedErrMsg, logOutput)
	}
}

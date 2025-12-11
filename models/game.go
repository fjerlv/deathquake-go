package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"sort"

	"github.com/fjerlv/deathquake-go/config"
)

type Game struct {
	// Players in the game
	Players map[string]*Player

	// Configuration
	Config *config.Config

	// Logger
	Logger *log.Logger

	// Game state
	CurrentRoundId string // Will only be assigned for the next map
	IsWarmup       bool
	CurrentMapName string
	MapChanges     int

	// Maximum statistics tracking
	MaxKills          int
	MaxDeaths         int
	MaxKillDeathRatio float64
	MaxKillingStreak  int
	MaxRocketKills    int
	MaxRailgunKills   int
	MaxGauntletKills  int
	MaxSuicides       int
}

// Constructor

// NewGame creates and initializes a new Game instance
func NewGame(cfg *config.Config, logger *log.Logger) *Game {
	logger.Printf("[GAME] Initializing new game")
	return &Game{
		Players:  make(map[string]*Player),
		Config:   cfg,
		Logger:   logger,
		IsWarmup: true,
	}
}

// Player Operations

// GetOrCreatePlayer retrieves an existing player or creates a new one if not found
func (g *Game) GetOrCreatePlayer(playerName string) *Player {
	if player, ok := g.Players[playerName]; ok {
		return player
	}

	g.Logger.Printf("[%s] [PLAYER] Creating new player: %s", g.CurrentRoundId, playerName)

	newPlayer := &Player{
		Name: playerName,
	}

	for _, c := range g.Config.IgnoredPlayers {
		if newPlayer.Name == c {
			newPlayer.SetIsIgnored(true)
			g.Logger.Printf("[%s] [PLAYER] Player %s marked as ignored", g.CurrentRoundId, playerName)
			break
		}
	}

	for _, c := range g.Config.DrinkingCiderPlayers {
		if newPlayer.Name == c {
			newPlayer.SetDrinkingCider(true)
			g.Logger.Printf("[%s] [PLAYER] Player %s marked as drinking cider", g.CurrentRoundId, playerName)
			break
		}
	}

	g.Players[playerName] = newPlayer
	return newPlayer
}

// GetSortedPlayers returns non-ignored players sorted for UI display
// Sorting priority:
// 1. Ranked players before unranked (rank 0 means not yet ranked)
// 2. Lower rank number first (1st place before 2nd place)
// 3. Tie-breaker: more kills first
// 4. Tie-breaker: alphabetical by name (descending)
func (g *Game) GetSortedPlayers() []*Player {
	playersAsSlice := make([]*Player, 0, len(g.Players))
	for _, player := range g.Players {
		if player.IsIgnored {
			continue
		}
		playersAsSlice = append(playersAsSlice, player)
	}

	sort.Slice(playersAsSlice, func(i, j int) bool {
		player1 := playersAsSlice[i]
		player2 := playersAsSlice[j]

		// Handle unranked players (rank 0) - they go to the bottom
		if player1.Rank == 0 {
			return false // player1 goes after player2
		}
		if player2.Rank == 0 {
			return true // player1 goes before player2
		}

		// Both players are ranked - compare ranks
		if player1.Rank != player2.Rank {
			return player1.Rank < player2.Rank // Lower rank number = better placement
		}

		// Same rank - tie-break by kills
		if player1.Kills != player2.Kills {
			return player1.Kills > player2.Kills // More kills = better
		}

		// Same kills - tie-break alphabetically (descending)
		return player1.Name > player2.Name
	})

	return playersAsSlice
}

// GetFragLimit returns the maximum round kills across all players
func (g *Game) GetFragLimit() int {
	var fragLimit int
	for _, p := range g.Players {
		fragLimit = max(p.RoundKills, fragLimit)
	}
	return fragLimit
}

// Game Actions

// NewMap updates the map and handles warmup state transitions
func (g *Game) NewMap(newMapName string, timestamp string) *Game {
	if newMapName != g.CurrentMapName {
		g.Logger.Printf("[%s] [MAP] Changing map from '%s' to '%s'", g.CurrentRoundId, g.CurrentMapName, newMapName)
		g.CurrentMapName = newMapName
		g.MapChanges++

		hash := md5.Sum([]byte(timestamp))
		g.CurrentRoundId = hex.EncodeToString(hash[:])
		g.Logger.Printf("[%s] [MAP] Round ID generated: %s", g.CurrentRoundId, g.CurrentRoundId)

		// After first map change, warmup is over
		if g.MapChanges > 1 {
			g.IsWarmup = false
			g.Logger.Printf("[%s] [MAP] Warmup ended, game is now active", g.CurrentRoundId)
		} else {
			g.Logger.Printf("[%s] [MAP] First map change, still in warmup", g.CurrentRoundId)
		}
	}

	// Discard round variables for all players
	g.Logger.Printf("[%s] [MAP] Discarding round stats for %d players", g.CurrentRoundId, len(g.Players))
	for _, p := range g.Players {
		p.DiscardRound()
	}

	return g
}

// RecordKill records a kill event between two players
// Handles world kills, suicides, and normal kills with weapon tracking
func (g *Game) RecordKill(attackerName, victimName, weapon string) *Game {
	if g.IsWarmup {
		g.Logger.Printf("[%s] [KILL] Ignoring kill during warmup: %s killed %s", g.CurrentRoundId, attackerName, victimName)
		return g
	}

	attacker := g.GetOrCreatePlayer(attackerName)
	victim := g.GetOrCreatePlayer(victimName)

	if attacker.Name == "<world>" || attacker.Name == victim.Name {
		// World kills and suicides both penalize the victim/player
		if attacker.Name == "<world>" {
			g.Logger.Printf("[%s] [KILL] World kill: %s died to environment", g.CurrentRoundId, victimName)
		} else {
			g.Logger.Printf("[%s] [KILL] Suicide: %s killed themselves", g.CurrentRoundId, victimName)
		}
		victim.SubtractKills()
		victim.IncrementDeaths()
		victim.IncrementSuicideDeaths()
	} else {
		// Normal kill
		g.Logger.Printf("[%s] [KILL] %s killed %s with %s", g.CurrentRoundId, attackerName, victimName, weapon)
		attacker.IncrementKills()
		victim.IncrementDeaths()

		// Track weapon-specific kills
		switch weapon {
		case "MOD_ROCKET", "MOD_ROCKET_SPLASH":
			attacker.IncrementRocketKills()
		case "MOD_RAILGUN":
			attacker.IncrementRailgunKills()
		case "MOD_GAUNTLET":
			attacker.IncrementGauntletKills()
		}
	}
	return g
}

// Save saves the current round for all players
func (g *Game) Save() *Game {
	g.Logger.Printf("[%s] [SAVE] Saving round results", g.CurrentRoundId)

	fragLimit := g.GetFragLimit()
	g.Logger.Printf("[%s] [SAVE] Frag limit for this round: %d", g.CurrentRoundId, fragLimit)

	for _, p := range g.Players {
		p.SaveRound(fragLimit)
	}

	playerSlice := make([]*Player, 0, len(g.Players))
	for _, p := range g.Players {
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
	g.Logger.Printf("[%s] [SAVE] Assigning ranks to %d players", g.CurrentRoundId, len(playerSlice))
	for i, player := range playerSlice {
		player.SetRank(i + 1)
		g.Logger.Printf("[%s] [RANK] #%d: %s (score: %.2f, kills: %d)",
			g.CurrentRoundId, i+1, player.Name, player.Score, player.Kills)
	}

	// Update the max values for the whole game
	g.Logger.Printf("[%s] [SAVE] Updating maximum statistics", g.CurrentRoundId)
	g.MaxKillDeathRatio = 0
	g.MaxKills = 0
	for _, p := range g.Players {
		if p.IsIgnored {
			continue
		}
		g.MaxKills = max(g.MaxKills, p.Kills)
		g.MaxDeaths = max(g.MaxDeaths, p.Deaths)
		g.MaxKillDeathRatio = max(g.MaxKillDeathRatio, p.KillDeathRatio)
		g.MaxRocketKills = max(g.MaxRocketKills, p.RocketKills)
		g.MaxRailgunKills = max(g.MaxRailgunKills, p.RailgunKills)
		g.MaxGauntletKills = max(g.MaxGauntletKills, p.GauntletKills)
		g.MaxSuicides = max(g.MaxSuicides, p.SuicideDeaths)
		g.MaxKillingStreak = max(g.MaxKillingStreak, p.KillingStreak)
	}

	g.IsWarmup = true
	g.Logger.Printf("[%s] [SAVE] Round saved, entering warmup mode", g.CurrentRoundId)
	return g
}

// Utility Functions

// IsSkipped returns true if the game's hash is in the skip list
func (g *Game) IsSkipped() bool {
	for _, gameId := range g.Config.IgnoredRounds {
		if g.CurrentRoundId == gameId {
			g.Logger.Printf("[%s] [SKIP] Round is in ignored list, skipping", g.CurrentRoundId)
			return true
		}
	}
	return false
}

// Print returns a formatted string with game information for logging
func (g *Game) Print() string {
	return fmt.Sprintf("%s", g.CurrentRoundId)
}

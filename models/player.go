package models

import (
	"encoding/json"
	"fmt"
	"math"
)

type Player struct {
	// Player identity
	Name string

	// Ranking
	Rank     int
	PrevRank int

	// Scores
	Score   float64
	Diff    float64
	Score14 string
	Diff14  string

	// Overall stats
	Kills          int
	Deaths         int
	KillDeathRatio float64

	// Round stats
	RoundKills  int
	RoundDeaths int

	// Weapon-specific kills
	RocketKills      int
	RoundRocketKills int

	RailgunKills      int
	RoundRailgunKills int

	GauntletKills      int
	RoundGauntletKills int

	// Special deaths
	SuicideDeaths      int
	RoundSuicideDeaths int

	// Killing streaks
	KillingStreak             int
	RoundKillingStreak        int
	RoundCurrentKillingStreak int

	// Flags
	IsDrinkingCider bool
	IsIgnored       bool
}

func formatCiders(score float64) string {
	ciders := (score / 0.5) * 0.33
	if ciders > 1 {
		return fmt.Sprintf("%.2f ciders", ciders)
	}
	return fmt.Sprintf("%.2f cider", ciders)
}

func calculateScore14(score float64, drinkingCider bool) string {
	if drinkingCider {
		return formatCiders(score)
	}
	return formatBeersAndSips(score)
}

func formatBeersAndSips(score float64) string {
	beers := int(math.Floor(score))
	sips := int(math.Round((score - float64(beers)) * 14))

	// Handle sips overflow (14 sips = 1 beer)
	if sips == 14 {
		beers++
		sips = 0
	}

	// Empty result for zero or negative scores
	if beers <= 0 && sips == 0 {
		return ""
	}
	if beers < 0 {
		return ""
	}

	// Only sips
	if beers == 0 {
		return formatSips(sips)
	}

	// Only beers
	if sips == 0 {
		return formatBeers(beers)
	}

	// Beers and sips
	return formatBeers(beers) + " & " + formatSips(sips)
}

func formatBeers(count int) string {
	if count == 1 {
		return "1 beer"
	}
	return fmt.Sprintf("%d beers", count)
}

func formatSips(count int) string {
	if count == 1 {
		return "1 sip"
	}
	return fmt.Sprintf("%d sips", count)
}

// Stats calculation

func (p *Player) RecalculateKillDeathRatio() *Player {
	if p.IsIgnored {
		return p
	}

	if p.Deaths == 0 {
		p.KillDeathRatio = float64(p.Kills)
	} else {
		p.KillDeathRatio = float64(p.Kills) / (float64(p.Deaths) - float64(p.SuicideDeaths))
	}

	return p
}

// Kill tracking

func (p *Player) IncrementKills() *Player {
	if p.IsIgnored {
		return p
	}

	p.RoundKills++
	p.RoundCurrentKillingStreak++
	p.RoundKillingStreak = max(p.RoundKillingStreak, p.RoundCurrentKillingStreak)

	return p
}

func (p *Player) SubtractKills() *Player {
	if p.IsIgnored {
		return p
	}

	p.RoundCurrentKillingStreak = 0
	p.RoundKills--

	return p
}

// Weapon-specific kills

func (p *Player) IncrementRocketKills() *Player {
	if p.IsIgnored {
		return p
	}
	p.RoundRocketKills++
	return p
}

func (p *Player) IncrementRailgunKills() *Player {
	if p.IsIgnored {
		return p
	}
	p.RoundRailgunKills++
	return p
}

func (p *Player) IncrementGauntletKills() *Player {
	if p.IsIgnored {
		return p
	}
	p.RoundGauntletKills++
	return p
}

// Death tracking

func (p *Player) IncrementDeaths() *Player {
	if p.IsIgnored {
		return p
	}

	p.RoundCurrentKillingStreak = 0
	p.RoundDeaths++

	return p
}

func (p *Player) IncrementSuicideDeaths() *Player {
	if p.IsIgnored {
		return p
	}

	p.RoundCurrentKillingStreak = 0
	p.RoundSuicideDeaths++

	return p
}

// Round management

func (p *Player) SaveRound(fragLimit int) *Player {
	if p.IsIgnored {
		return p
	}

	// Calculate score difference
	oldScore := p.Score
	diff := float64(p.RoundKills) / float64(fragLimit)
	p.Score += diff

	// Update formatted scores
	p.Score14 = calculateScore14(p.Score, p.IsDrinkingCider)
	p.Diff = diff
	p.Diff14 = calculateScore14(p.Score-oldScore, p.IsDrinkingCider)

	// Commit round stats to overall stats
	p.Kills += p.RoundKills
	p.Deaths += p.RoundDeaths
	p.RocketKills += p.RoundRocketKills
	p.RailgunKills += p.RoundRailgunKills
	p.GauntletKills += p.RoundGauntletKills
	p.SuicideDeaths += p.RoundSuicideDeaths
	p.KillingStreak = max(p.KillingStreak, p.RoundKillingStreak)

	// Reset round stats
	p.RoundKills = 0
	p.RoundDeaths = 0
	p.RoundRocketKills = 0
	p.RoundRailgunKills = 0
	p.RoundGauntletKills = 0
	p.RoundSuicideDeaths = 0

	p.RecalculateKillDeathRatio()

	return p
}

func (p *Player) DiscardRound() *Player {
	if p.IsIgnored {
		return p
	}

	p.RoundKills = 0
	p.RoundDeaths = 0
	p.RoundRocketKills = 0
	p.RoundRailgunKills = 0
	p.RoundGauntletKills = 0
	p.RoundSuicideDeaths = 0
	p.RoundKillingStreak = 0

	p.RecalculateKillDeathRatio()

	return p
}

// Player state setters

func (p *Player) SetDrinkingCider(b bool) *Player {
	p.IsDrinkingCider = b
	return p
}

func (p *Player) SetIsIgnored(b bool) *Player {
	p.IsIgnored = b
	return p
}

func (p *Player) SetRank(rank int) *Player {
	if p.IsIgnored {
		return p
	}

	p.PrevRank = p.Rank
	p.Rank = rank

	return p
}

// ToJson returns the JSON representation of the player state
func (p *Player) ToJson() string {
	playerJSON, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(playerJSON)
}

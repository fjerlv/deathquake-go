package parser

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fjerlv/deathquake-go/models"
	"github.com/fjerlv/deathquake-go/ui"
	"github.com/hpcloud/tail"
)

const (
	ActionKill   = "Kill:"
	ActionScore  = "score:"
	ActionServer = "Server:"
)

func Tail(fileName string, teaProgram *tea.Program, game *models.Game, logger *log.Logger) error {
	t, err := tail.TailFile(fileName, tail.Config{Follow: true})
	if err != nil {
		return err
	}

	receivingScores := false
	for line := range t.Lines {
		if err, receivingScores = ParseLine(line.Text, game, logger, receivingScores); err != nil {
			logger.Println("[ERROR]", err)
		}
		if teaProgram != nil {
			teaProgram.Send(
				ui.CreateGameUpdate(
					ui.GameUpdate{
						Players: game.GetSortedPlayers(),
						Game:    game,
					},
				),
			)
		}
	}
	return nil
}

func ParseLine(line string, game *models.Game, logger *log.Logger, receivingScores bool) (error, bool) {
	line = strings.Replace(line, "]\b \b", "", 1)
	messageSplit := strings.Split(line, " ")

	// Validate line format - need at least 3 parts
	if len(messageSplit) < 3 {
		return fmt.Errorf("invalid log line format: expected at least 3 parts, got %d: %q", len(messageSplit), line), receivingScores
	}

	timestamp := messageSplit[0] + " " + messageSplit[1]
	action := messageSplit[2]

	// Handle kill action
	if action == ActionKill {
		logger.Printf("[%s] [KILL] %s", game.CurentGameId, line)
		attackerName, victimName, weapon := parseKillEvent(messageSplit)
		if err := validateActionKill(attackerName, victimName); err != nil {
			return err, receivingScores
		}

		game.RecordKill(attackerName, victimName, weapon)
	} else if action == ActionServer {
		// Handle server/map change
		if len(messageSplit) >= 4 {
			newMapName := messageSplit[3]
			game.NewMap(newMapName, timestamp)
		}
	}

	// Update score state (handles both receiving and ending scores)
	if action == ActionScore {
		// First time receiving score line - save the round
		if !receivingScores && !game.IsWarmup {
			receivingScores = true
			if !game.IsSkipped() {
				game.Save(logger)
			}
		}
	} else {
		// If we were receiving scores and now got a different action, scores have ended
		if receivingScores {
			receivingScores = false
		}
	}

	return nil, receivingScores
}

func validateActionKill(attackerName string, victimName string) error {
	if attackerName == "" || victimName == "" {
		return fmt.Errorf("invalid kill event: empty player names (attacker: %q, victim: %q)", attackerName, victimName)
	}

	if strings.Contains(attackerName, "killed") || strings.Contains(victimName, "killed") {
		return fmt.Errorf("invalid kill event: player name contains 'killed' (attacker: %q, victim: %q)", attackerName, victimName)
	}
	return nil
}

// Utility functions

// parseKillEvent extracts attacker name, victim name, and weapon from a kill event
// Expected format: YYYY-MM-DD HH:MM:SS Kill: id1 id2 weaponId: AttackerName killed VictimName by WEAPON
// Returns empty strings if the format is invalid
func parseKillEvent(messageSplit []string) (attackerName, victimName, weapon string) {
	// Find the "killed" keyword index (only search once)
	killedIndex := -1
	for i, word := range messageSplit {
		if word == "killed" {
			killedIndex = i
			break
		}
	}

	// Invalid format - no "killed" found
	if killedIndex == -1 {
		return "", "", ""
	}

	// Weapon is always the last element (safe since we validated killedIndex exists)
	weapon = messageSplit[len(messageSplit)-1]

	// Build attacker name from index 6 to killedIndex
	// Index 6 is where player names start after: YYYY-MM-DD HH:MM:SS Kill: id1 id2 weaponId:
	if killedIndex > 6 {
		var attackerBuilder strings.Builder
		for i := 6; i < killedIndex; i++ {
			if i > 6 {
				attackerBuilder.WriteString(" ")
			}
			attackerBuilder.WriteString(messageSplit[i])
		}
		attackerName = attackerBuilder.String()
	}

	// Build victim name from killedIndex+1 to len-2 (excluding "by WEAPON")
	if killedIndex+1 < len(messageSplit)-2 {
		var victimBuilder strings.Builder
		for i := killedIndex + 1; i < len(messageSplit)-2; i++ {
			if i > killedIndex+1 {
				victimBuilder.WriteString(" ")
			}
			victimBuilder.WriteString(messageSplit[i])
		}
		victimName = victimBuilder.String()
	}

	return attackerName, victimName, weapon
}

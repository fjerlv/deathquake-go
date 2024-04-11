package ui

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/fjerlv/deathquake-go/models"
)

const (
	columnKeyRank           = "rank"
	columnKeyName           = "id"
	columnKeyScore          = "score"
	columnKeyScore14        = "score14"
	columnKeyDiff14         = "diff14"
	columnKeyKills          = "kills"
	columnKeyDeaths         = "deaths"
	columnKeyKillDeathRatio = "ratio"
	columnKeyRocket         = "rocket"
	columnKeyRailgun        = "railgun"
	columnKeyGauntlet       = "gauntlet"
	columnKeySuicide        = "suicide"
	columnKeyKillStreak     = "kill_streak"

	winningScore = 16
)

var (
	black       = lipgloss.Color("0")
	white       = lipgloss.Color("15")
	red         = lipgloss.Color("9")
	blue        = lipgloss.Color("4")
	gameWinner  = lipgloss.NewStyle().Background(red).Foreground(white).Bold(true)
	roundWinner = lipgloss.NewStyle().Background(blue).Foreground(white).Bold(true)
	highlight   = lipgloss.NewStyle().Background(black).Foreground(white)
	normal      = lipgloss.NewStyle()
)

type Model struct {
	title string
	table table.Model
}

type GameUpdate struct {
	Players []*models.Player
	Game    *models.Game
}

func NewModel() Model {
	return Model{
		table: table.New(generateColumns()),
	}
}

func CreateGameUpdate(update GameUpdate) tea.Msg {
	return update
}

const float64EqualityThreshold = 1e-5

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func formatIntStat(value, maxValue int, shouldHighlight bool) string {
	if shouldHighlight && value == maxValue {
		return highlight.Render(fmt.Sprintf("%d", value))
	}
	return normal.Render(fmt.Sprintf("%d", value))
}

func formatFloatStat(value, maxValue float64, shouldHighlight bool) string {
	if shouldHighlight && almostEqual(value, maxValue) {
		return highlight.Render(fmt.Sprintf("%.4f", value))
	}
	return normal.Render(fmt.Sprintf("%.4f", value))
}

func rankToString(rank int, prevRank int) string {
	if rank == 0 {
		return ""
	}
	if rank != prevRank && prevRank != 0 {
		rankDiff := prevRank - rank
		if rankDiff > 0 {
			return fmt.Sprintf("(+%d) %d", rankDiff, rank)
		} else {
			return fmt.Sprintf("(%d) %d", rankDiff, rank)
		}
	} else {
		return fmt.Sprintf("%d", rank)
	}
}

func generateColumns() []table.Column {
	return []table.Column{
		table.NewColumn(columnKeyRank, "Rank", 8),
		table.NewColumn(columnKeyName, "Name", 20),
		table.NewColumn(columnKeyScore, "Score", 10),
		table.NewColumn(columnKeyScore14, "Score 14", 20),
		table.NewColumn(columnKeyDiff14, "Diff 14", 12),
		table.NewColumn(columnKeyKills, "Kills", 8),
		table.NewColumn(columnKeyDeaths, "Deaths", 8),
		table.NewColumn(columnKeyKillDeathRatio, "Ratio", 8),
		table.NewColumn(columnKeyRocket, "Rocket Kills", 8),
		table.NewColumn(columnKeyRailgun, "Railgun Kills", 8),
		table.NewColumn(columnKeyGauntlet, "Gauntlet Kills", 8),
		table.NewColumn(columnKeyKillStreak, "Streak", 8),
		table.NewColumn(columnKeySuicide, "Suicide Deaths", 8),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			cmds = append(cmds, tea.Quit)
		}

	case GameUpdate:
		m.title = fmt.Sprintf("%s Deathquake", "💀")
		m.table = m.table.WithRows(generateRowsFromData(msg)).WithColumns(generateColumns())
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	body := strings.Builder{}
	title := lipgloss.NewStyle().MarginLeft(2).Bold(true).MarginTop(1)
	body.WriteString(title.Render(m.title))
	pad := lipgloss.NewStyle().Margin(1)
	body.WriteString(pad.Render(m.table.View()))
	return body.String()
}

func generateRowsFromData(update GameUpdate) []table.Row {
	var rows []table.Row

	for _, player := range update.Players {
		row := table.NewRow(table.RowData{
			columnKeyRank:           rankToString(player.Rank, player.PrevRank),
			columnKeyName:           player.Name,
			columnKeyScore:          fmt.Sprintf("%.4f", player.Score),
			columnKeyScore14:        player.Score14,
			columnKeyDiff14:         normal.Render(player.Diff14),
			columnKeyKills:          formatIntStat(player.Kills, update.Game.MaxKills, true),
			columnKeyDeaths:         formatIntStat(player.Deaths, update.Game.MaxDeaths, true),
			columnKeyKillDeathRatio: formatFloatStat(player.KillDeathRatio, update.Game.MaxKillDeathRatio, true),
			columnKeyRocket:         formatIntStat(player.RocketKills, update.Game.MaxRocketKills, true),
			columnKeyRailgun:        formatIntStat(player.RailgunKills, update.Game.MaxRailgunKills, true),
			columnKeyGauntlet:       formatIntStat(player.GauntletKills, update.Game.MaxGauntletKills, true),
			columnKeySuicide:        formatIntStat(player.SuicideDeaths, update.Game.MaxSuicides, true),
			columnKeyKillStreak:     formatIntStat(player.KillingStreak, update.Game.MaxKillingStreak, true),
		})

		if almostEqual(player.Diff, float64(1)) {
			row = row.WithStyle(roundWinner)
		}

		if player.Score > winningScore && player.Rank == 1 {
			row = row.WithStyle(gameWinner)
		}

		rows = append(rows, row)
	}

	return rows
}

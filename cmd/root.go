package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fjerlv/deathquake-go/config"
	"github.com/fjerlv/deathquake-go/models"
	"github.com/fjerlv/deathquake-go/parser"
	"github.com/fjerlv/deathquake-go/ui"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

var (
	filename string
	debug    bool
)

var rootCmd = &cobra.Command{
	Use:   "deathquake-go",
	Short: "Real-time Quake 3 Arena game statistics tracker",
	Long: `DeathQuake-Go monitors Quake 3 Arena game logs and displays
live player statistics, rankings, and match information in your terminal.

The tool tracks kills, deaths, weapon usage, killing streaks, and more,
with a fun beer/cider scoring system for match performance.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if filename is provided
		if filename == "" {
			log.Fatal("filename is required (use -f or --filename)")
		}

		// Check if file exists
		if _, err := os.Stat(filename); err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("file does not exist: %s", filename)
			}
			log.Fatalf("error accessing file %s: %v", filename, err)
		}

		// Load config.json from current directory
		cfg, err := config.LoadFromFile("config.json")
		if err != nil {
			log.Fatal(err)
		}

		// Create logger based on debug mode
		var logger *log.Logger
		if debug {
			logger = log.New(os.Stdout, "[DEBUG] ", log.Lshortfile)
		} else {
			logger = log.New(io.Discard, "", 0)
		}

		game := models.NewGame(cfg, logger)

		if debug {
			// Debug mode: run without UI
			if err := parser.Tail(filename, nil, game, logger); err != nil {
				log.Fatal(err)
			}
		} else {
			// Normal mode: run with tea UI
			program := tea.NewProgram(ui.NewModel())

			go func() {
				if err := parser.Tail(filename, program, game, logger); err != nil {
					log.Fatal(err)
				}
			}()

			if err := program.Start(); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&filename, "filename", "f", "", "Path to the Quake 3 game log file (required)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode")

	rootCmd.Example = `  # Monitor a game log (requires config.json in current directory)
  deathquake-go -f /path/to/games.log

  # Using relative path
  deathquake-go -f games.log`
}

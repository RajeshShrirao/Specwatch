package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
	"github.com/rajeshshrirao/specwatch/internal/tui"
	"github.com/rajeshshrirao/specwatch/internal/watcher"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Watch files and analyze on change",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		specPath := findSpecFile()
		if specPath == "" {
			fmt.Println("No spec.md found. Run 'specwatch init' to create one.")
			os.Exit(1)
		}

		rules, err := spec.Parse(specPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing spec: %v\n", err)
			os.Exit(1)
		}

		engine := analyzer.NewEngine(rules)

		watcher, err := watcher.NewWatcher()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating watcher: %v\n", err)
			os.Exit(1)
		}
		defer watcher.Close()

		model := tui.InitialModel()

		if err := watcher.Watch(path, func(file string) {
			violations, duration := engine.Analyze(file)
			newModel, _ := model.Update(tui.NewViolationMsg{
				File:       file,
				Violations: violations,
				Duration:   duration,
			})
			model = newModel.(tui.Model)
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting watcher: %v\n", err)
			os.Exit(1)
		}

		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
			os.Exit(1)
		}

		return nil
	},
}

package cmd

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/spec"
	"github.com/rajeshshrirao/specwatch/internal/tui"
	"github.com/rajeshshrirao/specwatch/internal/watcher"
	"github.com/spf13/cobra"
)

var (
	extensions []string
	debounce   int
	skip       []string
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
		engine.SkipCategories = skip

		w, err := watcher.NewWatcher(watcher.Options{
			Debounce:   time.Duration(debounce) * time.Millisecond,
			Extensions: extensions,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating watcher: %v\n", err)
			os.Exit(1)
		}
		defer w.Close()

		p := tea.NewProgram(tui.InitialModel())

		if err := w.Watch(path, func(file string) {
			violations, duration := engine.Analyze(file)
			p.Send(tui.NewViolationMsg{
				File:       file,
				Violations: violations,
				Duration:   duration,
			})
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting watcher: %v\n", err)
			os.Exit(1)
		}

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	watchCmd.Flags().StringSliceVarP(&extensions, "ext", "e", []string{"ts", "tsx", "go"}, "Watch specific extensions")
	watchCmd.Flags().IntVarP(&debounce, "debounce", "d", 800, "Debounce time in milliseconds")
	watchCmd.Flags().StringSliceVarP(&skip, "skip", "s", []string{}, "Skip rule categories")
}

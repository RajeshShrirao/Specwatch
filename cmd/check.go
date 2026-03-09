package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/reporter"
	"github.com/rajeshshrirao/specwatch/internal/spec"
	"github.com/spf13/cobra"
)

var format string

var checkCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "Check files once (for CI)",
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

		start := time.Now()
		violations, _ := engine.AnalyzeAll(path)
		duration := time.Since(start)

		if format == "" {
			if err := reporter.ConsoleOutput(violations, duration.Milliseconds()); err != nil {
				fmt.Fprintf(os.Stderr, "Error outputting results: %v\n", err)
				os.Exit(1)
			}
		} else {
			err := reporter.GenerateReport(violations, duration.Milliseconds(), format, os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
				os.Exit(1)
			}
		}

		os.Exit(reporter.ExitCode(violations))
		return nil
	},
}

func init() {
	checkCmd.Flags().StringVarP(&format, "format", "f", "", "Output format: json, text")
}

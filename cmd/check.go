package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		if err := loadRuntimeConfig(specPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		rules, err := spec.Parse(specPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing spec: %v\n", err)
			os.Exit(1)
		}

		engine := analyzer.NewEngine(rules)
		llmClient, _ := setupLLMClient(engine)
		if llmClient != nil {
			defer llmClient.Close()
		}

		start := time.Now()
		violations, err := runCheck(engine, rules, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking files: %v\n", err)
			os.Exit(1)
		}
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

func runCheck(engine *analyzer.Engine, rules *spec.RuleSet, root string) ([]analyzer.Violation, error) {
	var all []analyzer.Violation

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		staticViolations, _ := engine.Analyze(path)
		all = append(all, staticViolations...)

		if len(staticViolations) == 0 && len(rules.Architecture) > 0 && fileTouchesArchitectureRules(path, rules) {
			all = append(all, runArchitectureAI(engine, path, rules)...)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return all, nil
}

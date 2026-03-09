package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rajeshshrirao/specwatch/internal/analyzer"
)

type CheckResult struct {
	Violations []analyzer.Violation `json:"violations"`
	Summary    Summary              `json:"summary"`
}

type Summary struct {
	FilesChecked int   `json:"files_checked"`
	Errors       int   `json:"errors"`
	Warnings     int   `json:"warnings"`
	DurationMs   int64 `json:"duration_ms"`
}

func GenerateReport(violations []analyzer.Violation, durationMs int64, format string, writer interface{ Write([]byte) (int, error) }) error {
	errors := 0
	warnings := 0
	for _, v := range violations {
		if v.Severity == "error" {
			errors++
		} else {
			warnings++
		}
	}

	uniqueFiles := make(map[string]bool)
	for _, v := range violations {
		uniqueFiles[v.File] = true
	}

	result := CheckResult{
		Violations: violations,
		Summary: Summary{
			FilesChecked: len(uniqueFiles),
			Errors:       errors,
			Warnings:     warnings,
			DurationMs:   durationMs,
		},
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		_, err = writer.Write(data)
		return err

	case "text":
		tw := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintln(tw, "FILE\tLINE\tSEVERITY\tRULE\tEXCERPT"); err != nil {
			return err
		}
		for _, v := range violations {
			if _, err := fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t%s\n", v.File, v.Line, v.Severity, v.Rule, v.Excerpt); err != nil {
				return err
			}
		}
		if err := tw.Flush(); err != nil {
			return err
		}

		if _, err := fmt.Fprintf(writer, "\nSummary: %d files checked, %d errors, %d warnings, %dms\n",
			result.Summary.FilesChecked, result.Summary.Errors, result.Summary.Warnings, result.Summary.DurationMs); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func ExitCode(violations []analyzer.Violation) int {
	for _, v := range violations {
		if v.Severity == "error" {
			return 1
		}
	}
	return 0
}

func ConsoleOutput(violations []analyzer.Violation, durationMs int64) error {
	errors := 0
	warnings := 0

	fmt.Println()
	for _, v := range violations {
		if v.Severity == "error" {
			errors++
			fmt.Printf("✗ %s:%d [%s]\n", v.File, v.Line, v.Rule)
			fmt.Printf("  Found:    %s\n", v.Excerpt)
			fmt.Printf("  Fix:      %s\n\n", v.Suggestion)
		} else {
			warnings++
		}
	}

	uniqueFiles := make(map[string]bool)
	for _, v := range violations {
		uniqueFiles[v.File] = true
	}

	fmt.Printf("\n%d files checked, %d errors, %d warnings, %dms\n",
		len(uniqueFiles), errors, warnings, durationMs)

	return nil
}

func GetResult(violations []analyzer.Violation, durationMs int64) CheckResult {
	errors := 0
	warnings := 0
	for _, v := range violations {
		if v.Severity == "error" {
			errors++
		} else {
			warnings++
		}
	}

	uniqueFiles := make(map[string]bool)
	for _, v := range violations {
		uniqueFiles[v.File] = true
	}

	return CheckResult{
		Violations: violations,
		Summary: Summary{
			FilesChecked: len(uniqueFiles),
			Errors:       errors,
			Warnings:     warnings,
			DurationMs:   durationMs,
		},
	}
}

func WriteJSON(result CheckResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rajeshshrirao/specwatch/internal/analyzer"
	"github.com/rajeshshrirao/specwatch/internal/auth"
	"github.com/rajeshshrirao/specwatch/internal/llm"
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
		engine.SkipCategories = skip
		if !cmd.Flags().Changed("ext") {
			extensions = appConfig.Watch.Extensions
		}
		if !cmd.Flags().Changed("debounce") {
			debounce = appConfig.Watch.Debounce
		}

		llmClient, warned := setupLLMClient(engine)
		if llmClient != nil {
			defer func() { _ = llmClient.Close() }()
		}

		w, err := watcher.NewWatcher(watcher.Options{
			Debounce:   time.Duration(debounce) * time.Millisecond,
			Extensions: extensions,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating watcher: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = w.Close() }()

		p := tea.NewProgram(tui.InitialModel())
		saveCounter := 0

		if err := w.Watch(path, func(file string) {
			violations, duration := engine.Analyze(file)
			saveCounter++
			if shouldRunAIForWatch(saveCounter, violations, rules, file, warned) {
				aiViolations := runArchitectureAI(engine, file, rules)
				violations = append(violations, aiViolations...)
			}
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

func setupLLMClient(engine *analyzer.Engine) (llm.LLMClient, bool) {
	if !appConfig.LLM.Enabled {
		return nil, false
	}

	provider := llm.ProviderType(strings.ToLower(appConfig.LLM.Provider))
	if provider == "" {
		provider = llm.ProviderAnthropic
	}

	manager := auth.NewManager()
	if provider == llm.ProviderAnthropic && os.Getenv("ANTHROPIC_API_KEY") == "" {
		fmt.Println("AI checks disabled — set ANTHROPIC_API_KEY to enable")
		return nil, true
	}

	client, err := manager.CreateClient(provider, appConfig.LLM.Model)
	if err != nil {
		if provider == llm.ProviderAnthropic {
			fmt.Println("AI checks disabled — set ANTHROPIC_API_KEY to enable")
			return nil, true
		}
		return nil, false
	}

	engine.SetLLMClient(client)
	return client, false
}

func shouldRunAIForWatch(saveCounter int, violations []analyzer.Violation, rules *spec.RuleSet, file string, warned bool) bool {
	if warned || len(violations) > 0 || len(rules.Architecture) == 0 {
		return false
	}
	if saveCounter%10 != 0 {
		return false
	}
	return fileTouchesArchitectureRules(file, rules)
}

func runArchitectureAI(engine *analyzer.Engine, file string, rules *spec.RuleSet) []analyzer.Violation {
	if !engine.HasLLM() {
		return nil
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return nil
	}

	var all []analyzer.Violation
	for _, rule := range rules.Architecture {
		violations, err := engine.AnalyzeWithAI(context.Background(), file, string(content), rule.Description)
		if err != nil {
			continue
		}
		all = append(all, violations...)
	}

	return all
}

func fileTouchesArchitectureRules(file string, rules *spec.RuleSet) bool {
	content, err := os.ReadFile(file)
	if err != nil {
		return false
	}

	text := strings.ToLower(string(content))
	path := strings.ToLower(file)
	keywords := []string{
		"handler", "service", "repository", "component", "db", "database",
		"prisma", "mongoose", "sequelize", "import", "ui", "api",
	}

	for _, rule := range rules.Architecture {
		desc := strings.ToLower(rule.Description)
		for _, keyword := range keywords {
			if strings.Contains(desc, keyword) && (strings.Contains(text, keyword) || strings.Contains(path, keyword)) {
				return true
			}
		}
	}

	return false
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rajeshshrirao/specwatch/internal/llm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"

	appConfig = runtimeConfig{
		LLM: llmConfig{
			Provider: string(llm.ProviderAnthropic),
			Model:    llm.DefaultModels[llm.ProviderAnthropic],
		},
		Watch: watchConfig{
			Debounce:   800,
			Extensions: []string{"ts", "tsx", "go"},
		},
	}
)

type runtimeConfig struct {
	LLM   llmConfig
	Watch watchConfig
}

type llmConfig struct {
	Enabled  bool
	Provider string
	Model    string
}

type watchConfig struct {
	Debounce   int
	Extensions []string
}

var rootCmd = &cobra.Command{
	Use:   "specwatch",
	Short: "specwatch is a tool for watching and analyzing specs",
	Long:  `A fast, structured spec-driven static analysis tool for modern web development.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to specwatch! Use 'specwatch help' for more information.")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("specwatch version %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  date: %s\n", date)
		},
	})
}

func findSpecFile() string {
	paths := []string{"spec.md", "./spec.md", "../spec.md"}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func loadRuntimeConfig(specPath string) error {
	appConfig = runtimeConfig{
		LLM: llmConfig{
			Provider: string(llm.ProviderAnthropic),
			Model:    llm.DefaultModels[llm.ProviderAnthropic],
		},
		Watch: watchConfig{
			Debounce:   800,
			Extensions: []string{"ts", "tsx", "go"},
		},
	}

	configPath := filepath.Join(filepath.Dir(specPath), ".specwatch.yml")
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cfg := viper.New()
	cfg.SetConfigFile(configPath)
	cfg.SetConfigType("yaml")

	if err := cfg.ReadInConfig(); err != nil {
		return err
	}

	appConfig.LLM.Enabled = cfg.GetBool("llm.enabled")

	provider := strings.TrimSpace(cfg.GetString("llm.provider"))
	if provider != "" {
		appConfig.LLM.Provider = strings.ToLower(provider)
	}

	model := strings.TrimSpace(cfg.GetString("llm.model"))
	if model != "" {
		appConfig.LLM.Model = model
	}

	if cfg.IsSet("watch.debounce") {
		appConfig.Watch.Debounce = cfg.GetInt("watch.debounce")
	}

	if cfg.IsSet("watch.extensions") {
		appConfig.Watch.Extensions = cfg.GetStringSlice("watch.extensions")
	}

	return nil
}

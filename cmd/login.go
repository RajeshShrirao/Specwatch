package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rajeshshrirao/specwatch/internal/auth"
	"github.com/rajeshshrirao/specwatch/internal/llm"
	"github.com/spf13/cobra"
)

var (
	loginProvider string
	loginAPIKey   string
	loginModel    string
	loginList     bool
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure LLM provider authentication",
	Long: `Configure authentication for LLM providers (Anthropic, OpenRouter, Gemini).

Examples:
  # Set Anthropic API key
  specwatch login --provider anthropic --api-key sk-ant-...

  # Set OpenRouter API key  
  specwatch login --provider openrouter --api-key sk-or-v1-...

  # Set Gemini API key
  specwatch login --provider gemini --api-key AIza...

  # List available models for a provider
  specwatch login --provider anthropic --list-models`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().StringVar(&loginProvider, "provider", "", "LLM provider (anthropic, openrouter, gemini)")
	loginCmd.Flags().StringVar(&loginAPIKey, "api-key", "", "API key for the provider")
	loginCmd.Flags().StringVar(&loginModel, "model", "", "Model to use (default varies by provider)")
	loginCmd.Flags().BoolVar(&loginList, "list-models", false, "List available models for the provider")
}

func runLogin(cmd *cobra.Command, args []string) error {
	// If --list-models is used, show available models
	if loginList {
		return listModels()
	}

	// Validate provider
	if loginProvider == "" {
		return fmt.Errorf("provider is required (use --provider)")
	}

	provider := llm.ProviderType(strings.ToLower(loginProvider))
	if provider != llm.ProviderAnthropic && provider != llm.ProviderOpenRouter && provider != llm.ProviderGemini {
		return fmt.Errorf("invalid provider: %s (valid: anthropic, openrouter, gemini)", loginProvider)
	}

	// Get API key from flag or prompt
	apiKey := loginAPIKey
	if apiKey == "" {
		// Check environment variable
		apiKey = os.Getenv(llm.GetAPIKeyEnvVar(provider))
		if apiKey == "" {
			return fmt.Errorf("API key not provided. Set %s environment variable or use --api-key flag",
				llm.GetAPIKeyEnvVar(provider))
		}
	}

	// Validate API key by trying to list models
	fmt.Printf("Validating API key for %s...\n", llm.ProviderNames[provider])

	manager := auth.NewManager()
	ctx := context.Background()

	err := manager.ValidateAPIKey(ctx, provider, apiKey)
	if err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}

	// Show provider info
	fmt.Printf("\n%s\n", lipgloss.NewStyle().Bold(true).Render("✓ Authentication successful!"))
	fmt.Printf("Provider: %s\n", llm.ProviderNames[provider])

	// Show model info
	model := loginModel
	if model == "" {
		model = llm.DefaultModels[provider]
	}
	fmt.Printf("Default model: %s\n", model)

	// Show instructions for setting environment variable
	fmt.Printf("\n%s\n", lipgloss.NewStyle().Bold(true).Render("Environment variable:"))
	fmt.Printf("export %s=\"%s\"\n", llm.GetAPIKeyEnvVar(provider), apiKey)

	fmt.Printf("\n%s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("Authentication configured successfully!"))

	return nil
}

func listModels() error {
	if loginProvider == "" {
		return fmt.Errorf("provider is required for --list-models (use --provider)")
	}

	provider := llm.ProviderType(strings.ToLower(loginProvider))
	if provider != llm.ProviderAnthropic && provider != llm.ProviderOpenRouter && provider != llm.ProviderGemini {
		return fmt.Errorf("invalid provider: %s (valid: anthropic, openrouter, gemini)", loginProvider)
	}

	// Get API key
	apiKey := loginAPIKey
	if apiKey == "" {
		apiKey = os.Getenv(llm.GetAPIKeyEnvVar(provider))
		if apiKey == "" {
			return fmt.Errorf("API key not provided. Set %s environment variable or use --api-key flag",
				llm.GetAPIKeyEnvVar(provider))
		}
	}

	fmt.Printf("Fetching available models from %s...\n\n", llm.ProviderNames[provider])

	factory := llm.NewFactory()
	ctx := context.Background()

	models, err := factory.ListModelsForProvider(ctx, provider, apiKey)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	// Print models
	fmt.Printf("%-40s %s\n", lipgloss.NewStyle().Bold(true).Render("Model ID"), lipgloss.NewStyle().Bold(true).Render("Name"))
	fmt.Println(strings.Repeat("-", 80))

	for _, m := range models {
		name := m.Name
		if name == "" {
			name = m.ID
		}
		fmt.Printf("%-40s %s\n", m.ID, name)
	}

	fmt.Printf("\nTotal: %d models\n", len(models))

	return nil
}

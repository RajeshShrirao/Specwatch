package plugin

import (
	_ "embed" // Required for go:embed
)

// Auto-register built-in plugins on import
//
// To add a new LLM provider plugin:
// 1. Create a new file in internal/llm/ (e.g., custom.go)
// 2. Implement the Plugin interface
// 3. Add an init() function that calls plugin.Register(&YourPlugin{})
// 4. Import this package in your plugin
func init() {
	// Built-in plugins are auto-registered via their own init() functions
	// This file serves as the entry point for the plugin system
	// and ensures proper initialization order
}

// RegisterBuiltin registers all built-in LLM provider plugins
func RegisterBuiltin() error {
	// Import and register built-in providers
	// Note: This is handled by the init() functions in each provider package
	// This function can be used for explicit registration if needed
	return nil
}

package plugin

import (
	"fmt"
	"sync"
)

// Registry manages LLM provider plugins
type Registry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin to the registry
func (r *Registry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := plugin.ID()
	if id == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	if _, exists := r.plugins[id]; exists {
		return fmt.Errorf("plugin already registered: %s", id)
	}

	r.plugins[id] = plugin
	return nil
}

// Get retrieves a plugin by ID
func (r *Registry) Get(id string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[id]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", id)
	}
	return plugin, nil
}

// List returns all registered plugins
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// Unregister removes a plugin from the registry
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[id]; !exists {
		return fmt.Errorf("plugin not found: %s", id)
	}

	delete(r.plugins, id)
	return nil
}

// Len returns the number of registered plugins
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.plugins)
}

// Default default registry
var defaultRegistry = NewRegistry()

// Register is a convenience function that registers to the default registry
func Register(plugin Plugin) error {
	return defaultRegistry.Register(plugin)
}

// Get is a convenience function that gets from the default registry
func Get(id string) (Plugin, error) {
	return defaultRegistry.Get(id)
}

// ListAll is a convenience function that lists all from the default registry
func ListAll() []Plugin {
	return defaultRegistry.List()
}

// GetDefaultRegistry returns the default registry
func GetDefaultRegistry() *Registry {
	return defaultRegistry
}

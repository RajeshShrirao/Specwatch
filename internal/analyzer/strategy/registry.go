package strategy

import (
	"fmt"
	"sync"
)

// Registry manages strategy registration and lookup
type Registry struct {
	strategies    map[string]RuleStrategy
	categoryIndex map[string][]string // category -> strategy names
	mu            sync.RWMutex
}

// NewRegistry creates a new strategy registry
func NewRegistry() *Registry {
	return &Registry{
		strategies:    make(map[string]RuleStrategy),
		categoryIndex: make(map[string][]string),
	}
}

// Register adds a strategy to the registry
func (r *Registry) Register(strategy RuleStrategy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := strategy.Name()
	if name == "" {
		return fmt.Errorf("strategy name cannot be empty")
	}

	if _, exists := r.strategies[name]; exists {
		return fmt.Errorf("strategy already registered: %s", name)
	}

	r.strategies[name] = strategy

	// Index by category
	category := strategy.Category()
	if category != "" {
		r.categoryIndex[category] = append(r.categoryIndex[category], name)
	}

	return nil
}

// Get retrieves a strategy by name
func (r *Registry) Get(name string) (RuleStrategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	strategy, exists := r.strategies[name]
	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", name)
	}
	return strategy, nil
}

// GetByCategory returns all strategies for a given category
func (r *Registry) GetByCategory(category string) []RuleStrategy {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names, exists := r.categoryIndex[category]
	if !exists {
		return nil
	}

	strategies := make([]RuleStrategy, 0, len(names))
	for _, name := range names {
		if s, ok := r.strategies[name]; ok {
			strategies = append(strategies, s)
		}
	}
	return strategies
}

// List returns all registered strategies
func (r *Registry) List() []RuleStrategy {
	r.mu.RLock()
	defer r.mu.RUnlock()

	strategies := make([]RuleStrategy, 0, len(r.strategies))
	for _, s := range r.strategies {
		strategies = append(strategies, s)
	}
	return strategies
}

// Unregister removes a strategy from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	strategy, exists := r.strategies[name]
	if !exists {
		return fmt.Errorf("strategy not found: %s", name)
	}

	delete(r.strategies, name)

	// Remove from category index
	category := strategy.Category()
	if category != "" {
		names := r.categoryIndex[category]
		for i, n := range names {
			if n == name {
				r.categoryIndex[category] = append(names[:i], names[i+1:]...)
				break
			}
		}
	}

	return nil
}

// Len returns the number of registered strategies
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.strategies)
}

// DefaultRegistry is the global default registry
var DefaultRegistry = NewRegistry()

// Register is a convenience function that registers to the default registry
func Register(strategy RuleStrategy) error {
	return DefaultRegistry.Register(strategy)
}

// Get is a convenience function that gets from the default registry
func Get(name string) (RuleStrategy, error) {
	return DefaultRegistry.Get(name)
}

// GetByCategory is a convenience function that gets by category from the default registry
func GetByCategory(category string) []RuleStrategy {
	return DefaultRegistry.GetByCategory(category)
}

// ListAll is a convenience function that lists all from the default registry
func ListAll() []RuleStrategy {
	return DefaultRegistry.List()
}

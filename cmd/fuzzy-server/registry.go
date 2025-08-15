package main

import "github.com/bornholm/go-fuzzy"

type registryEntry struct {
	Rules     []*fuzzy.Rule
	Variables []*fuzzy.Variable
}

// Registry holds all the loaded fuzzy engine definitions.
type Registry struct {
	entries map[string]registryEntry
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		entries: make(map[string]registryEntry),
	}
}

// Get returns a fuzzy engine definition by name
func (r *Registry) Get(name string) ([]*fuzzy.Variable, []*fuzzy.Rule, bool) {
	entry, exists := r.entries[name]
	if !exists {
		return nil, nil, false
	}

	return entry.Variables, entry.Rules, exists
}

// Register adds a fuzzy engine definition to the registry
func (r *Registry) Register(name string, variables []*fuzzy.Variable, rules []*fuzzy.Rule) {
	r.entries[name] = registryEntry{
		Rules:     rules,
		Variables: variables,
	}
}

// Names returns all registered fuzzy engine definition names
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.entries))
	for name := range r.entries {
		names = append(names, name)
	}
	return names
}

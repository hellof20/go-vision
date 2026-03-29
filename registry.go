package vision

import (
	"context"
	"fmt"
	"log/slog"
)

// Registry holds multiple vision resolvers and routes requests
// to the correct resolver by name. It implements the Resolver interface.
type Registry struct {
	resolvers   map[string]Resolver
	defaultName string
}

// NewRegistry creates a new resolver registry with a default resolver name.
func NewRegistry(defaultName string) *Registry {
	return &Registry{
		resolvers:   make(map[string]Resolver),
		defaultName: defaultName,
	}
}

// Register adds a resolver to the registry.
func (r *Registry) Register(name string, resolver Resolver) {
	r.resolvers[name] = resolver
}

// Get returns a resolver by name.
func (r *Registry) Get(name string) (Resolver, bool) {
	res, ok := r.resolvers[name]
	return res, ok
}

// resolve selects the resolver to use.
func (r *Registry) resolve(ctx context.Context, name string) (Resolver, error) {
	resolver := r.resolvers[r.defaultName]
	if name != "" {
		if res, ok := r.resolvers[name]; ok {
			resolver = res
		} else {
			slog.WarnContext(ctx, "requested resolver not found, using default",
				"requested", name,
				"default", r.defaultName,
			)
		}
	}
	if resolver == nil {
		return nil, fmt.Errorf("no resolver available (requested=%s, default=%s)", name, r.defaultName)
	}
	return resolver, nil
}

// Resolve routes to the default resolver.
func (r *Registry) Resolve(ctx context.Context, locate string, screenshot []byte) (Coordinates, error) {
	resolver, err := r.resolve(ctx, "")
	if err != nil {
		return Coordinates{}, err
	}
	return resolver.Resolve(ctx, locate, screenshot)
}

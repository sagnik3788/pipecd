package toolregistry

import (
	"cmp"
	"context"
)

const (
	defaultOpenTofuVersion = "1.9.1"
)

type client interface {
	InstallTool(ctx context.Context, name, version, script string) (string, error)
}

// Registry provides functions to get path to the needed tools.
type Registry struct {
	client client
}

// NewRegistry creates a new Registry instance
func NewRegistry(client client) *Registry {
	return &Registry{
		client: client,
	}
}

// OpenTofu installs the OpenTofu tool with the given version and return the path to the installed binary.
// If the version is empty, the default version will be used.
func (r *Registry) OpenTofu(ctx context.Context, version string) (string, error) {
	return r.client.InstallTool(ctx, "tofu", cmp.Or(version, defaultOpenTofuVersion), installScript)
}

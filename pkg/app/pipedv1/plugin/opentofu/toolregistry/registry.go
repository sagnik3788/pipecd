// Package toolregistry installs and manages the needed tools
// such as tofu for executing tasks in pipeline.
package toolregistry

import (
	"context"
)

type client interface {
	InstallTool(ctx context.Context, name, version, script string) (path string, err error)
}

func NewRegistry(client client) *Registry {
	return &Registry{
		client: client,
	}
}

// Registry provides functions to get path to the needed tools.
type Registry struct {
	client client
}

func (r *Registry) OpenTofu(ctx context.Context, version string) (path string, err error) {
	return r.client.InstallTool(ctx, "tofu", version, installScript)
}

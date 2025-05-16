package toolregistry

import (
	"context"
)

// TestToolRegistry implements the client interface for tests.
type TestToolRegistry struct{}

func NewTestToolRegistry() *TestToolRegistry {
	return &TestToolRegistry{}
}

func (t *TestToolRegistry) InstallTool(ctx context.Context, name, version, script string) (string, error) {
	// Return a dummy path to a valid executable for testing.
	return "/bin/true", nil
}

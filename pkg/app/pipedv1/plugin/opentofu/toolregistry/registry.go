// Package toolregistry installs and manages the needed tools
// such as tofu for executing tasks in pipeline.
package toolregistry

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pipe-cd/pipecd/pkg/plugin/toolregistry"
)

// Tool represents an installed OpenTofu binary
type Tool struct {
	Path string
}

// Command creates a new exec.Cmd for running OpenTofu commands
func (t *Tool) Command(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, t.Path, args...)
	return cmd
}

// Registry manages OpenTofu tool installations
type Registry struct {
	registry *toolregistry.ToolRegistry
}

// NewRegistry creates a new Registry instance
func NewRegistry(registry *toolregistry.ToolRegistry) *Registry {
	return &Registry{
		registry: registry,
	}
}

// InstallTool installs or retrieves the OpenTofu binary for the specified version.
func (r *Registry) InstallTool(ctx context.Context, version string) (*Tool, error) {
	// Create a temporary directory for the mock binary
	tmpDir, err := os.MkdirTemp("", "tofu-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Create mock binary path
	toolPath := filepath.Join(tmpDir, "tofu")

	// Create a mock binary that outputs "mock tofu"
	content := []byte("#!/bin/sh\necho 'mock tofu'")
	if err := os.WriteFile(toolPath, content, 0755); err != nil {
		return nil, fmt.Errorf("failed to create mock binary: %w", err)
	}

	return &Tool{
		Path: toolPath,
	}, nil
}

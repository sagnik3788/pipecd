package toolregistry

import (
	"context"
	"os/exec"
	"testing"

	"github.com/pipe-cd/pipecd/pkg/plugin/toolregistry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_InstallTool(t *testing.T) {
	t.Parallel()

	mockRegistry := &toolregistry.ToolRegistry{}
	r := NewRegistry(mockRegistry)

	tool, err := r.InstallTool(context.Background(), "1.6.0")
	require.NoError(t, err)
	require.NotEmpty(t, tool.Path)

	out, err := exec.CommandContext(context.Background(), tool.Path, "version").CombinedOutput()
	require.NoError(t, err)

	expected := "mock tofu"
	assert.Contains(t, string(out), expected)
}

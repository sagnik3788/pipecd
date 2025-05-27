package toolregistry

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pipe-cd/pipecd/pkg/plugin/toolregistry/toolregistrytest"
)

func TestRegistry_OpenTofu(t *testing.T) {
	t.Parallel()

	c := toolregistrytest.NewTestToolRegistry(t)
	r := NewRegistry(c)

	// Test with default version
	path, err := r.OpenTofu(context.Background(), "")
	require.NoError(t, err)
	require.NotEmpty(t, path)

	// Verify the installed binary
	out, err := exec.CommandContext(context.Background(), path, "version").CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), defaultOpenTofuVersion)

	// Test with specific version
	specificVersion := "1.9.1"
	path, err = r.OpenTofu(context.Background(), specificVersion)
	require.NoError(t, err, string(out))
	require.NotEmpty(t, path)
	t.Logf("Specific version tofu binary path: %s", path)

	t.Cleanup(func() {
		dir := filepath.Dir(path)
		os.RemoveAll(dir)
	})

	// Verify the installed binary with specific version
	out, err = exec.CommandContext(context.Background(), path, "version").CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), specificVersion)

	// Test installation script template variables
	script := strings.TrimSpace(installScript)
	assert.Contains(t, script, "{{ .TmpDir }}")
	assert.Contains(t, script, "{{ .Version }}")
	assert.Contains(t, script, "{{ .Os }}")
	assert.Contains(t, script, "{{ .Arch }}")
	assert.Contains(t, script, "{{ .OutPath }}")

}

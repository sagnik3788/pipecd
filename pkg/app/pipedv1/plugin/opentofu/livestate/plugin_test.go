package livestate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/toolregistry"
	"github.com/pipe-cd/pipecd/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRegistry implements the toolregistry.Registry interface for testing
type mockRegistry struct {
	workingDir string
}

func (m *mockRegistry) InstallTool(ctx context.Context, version string) (*toolregistry.Tool, error) {
	toolPath := filepath.Join(m.workingDir, "tofu")

	// Create a mock script that returns different outputs based on command
	script := `#!/bin/sh
case "$1" in
  "show")
    # Check if we're in the invalid config directory
    if [ -f "$(pwd)/testdata/invalid/main.tf" ]; then
      echo "Error: Invalid resource type 'invalid_resource'" >&2
      exit 1
    fi
    echo '{
      "values": {
        "root_module": {
          "resources": [
            {
              "address": "null_resource.test",
              "values": {
                "id": "123",
                "triggers": null
              }
            }
          ],
          "outputs": [
            {
              "name": "test_output",
              "value": "test_value"
            }
          ]
        }
      }
    }'
    ;;
  "plan")
    # Check if we're in the invalid config directory
    if [ -f "$(pwd)/testdata/invalid/main.tf" ]; then
      echo "Error: Invalid resource type 'invalid_resource'" >&2
      exit 1
    fi
    echo '{
      "values": {
        "root_module": {
          "resources": [
            {
              "address": "null_resource.test",
              "values": {
                "id": "456",
                "triggers": null
              }
            }
          ],
          "outputs": [
            {
              "name": "test_output",
              "value": "new_value"
            }
          ]
        }
      }
    }'
    ;;
  "init")
    # Check if we're in the invalid config directory
    if [ -f "$(pwd)/testdata/invalid/main.tf" ]; then
      echo "Error: Invalid resource type 'invalid_resource'" >&2
      exit 1
    fi
    ;;
  *)
    echo "mock tofu"
    ;;
esac`

	if err := os.WriteFile(toolPath, []byte(script), 0755); err != nil {
		return nil, err
	}
	return &toolregistry.Tool{Path: toolPath}, nil
}

func setupTestData(t *testing.T, workingDir string) {
	// Create test data directory
	testDataDir := filepath.Join(workingDir, "test-app", "testdata", "simple")
	require.NoError(t, os.MkdirAll(testDataDir, 0755))

	// Create a simple main.tf file
	mainTf := `resource "null_resource" "test" {
  provisioner "local-exec" {
    command = "echo 'test'"
  }
}

output "test_output" {
  value = "test_value"
}`
	require.NoError(t, os.WriteFile(filepath.Join(testDataDir, "main.tf"), []byte(mainTf), 0644))
}

func TestPlugin_GetLiveState(t *testing.T) {
	// Setup test environment
	workingDir, err := os.MkdirTemp("", "opentofu-livestate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	// Setup test data
	setupTestData(t, workingDir)

	registry := &mockRegistry{workingDir: workingDir}
	plugin := NewPlugin(workingDir, "1.6.0", registry)

	// Create test application
	app := &model.Application{
		Id:   "test-app",
		Kind: model.ApplicationKind_TERRAFORM,
		GitPath: &model.ApplicationGitPath{
			Path: "testdata/simple",
		},
	}

	// Test GetLiveState
	result, err := plugin.GetLiveState(context.Background(), app)
	require.NoError(t, err)
	assert.NotEmpty(t, result.State)

	// Verify state format
	var state State
	err = json.Unmarshal([]byte(result.State), &state)
	require.NoError(t, err)
	assert.Len(t, state.Values.RootModule.Resources, 1)
	assert.Equal(t, "null_resource.test", state.Values.RootModule.Resources[0].Address)
	assert.Equal(t, "123", state.Values.RootModule.Resources[0].Values["id"])
}

func TestPlugin_GetDesiredState(t *testing.T) {
	// Setup test environment
	workingDir, err := os.MkdirTemp("", "opentofu-livestate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	// Setup test data
	setupTestData(t, workingDir)

	registry := &mockRegistry{workingDir: workingDir}
	plugin := NewPlugin(workingDir, "1.6.0", registry)

	// Create test application
	app := &model.Application{
		Id:   "test-app",
		Kind: model.ApplicationKind_TERRAFORM,
		GitPath: &model.ApplicationGitPath{
			Path: "testdata/simple",
		},
	}

	// Test GetDesiredState
	result, err := plugin.GetDesiredState(context.Background(), app)
	require.NoError(t, err)
	assert.NotEmpty(t, result.State)

	// Verify state format
	var state State
	err = json.Unmarshal([]byte(result.State), &state)
	require.NoError(t, err)
	assert.Len(t, state.Values.RootModule.Resources, 1)
	assert.Equal(t, "null_resource.test", state.Values.RootModule.Resources[0].Address)
	assert.Equal(t, "456", state.Values.RootModule.Resources[0].Values["id"])
}

func TestPlugin_GetDiff(t *testing.T) {
	// Setup test environment
	workingDir, err := os.MkdirTemp("", "opentofu-livestate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	// Setup test data
	setupTestData(t, workingDir)

	registry := &mockRegistry{workingDir: workingDir}
	plugin := NewPlugin(workingDir, "1.6.0", registry)

	// Create test application
	app := &model.Application{
		Id:   "test-app",
		Kind: model.ApplicationKind_TERRAFORM,
		GitPath: &model.ApplicationGitPath{
			Path: "testdata/simple",
		},
	}

	// Test GetDiff
	result, err := plugin.GetDiff(context.Background(), app)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Diff)

	// Verify diff format
	var diff map[string]interface{}
	err = json.Unmarshal([]byte(result.Diff), &diff)
	require.NoError(t, err)
	assert.Contains(t, diff, "changes")

	changes := diff["changes"].([]interface{})
	assert.Len(t, changes, 1)

	change := changes[0].(map[string]interface{})
	assert.Equal(t, "update", change["change"])
	assert.Equal(t, "null_resource.test", change["address"])
}

func TestPlugin_GetLiveState_InvalidConfig(t *testing.T) {
	// Setup test environment
	workingDir, err := os.MkdirTemp("", "opentofu-livestate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	// Create test data directory with invalid configuration
	testDataDir := filepath.Join(workingDir, "test-app", "testdata", "invalid")
	require.NoError(t, os.MkdirAll(testDataDir, 0755))

	// Create an invalid main.tf file
	mainTf := `resource "invalid_resource" "test" {
  invalid_attribute = "value"
}`
	require.NoError(t, os.WriteFile(filepath.Join(testDataDir, "main.tf"), []byte(mainTf), 0644))

	registry := &mockRegistry{workingDir: workingDir}
	plugin := NewPlugin(workingDir, "1.6.0", registry)

	// Create test application with invalid config
	app := &model.Application{
		Id:   "test-app",
		Kind: model.ApplicationKind_TERRAFORM,
		GitPath: &model.ApplicationGitPath{
			Path: "testdata/invalid",
		},
	}

	// Test GetLiveState with invalid config
	_, err = plugin.GetLiveState(context.Background(), app)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize OpenTofu")
}

func TestPlugin_GetDesiredState_InvalidConfig(t *testing.T) {
	// Setup test environment
	workingDir, err := os.MkdirTemp("", "opentofu-livestate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	// Create test data directory with invalid configuration
	testDataDir := filepath.Join(workingDir, "test-app", "testdata", "invalid")
	require.NoError(t, os.MkdirAll(testDataDir, 0755))

	// Create an invalid main.tf file
	mainTf := `resource "invalid_resource" "test" {
  invalid_attribute = "value"
}`
	require.NoError(t, os.WriteFile(filepath.Join(testDataDir, "main.tf"), []byte(mainTf), 0644))

	registry := &mockRegistry{workingDir: workingDir}
	plugin := NewPlugin(workingDir, "1.6.0", registry)

	// Create test application with invalid config
	app := &model.Application{
		Id:   "test-app",
		Kind: model.ApplicationKind_TERRAFORM,
		GitPath: &model.ApplicationGitPath{
			Path: "testdata/invalid",
		},
	}

	// Test GetDesiredState with invalid config
	_, err = plugin.GetDesiredState(context.Background(), app)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize OpenTofu")
}

func TestPlugin_GetDiff_InvalidConfig(t *testing.T) {
	// Setup test environment
	workingDir, err := os.MkdirTemp("", "opentofu-livestate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	// Create test data directory with invalid configuration
	testDataDir := filepath.Join(workingDir, "test-app", "testdata", "invalid")
	require.NoError(t, os.MkdirAll(testDataDir, 0755))

	// Create an invalid main.tf file
	mainTf := `resource "invalid_resource" "test" {
  invalid_attribute = "value"
}`
	require.NoError(t, os.WriteFile(filepath.Join(testDataDir, "main.tf"), []byte(mainTf), 0644))

	registry := &mockRegistry{workingDir: workingDir}
	plugin := NewPlugin(workingDir, "1.6.0", registry)

	// Create test application with invalid config
	app := &model.Application{
		Id:   "test-app",
		Kind: model.ApplicationKind_TERRAFORM,
		GitPath: &model.ApplicationGitPath{
			Path: "testdata/invalid",
		},
	}

	// Test GetDiff with invalid config
	_, err = plugin.GetDiff(context.Background(), app)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize OpenTofu")
}

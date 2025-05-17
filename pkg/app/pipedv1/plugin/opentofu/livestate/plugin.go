package livestate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/toolregistry"
	"github.com/pipe-cd/pipecd/pkg/model"
)

// State represents the OpenTofu state structure
type State struct {
	Values struct {
		RootModule struct {
			Resources []Resource `json:"resources"`
			Outputs   []Output   `json:"outputs"`
		} `json:"root_module"`
	} `json:"values"`
}

// Resource represents an OpenTofu resource
type Resource struct {
	Address string                 `json:"address"`
	Values  map[string]interface{} `json:"values"`
}

// Output represents an OpenTofu output
type Output struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// Result types for plugin responses
type GetLiveStateResult struct {
	State string
}

type GetDesiredStateResult struct {
	State string
}

type GetDiffResult struct {
	Diff string
}

// Registry defines the interface for tool management
type Registry interface {
	InstallTool(ctx context.Context, version string) (*toolregistry.Tool, error)
}

type Plugin struct {
	workingDir string
	version    string
	registry   Registry
}

func NewPlugin(workingDir, version string, registry Registry) *Plugin {
	return &Plugin{
		workingDir: workingDir,
		version:    version,
		registry:   registry,
	}
}

func (p *Plugin) validateConfig(app *model.Application) error {
	if app.GitPath == nil || app.GitPath.Path == "" {
		return fmt.Errorf("git path is required")
	}

	configPath := filepath.Join(p.workingDir, app.Id, app.GitPath.Path)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration directory does not exist: %s", configPath)
	}

	return nil
}

func (p *Plugin) validateState(state []byte) error {
	var s State
	if err := json.Unmarshal(state, &s); err != nil {
		return fmt.Errorf("invalid state format: %w", err)
	}
	return nil
}

func (p *Plugin) GetLiveState(ctx context.Context, app *model.Application) (*GetLiveStateResult, error) {
	if err := p.validateConfig(app); err != nil {
		return nil, err
	}

	// Get OpenTofu binary
	tool, err := p.registry.InstallTool(ctx, p.version)
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenTofu binary: %w", err)
	}

	// Create working directory
	workDir := filepath.Join(p.workingDir, app.Id)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create working directory: %w", err)
	}

	// Initialize OpenTofu
	initCmd := tool.Command(ctx, "init", "-input=false")
	initCmd.Dir = workDir
	if err := initCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to initialize OpenTofu: %w", err)
	}

	// Get state
	stateCmd := tool.Command(ctx, "show", "-json")
	stateCmd.Dir = workDir
	output, err := stateCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenTofu state: %w", err)
	}

	// Validate state format
	if err := p.validateState(output); err != nil {
		return nil, err
	}

	return &GetLiveStateResult{
		State: string(output),
	}, nil
}

func (p *Plugin) GetDesiredState(ctx context.Context, app *model.Application) (*GetDesiredStateResult, error) {
	if err := p.validateConfig(app); err != nil {
		return nil, err
	}

	// Get OpenTofu binary
	tool, err := p.registry.InstallTool(ctx, p.version)
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenTofu binary: %w", err)
	}

	// Create working directory
	workDir := filepath.Join(p.workingDir, app.Id)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create working directory: %w", err)
	}

	// Initialize OpenTofu
	initCmd := tool.Command(ctx, "init", "-input=false")
	initCmd.Dir = workDir
	if err := initCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to initialize OpenTofu: %w", err)
	}

	// Get plan
	planCmd := tool.Command(ctx, "plan", "-input=false", "-json")
	planCmd.Dir = workDir
	output, err := planCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenTofu plan: %w", err)
	}

	// Validate state format
	if err := p.validateState(output); err != nil {
		return nil, err
	}

	return &GetDesiredStateResult{
		State: string(output),
	}, nil
}

func (p *Plugin) GetDiff(ctx context.Context, app *model.Application) (*GetDiffResult, error) {
	if err := p.validateConfig(app); err != nil {
		return nil, err
	}

	tool, err := p.registry.InstallTool(ctx, p.version)
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenTofu binary: %w", err)
	}

	workDir := filepath.Join(p.workingDir, app.Id)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create working directory: %w", err)
	}

	// Initialize OpenTofu
	initCmd := tool.Command(ctx, "init", "-input=false")
	initCmd.Dir = workDir
	if err := initCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to initialize OpenTofu: %w", err)
	}

	// 1. Get current state (implemented state)
	stateCmd := tool.Command(ctx, "show", "-json")
	stateCmd.Dir = workDir
	currentState, err := stateCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current state: %w", err)
	}

	// Validate current state
	if err := p.validateState(currentState); err != nil {
		return nil, fmt.Errorf("invalid current state: %w", err)
	}

	// 2. Get desired state (from plan)
	planCmd := tool.Command(ctx, "plan", "-input=false", "-json")
	planCmd.Dir = workDir
	desiredState, err := planCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get desired state: %w", err)
	}

	// Validate desired state
	if err := p.validateState(desiredState); err != nil {
		return nil, fmt.Errorf("invalid desired state: %w", err)
	}

	// 3. Compare states to get actual differences
	var current, desired State
	if err := json.Unmarshal(currentState, &current); err != nil {
		return nil, fmt.Errorf("failed to parse current state: %w", err)
	}
	if err := json.Unmarshal(desiredState, &desired); err != nil {
		return nil, fmt.Errorf("failed to parse desired state: %w", err)
	}

	// 4. Generate diff by comparing resources
	diff := make(map[string]interface{})

	// Compare resources and track changes
	changes := []map[string]interface{}{}
	for _, desiredRes := range desired.Values.RootModule.Resources {
		// Find matching resource in current state
		var currentRes *Resource
		for _, currRes := range current.Values.RootModule.Resources {
			if currRes.Address == desiredRes.Address {
				currentRes = &currRes
				break
			}
		}

		// If resource doesn't exist in current state, it's new
		if currentRes == nil {
			changes = append(changes, map[string]interface{}{
				"address":  desiredRes.Address,
				"change":   "create",
				"resource": desiredRes,
			})
			continue
		}

		// Compare attributes to find changes
		if !reflect.DeepEqual(currentRes.Values, desiredRes.Values) {
			changes = append(changes, map[string]interface{}{
				"address": desiredRes.Address,
				"change":  "update",
				"current": currentRes.Values,
				"desired": desiredRes.Values,
			})
		}
	}

	// Check for resources to be deleted
	for _, currentRes := range current.Values.RootModule.Resources {
		// Check if resource exists in desired state
		exists := false
		for _, desiredRes := range desired.Values.RootModule.Resources {
			if desiredRes.Address == currentRes.Address {
				exists = true
				break
			}
		}

		// If resource doesn't exist in desired state, it should be deleted
		if !exists {
			changes = append(changes, map[string]interface{}{
				"address":  currentRes.Address,
				"change":   "delete",
				"resource": currentRes,
			})
		}
	}

	diff["changes"] = changes

	return &GetDiffResult{
		Diff: string(mustMarshal(diff)),
	}, nil
}

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

package livestate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/config"
	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/toolregistry"
	"github.com/pipe-cd/pipecd/pkg/plugin/sdk"
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

// Plan represents the OpenTofu plan structure
type Plan struct {
	PlannedValues struct {
		RootModule struct {
			Resources []Resource `json:"resources"`
			Outputs   []Output   `json:"outputs"`
		} `json:"root_module"`
	} `json:"planned_values"`
	ResourceChanges []ResourceChange `json:"resource_changes"`
}

// ResourceChange represents a change in a resource
type ResourceChange struct {
	Address      string `json:"address"`
	Change       Change `json:"change"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	ProviderName string `json:"provider_name"`
}

// Change represents the changes to be made to a resource
type Change struct {
	Actions []string               `json:"actions"`
	Before  map[string]interface{} `json:"before"`
	After   map[string]interface{} `json:"after"`
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

type Plugin struct{}

// GetLivestate implements sdk.LivestatePlugin.
func (p Plugin) GetLivestate(ctx context.Context, _ *sdk.ConfigNone, deployTargets []*sdk.DeployTarget[config.OpenTofuDeployTargetConfig], input *sdk.GetLivestateInput[config.OpenTofuApplicationSpec]) (*sdk.GetLivestateResponse, error) {
	if len(deployTargets) != 1 {
		return nil, fmt.Errorf("only 1 deploy target is allowed but got %d", len(deployTargets))
	}

	deployTarget := deployTargets[0]
	deployTargetConfig := deployTarget.Config

	// Create tool registry
	toolRegistry := toolregistry.NewRegistry(input.Client.ToolRegistry())

	// Get OpenTofu binary
	tofuPath, err := toolRegistry.OpenTofu(ctx, deployTargetConfig.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenTofu binary: %w", err)
	}

	// Create working directory
	workDir := filepath.Join(os.TempDir(), "opentofu-plugin", input.Request.ApplicationID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create working directory: %w", err)
	}

	// Copy OpenTofu configuration files from source to working directory
	sourceDir := input.Request.DeploymentSource.ApplicationConfig.Spec.Input.Config
	if err := copyDir(sourceDir, workDir); err != nil {
		return nil, fmt.Errorf("failed to copy OpenTofu configuration files: %w", err)
	}

	// Initialize OpenTofu
	initCmd := exec.CommandContext(ctx, tofuPath, "init", "-input=false")
	initCmd.Dir = workDir
	if err := initCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to initialize OpenTofu: %w", err)
	}

	// Get live state
	stateCmd := exec.CommandContext(ctx, tofuPath, "show", "-json")
	stateCmd.Dir = workDir
	output, err := stateCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenTofu state: %w", err)
	}

	// Parse live state
	var liveState State
	if err := json.Unmarshal(output, &liveState); err != nil {
		return nil, fmt.Errorf("failed to parse live state: %w", err)
	}

	// Get desired state (from plan)
	planCmd := exec.CommandContext(ctx, tofuPath, "plan", "-json")
	planCmd.Dir = workDir
	planOutput, err := planCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenTofu plan: %w", err)
	}

	// Parse plan
	var plan Plan
	if err := json.Unmarshal(planOutput, &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	// Compute diff between live and desired states
	diff, reasons := computeDiff(liveState, plan)

	// Convert live resources to ResourceState
	resourceStates := make([]sdk.ResourceState, 0, len(liveState.Values.RootModule.Resources))
	for _, res := range liveState.Values.RootModule.Resources {
		// Find if this resource has changes
		var healthStatus sdk.ResourceHealthStatus
		var healthDesc string
		for _, change := range plan.ResourceChanges {
			if change.Address == res.Address {
				if len(change.Change.Actions) > 0 {
					healthStatus = sdk.ResourceHealthStatus(1) // Degraded
					healthDesc = fmt.Sprintf("Resource will be %s", strings.Join(change.Change.Actions, ", "))
				}
				break
			}
		}

		resourceStates = append(resourceStates, sdk.ResourceState{
			ID:                res.Address,
			Name:              res.Address,
			ResourceType:      res.Address,
			ResourceMetadata:  map[string]string{"values": fmt.Sprintf("%v", res.Values)},
			HealthStatus:      healthStatus,
			HealthDescription: healthDesc,
		})
	}

	// Determine sync status based on diff
	syncStatus := sdk.ApplicationSyncStateSynced
	shortReason := ""
	reason := ""
	if len(diff) > 0 {
		syncStatus = sdk.ApplicationSyncStateOutOfSync
		shortReason = fmt.Sprintf("%d resources need to be updated", len(diff))
		reason = strings.Join(reasons, "\n")
	}

	return &sdk.GetLivestateResponse{
		LiveState: sdk.ApplicationLiveState{
			Resources:    resourceStates,
			HealthStatus: sdk.ApplicationHealthStateUnknown,
		},
		SyncState: sdk.ApplicationSyncState{
			Status:      syncStatus,
			ShortReason: shortReason,
			Reason:      reason,
		},
	}, nil
}

// computeDiff returns a list of differences between live and desired states.
func computeDiff(live State, plan Plan) ([]string, []string) {
	var diffs []string
	var reasons []string

	// Create a map of live resources for quick lookup
	liveResources := make(map[string]Resource)
	for _, res := range live.Values.RootModule.Resources {
		liveResources[res.Address] = res
	}

	// Check for resource changes
	for _, change := range plan.ResourceChanges {
		if len(change.Change.Actions) > 0 {
			diffs = append(diffs, change.Address)
			reason := fmt.Sprintf("%s will be %s", change.Address, strings.Join(change.Change.Actions, ", "))
			reasons = append(reasons, reason)
		}
	}

	return diffs, reasons
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Skip copying if destination file exists
			if _, err := os.Stat(dstPath); err == nil {
				continue
			}
			// Copy file
			content, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, content, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

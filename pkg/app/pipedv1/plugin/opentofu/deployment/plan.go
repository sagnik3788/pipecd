package deployment

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/config"
	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/toolregistry"
	"github.com/pipe-cd/pipecd/pkg/plugin/sdk"
	"go.uber.org/zap"
)

const (
	stageOpenTofuPlan  = "OPEN_TOFU_PLAN"
	stageOpenTofuApply = "OPEN_TOFU_APPLY"
)

func determineVersions(input *sdk.DetermineVersionsInput[config.OpenTofuApplicationSpec]) (*sdk.DetermineVersionsResponse, error) {
	// Extract version from the OpenTofu configuration
	// For now, we'll use a simple versioning scheme
	version := "v1"
	if input.Request.DeploymentSource.CommitHash != "" {
		version = fmt.Sprintf("v1-%s", input.Request.DeploymentSource.CommitHash)
	}

	return &sdk.DetermineVersionsResponse{
		Versions: []sdk.ArtifactVersion{
			{
				Kind:    sdk.ArtifactKindTerraformModule,
				Version: version,
				Name:    input.Request.DeploymentSource.ApplicationDirectory,
				URL:     input.Request.DeploymentSource.ApplicationDirectory,
			},
		},
	}, nil
}

func determineStrategy() (*sdk.DetermineStrategyResponse, error) {
	// For OpenTofu, we'll use QUICK_SYNC strategy by default
	// This means we'll apply changes directly without a pipeline
	return &sdk.DetermineStrategyResponse{
		Strategy: sdk.SyncStrategyQuickSync,
		Summary:  "Using QUICK_SYNC strategy for OpenTofu deployment",
	}, nil
}

func buildQuickSyncStages() (*sdk.BuildQuickSyncStagesResponse, error) {
	stages := []sdk.QuickSyncStage{
		{
			Name:               stageOpenTofuPlan,
			Description:        "Plan OpenTofu changes",
			Rollback:           false,
			Metadata:           map[string]string{"command": "plan"},
			AvailableOperation: sdk.ManualOperationNone,
		},
		{
			Name:               stageOpenTofuApply,
			Description:        "Apply OpenTofu changes",
			Rollback:           false,
			Metadata:           map[string]string{"command": "apply"},
			AvailableOperation: sdk.ManualOperationNone,
		},
	}

	return &sdk.BuildQuickSyncStagesResponse{
		Stages: stages,
	}, nil
}

func buildPipelineSyncStages(input *sdk.BuildPipelineSyncStagesInput) (*sdk.BuildPipelineSyncStagesResponse, error) {
	stages := make([]sdk.PipelineStage, 0, len(input.Request.Stages))
	for _, s := range input.Request.Stages {
		stages = append(stages, sdk.PipelineStage{
			Index:              s.Index,
			Name:               s.Name,
			Rollback:           false,
			Metadata:           map[string]string{},
			AvailableOperation: sdk.ManualOperationNone,
		})
	}

	return &sdk.BuildPipelineSyncStagesResponse{
		Stages: stages,
	}, nil
}

func fetchDefinedStages() []string {
	return []string{
		stageOpenTofuPlan,
		stageOpenTofuApply,
	}
}

func (p *Plugin) executeStage(input *sdk.ExecuteStageInput[config.OpenTofuApplicationSpec], appCfg *config.OpenTofuApplicationSpec) (sdk.StageStatus, error) {
	logger := input.Logger

	// Get the OpenTofu binary path from the tool registry
	toolRegistry := toolregistry.NewRegistry(input.Client.ToolRegistry())
	tofuPath, err := toolRegistry.OpenTofu(context.Background(), appCfg.Input.Version)
	if err != nil {
		logger.Error("failed to get OpenTofu binary", zap.Error(err))
		return sdk.StageStatusFailure, fmt.Errorf("failed to get OpenTofu binary: %w", err)
	}

	// Check if configuration file exists
	configPath := filepath.Join(appCfg.Input.WorkingDir, appCfg.Input.Config)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return sdk.StageStatusFailure, fmt.Errorf("configuration file %s does not exist", configPath)
	}

	// Always run init to ensure proper dependency initialization
	output, err := p.runTofuCommand(tofuPath, "init", appCfg)
	if err != nil {
		logger.Error("failed to execute OpenTofu init",
			zap.Error(err),
			zap.String("output", output),
		)
		return sdk.StageStatusFailure, fmt.Errorf("failed to execute OpenTofu init: %w", err)
	}

	// Determine the command based on the stage name
	var command string
	var args []string
	switch input.Request.StageName {
	case stageOpenTofuPlan:
		// Use a unique plan file name based on deployment ID
		planFile := fmt.Sprintf("tfplan-%s", input.Request.Deployment.ApplicationID)
		command = "plan"
		args = []string{"-out=" + planFile}
	case stageOpenTofuApply:
		// Use the same plan file name as generated in plan stage
		planFile := fmt.Sprintf("tfplan-%s", input.Request.Deployment.ApplicationID)
		// Check if plan file exists
		if _, err := os.Stat(filepath.Join(appCfg.Input.WorkingDir, planFile)); os.IsNotExist(err) {
			return sdk.StageStatusFailure, fmt.Errorf("plan file %s does not exist, run plan first", planFile)
		}
		command = "apply"
		args = []string{planFile}
	default:
		logger.Error("unknown stage", zap.String("stage", input.Request.StageName))
		return sdk.StageStatusFailure, fmt.Errorf("unknown stage: %s", input.Request.StageName)
	}

	// Execute the OpenTofu command
	output, err = p.runTofuCommand(tofuPath, command, appCfg, args...)
	if err != nil {
		logger.Error("failed to execute OpenTofu command",
			zap.String("command", command),
			zap.Error(err),
			zap.String("output", output),
		)
		return sdk.StageStatusFailure, fmt.Errorf("failed to execute stage: %w", err)
	}

	// Log the output for debugging
	logger.Info("OpenTofu command output",
		zap.String("command", command),
		zap.String("output", output),
	)

	return sdk.StageStatusSuccess, nil
}

// runTofuCommand executes the tofu command for the given stage.
func (p *Plugin) runTofuCommand(tofuPath string, command string, spec *config.OpenTofuApplicationSpec, args ...string) (string, error) {
	// Create command with context
	cmd := exec.CommandContext(context.Background(), tofuPath, append([]string{command}, args...)...)

	// Set environment variables if specified
	if len(spec.Input.Env) > 0 {
		cmd.Env = append(cmd.Env, spec.Input.Env...)
	}

	// Set working directory if specified
	if spec.Input.WorkingDir != "" {
		cmd.Dir = spec.Input.WorkingDir
	}

	// Execute the command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to execute tofu command: %w", err)
	}

	return string(output), nil
}

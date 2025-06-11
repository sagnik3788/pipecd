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
	planFileDir        = "/tmp/tofu-plans" // Shared directory for plan files
)

// getPlanFileMetadataPath returns the path to the metadata file for a deployment
func getPlanFileMetadataPath(deploymentID string) string {
	return filepath.Join(planFileDir, fmt.Sprintf("metadata_%s.json", deploymentID))
}

// savePlanFileMetadata saves the plan file path to a metadata file
func savePlanFileMetadata(deploymentID, planFilePath string) error {
	metadataPath := getPlanFileMetadataPath(deploymentID)
	data := []byte(planFilePath)
	return os.WriteFile(metadataPath, data, 0644)
}

// loadPlanFileMetadata loads the plan file path from a metadata file
func loadPlanFileMetadata(deploymentID string) (string, error) {
	metadataPath := getPlanFileMetadataPath(deploymentID)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

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

	// Create the shared plan file directory if it doesn't exist
	if err := os.MkdirAll(planFileDir, 0755); err != nil {
		logger.Error("[PLUGIN] Failed to create plan file directory",
			zap.String("dir", planFileDir),
			zap.Error(err),
		)
		return sdk.StageStatusFailure, fmt.Errorf("failed to create plan file directory: %w", err)
	}

	// Log the stage configuration at the start
	logger.Info("[PLUGIN] Starting stage execution",
		zap.String("stageName", input.Request.StageName),
		zap.String("stageConfig", string(input.Request.StageConfig)),
		zap.Any("appConfig", appCfg),
	)

	// Validate app configuration
	if appCfg == nil {
		logger.Error("[PLUGIN] application configuration is nil")
		return sdk.StageStatusFailure, fmt.Errorf("application configuration is nil")
	}

	// Get the working directory
	workingDir := appCfg.Input.WorkingDir
	if workingDir == "" {
		workingDir = "."
	}

	// Get the absolute path by joining with the application directory
	absWorkingDir := filepath.Join(input.Request.TargetDeploymentSource.ApplicationDirectory, workingDir)
	logger.Info("[PLUGIN] Working directory information",
		zap.String("workingDir", workingDir),
		zap.String("absWorkingDir", absWorkingDir),
		zap.String("configPath", appCfg.Input.Config),
		zap.String("applicationDir", input.Request.TargetDeploymentSource.ApplicationDirectory),
	)

	// List files in working directory
	files, err := os.ReadDir(absWorkingDir)
	if err != nil {
		logger.Error("[PLUGIN] Failed to read working directory",
			zap.String("workingDir", absWorkingDir),
			zap.Error(err),
		)
		return sdk.StageStatusFailure, fmt.Errorf("failed to read working directory: %w", err)
	}

	// Log all files in the directory
	for _, f := range files {
		logger.Info("[PLUGIN] File in working directory",
			zap.String("name", f.Name()),
			zap.Bool("isDir", f.IsDir()),
		)
	}

	// Check if configuration file exists
	configPath := filepath.Join(absWorkingDir, appCfg.Input.Config)
	logger.Info("[PLUGIN] Checking for configuration file",
		zap.String("configPath", configPath),
	)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Error("[PLUGIN] Configuration file does not exist",
			zap.String("configPath", configPath),
			zap.Error(err),
		)
		return sdk.StageStatusFailure, fmt.Errorf("configuration file %s does not exist", configPath)
	} else if err != nil {
		logger.Error("[PLUGIN] Error checking configuration file",
			zap.String("configPath", configPath),
			zap.Error(err),
		)
		return sdk.StageStatusFailure, fmt.Errorf("error checking configuration file %s: %w", configPath, err)
	} else {
		logger.Info("[PLUGIN] Configuration file exists",
			zap.String("configPath", configPath),
		)
		// Optionally, log the first few lines of the config file for confirmation
		if content, err := os.ReadFile(configPath); err == nil {
			maxLines := 10
			lines := 0
			for _, line := range splitLines(string(content)) {
				if lines >= maxLines {
					logger.Info("[PLUGIN] ... (truncated config file output)")
					break
				}
				logger.Info("[PLUGIN] main.tf line", zap.String("line", line))
				lines++
			}
		}
	}

	// Get the OpenTofu binary path from the tool registry
	toolRegistry := toolregistry.NewRegistry(input.Client.ToolRegistry())
	tofuPath, err := toolRegistry.OpenTofu(context.Background(), appCfg.Input.Version)
	if err != nil {
		logger.Error("Failed to get OpenTofu binary",
			zap.Error(err),
			zap.String("version", appCfg.Input.Version),
		)
		return sdk.StageStatusFailure, fmt.Errorf("failed to get OpenTofu binary: %w", err)
	}

	// Always run init to ensure proper dependency initialization
	output, err := p.runTofuCommand(tofuPath, "init", absWorkingDir, appCfg)
	if err != nil {
		logger.Error("failed to execute OpenTofu init",
			zap.Error(err),
			zap.String("output", output),
		)
		return sdk.StageStatusFailure, fmt.Errorf("failed to execute OpenTofu init: %w", err)
	}

	// Define planFile with absolute path
	planFileName := fmt.Sprintf("tfplan-%s", input.Request.Deployment.ApplicationID)
	planFile := filepath.Join(planFileDir, planFileName)
	logger.Info("[PLUGIN] Generated plan file path",
		zap.String("planFile", planFile),
		zap.String("stage", input.Request.StageName),
	)

	// Determine the command based on the stage name
	var command string
	var args []string
	switch input.Request.StageName {
	case stageOpenTofuPlan:
		command = "plan"
		args = []string{"-out=" + planFile}
		logger.Info("[PLUGIN] Executing plan stage",
			zap.String("planFile", planFile),
			zap.String("workingDir", absWorkingDir),
		)

	case stageOpenTofuApply:
		// Load the plan file path from metadata file
		planFilePath, err := loadPlanFileMetadata(input.Request.Deployment.ID)
		if err != nil {
			logger.Error("[PLUGIN] Failed to load plan file metadata",
				zap.Error(err),
				zap.String("stage", input.Request.StageName),
				zap.String("deploymentID", input.Request.Deployment.ID),
			)
			return sdk.StageStatusFailure, fmt.Errorf("failed to load plan file metadata: %w", err)
		}

		logger.Info("[PLUGIN] Retrieved plan file path from metadata file",
			zap.String("planFilePath", planFilePath),
			zap.String("stage", input.Request.StageName),
			zap.String("deploymentID", input.Request.Deployment.ID),
		)

		// Validate plan file path
		if planFilePath == "" {
			logger.Error("[PLUGIN] Empty plan file path retrieved from metadata file",
				zap.String("stage", input.Request.StageName),
				zap.String("deploymentID", input.Request.Deployment.ID),
			)
			return sdk.StageStatusFailure, fmt.Errorf("empty plan file path retrieved from metadata file")
		}

		// Check if plan file exists
		if _, err := os.Stat(planFilePath); os.IsNotExist(err) {
			logger.Error("[PLUGIN] Plan file does not exist",
				zap.String("planFilePath", planFilePath),
				zap.Error(err),
			)
			return sdk.StageStatusFailure, fmt.Errorf("plan file %s does not exist, run plan first", planFilePath)
		}

		command = "apply"
		args = []string{planFilePath}
	default:
		logger.Error("unknown stage", zap.String("stage", input.Request.StageName))
		return sdk.StageStatusFailure, fmt.Errorf("unknown stage: %s", input.Request.StageName)
	}

	// Execute the OpenTofu command
	output, err = p.runTofuCommand(tofuPath, command, absWorkingDir, appCfg, args...)
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

	input.Logger.Info(fmt.Sprintf(
		"Received ExecuteStage request. StageName=%s, StageConfig=%s",
		input.Request.StageName,
		string(input.Request.StageConfig),
	))

	// After PLAN, save the plan file path to metadata file
	if input.Request.StageName == stageOpenTofuPlan {
		logger.Info("[PLUGIN] Saving plan file path to metadata file",
			zap.String("planFile", planFile),
			zap.String("stage", input.Request.StageName),
			zap.String("deploymentID", input.Request.Deployment.ID),
		)

		if err := savePlanFileMetadata(input.Request.Deployment.ID, planFile); err != nil {
			logger.Error("[PLUGIN] Failed to save plan file metadata",
				zap.Error(err),
				zap.String("planFile", planFile),
				zap.String("stage", input.Request.StageName),
				zap.String("deploymentID", input.Request.Deployment.ID),
			)
			return sdk.StageStatusFailure, fmt.Errorf("failed to save plan file metadata: %w", err)
		}

		logger.Info("[PLUGIN] Successfully saved plan file path to metadata file",
			zap.String("planFile", planFile),
			zap.String("stage", input.Request.StageName),
			zap.String("deploymentID", input.Request.Deployment.ID),
		)
	}

	return sdk.StageStatusSuccess, nil
}

// runTofuCommand executes the tofu command for the given stage.
func (p *Plugin) runTofuCommand(tofuPath string, command string, workingDir string, spec *config.OpenTofuApplicationSpec, args ...string) (string, error) {
	// Create command with context
	cmd := exec.CommandContext(context.Background(), tofuPath, append([]string{command}, args...)...)

	// Set environment variables if specified
	if len(spec.Input.Env) > 0 {
		cmd.Env = append(cmd.Env, spec.Input.Env...)
	}

	// Set working directory to the absolute path
	cmd.Dir = workingDir

	// Execute the command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to execute tofu command: %w", err)
	}

	return string(output), nil
}

// Helper function to split file content into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

package deployment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/config"
	"github.com/pipe-cd/pipecd/pkg/plugin/logpersister/logpersistertest"
	"github.com/pipe-cd/pipecd/pkg/plugin/sdk"
	"github.com/pipe-cd/pipecd/pkg/plugin/toolregistry/toolregistrytest"
)

// TestPlugin_executeOpenTofuApplyStage tests the complete OpenTofu apply workflow.
// It first executes a plan stage to generate the plan file, then verifies that the apply stage
// executes successfully using the generated plan. The test ensures proper handling of
// configuration, environment variables, and command execution for both stages.
func TestPlugin_executeOpenTofuApplyStage(t *testing.T) {
	t.Parallel()

	// Initialize tool registry
	testRegistry := toolregistrytest.NewTestToolRegistry(t)

	// Create a valid config
	appCfg := &sdk.ApplicationConfig[config.OpenTofuApplicationSpec]{
		Spec: &config.OpenTofuApplicationSpec{
			Input: config.OpenTofuDeploymentInput{
				Version:    "1.6.0",
				Config:     "main.tf",
				WorkingDir: "testdata/simple",
				Env: []string{
					"TF_VAR_environment=test",
				},
			},
		},
	}

	// First run plan to generate tfplan
	planInput := &sdk.ExecuteStageInput[config.OpenTofuApplicationSpec]{
		Request: sdk.ExecuteStageRequest[config.OpenTofuApplicationSpec]{
			StageName:   stageOpenTofuPlan,
			StageConfig: []byte(``),
			RunningDeploymentSource: sdk.DeploymentSource[config.OpenTofuApplicationSpec]{
				CommitHash: "",
			},
			TargetDeploymentSource: sdk.DeploymentSource[config.OpenTofuApplicationSpec]{
				ApplicationDirectory:      "testdata/simple",
				CommitHash:                "0123456789",
				ApplicationConfig:         appCfg,
				ApplicationConfigFilename: "app.pipecd.yaml",
			},
			Deployment: sdk.Deployment{
				PipedID:       "piped-id",
				ApplicationID: "app-id",
			},
		},
		Client: sdk.NewClient(nil, "opentofu", "", "", logpersistertest.NewTestLogPersister(t), testRegistry),
		Logger: zaptest.NewLogger(t),
	}

	plugin := NewPlugin()

	// Run plan first
	status, err := plugin.executeStage(planInput, appCfg.Spec)
	require.NoError(t, err)
	assert.Equal(t, sdk.StageStatusSuccess, status)

	// Now run apply
	applyInput := &sdk.ExecuteStageInput[config.OpenTofuApplicationSpec]{
		Request: sdk.ExecuteStageRequest[config.OpenTofuApplicationSpec]{
			StageName:   stageOpenTofuApply,
			StageConfig: []byte(``),
			RunningDeploymentSource: sdk.DeploymentSource[config.OpenTofuApplicationSpec]{
				CommitHash: "",
			},
			TargetDeploymentSource: sdk.DeploymentSource[config.OpenTofuApplicationSpec]{
				ApplicationDirectory:      "testdata/simple",
				CommitHash:                "0123456789",
				ApplicationConfig:         appCfg,
				ApplicationConfigFilename: "app.pipecd.yaml",
			},
			Deployment: sdk.Deployment{
				PipedID:       "piped-id",
				ApplicationID: "app-id",
			},
		},
		Client: sdk.NewClient(nil, "opentofu", "", "", logpersistertest.NewTestLogPersister(t), testRegistry),
		Logger: zaptest.NewLogger(t),
	}

	status, err = plugin.executeStage(applyInput, appCfg.Spec)
	require.NoError(t, err)
	assert.Equal(t, sdk.StageStatusSuccess, status)
}

// TestPlugin_executeOpenTofuApplyStage_withInvalidConfig tests the apply stage execution with invalid configuration.
// It verifies that the stage fails appropriately when provided with non-existent configuration files
// and returns the correct error status. This test ensures proper error handling for invalid configurations.
func TestPlugin_executeOpenTofuApplyStage_withInvalidConfig(t *testing.T) {
	t.Parallel()

	// Initialize tool registry
	testRegistry := toolregistrytest.NewTestToolRegistry(t)

	// Create an invalid config
	invalidCfg := &config.OpenTofuApplicationSpec{
		Input: config.OpenTofuDeploymentInput{
			Version:    "1.6.0",
			Config:     "nonexistent.tf",
			WorkingDir: "testdata/simple",
		},
	}

	// Prepare the input
	input := &sdk.ExecuteStageInput[config.OpenTofuApplicationSpec]{
		Request: sdk.ExecuteStageRequest[config.OpenTofuApplicationSpec]{
			StageName:   stageOpenTofuApply,
			StageConfig: []byte(``),
			RunningDeploymentSource: sdk.DeploymentSource[config.OpenTofuApplicationSpec]{
				CommitHash: "",
			},
			TargetDeploymentSource: sdk.DeploymentSource[config.OpenTofuApplicationSpec]{
				ApplicationDirectory:      "testdata/simple",
				CommitHash:                "0123456789",
				ApplicationConfig:         &sdk.ApplicationConfig[config.OpenTofuApplicationSpec]{Spec: invalidCfg},
				ApplicationConfigFilename: "app.pipecd.yaml",
			},
			Deployment: sdk.Deployment{
				PipedID:       "piped-id",
				ApplicationID: "app-id",
			},
		},
		Client: sdk.NewClient(nil, "opentofu", "", "", logpersistertest.NewTestLogPersister(t), testRegistry),
		Logger: zaptest.NewLogger(t),
	}

	plugin := NewPlugin()

	status, err := plugin.executeStage(input, invalidCfg)
	require.Error(t, err)
	assert.Equal(t, sdk.StageStatusFailure, status)
}

package deployment

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/config"
	"github.com/pipe-cd/pipecd/pkg/plugin/logpersister/logpersistertest"
	"github.com/pipe-cd/pipecd/pkg/plugin/sdk"
	"github.com/pipe-cd/pipecd/pkg/plugin/toolregistry/toolregistrytest"
)

func TestBuildQuickSyncStages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected []sdk.QuickSyncStage
	}{
		{
			name: "default stages",
			expected: []sdk.QuickSyncStage{
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, err := buildQuickSyncStages()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual.Stages)
		})
	}
}

func TestBuildPipelineSyncStages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		stages   []sdk.StageConfig
		rollback bool
		expected []sdk.PipelineStage
	}{
		{
			name: "without rollback",
			stages: []sdk.StageConfig{
				{
					Index: 0,
					Name:  "Stage 1",
				},
				{
					Index: 1,
					Name:  "Stage 2",
				},
			},
			rollback: false,
			expected: []sdk.PipelineStage{
				{
					Index:              0,
					Name:               "Stage 1",
					Rollback:           false,
					Metadata:           map[string]string{},
					AvailableOperation: sdk.ManualOperationNone,
				},
				{
					Index:              1,
					Name:               "Stage 2",
					Rollback:           false,
					Metadata:           map[string]string{},
					AvailableOperation: sdk.ManualOperationNone,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, err := buildPipelineSyncStages(&sdk.BuildPipelineSyncStagesInput{
				Request: sdk.BuildPipelineSyncStagesRequest{
					Stages:   tt.stages,
					Rollback: tt.rollback,
				},
			})
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual.Stages)
		})
	}
}

func TestDetermineVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *sdk.DetermineVersionsInput[config.OpenTofuApplicationSpec]
		expected *sdk.DetermineVersionsResponse
	}{
		{
			name: "with commit hash",
			input: &sdk.DetermineVersionsInput[config.OpenTofuApplicationSpec]{
				Request: sdk.DetermineVersionsRequest[config.OpenTofuApplicationSpec]{
					DeploymentSource: sdk.DeploymentSource[config.OpenTofuApplicationSpec]{
						CommitHash:           "abc123",
						ApplicationDirectory: "/app",
					},
				},
			},
			expected: &sdk.DetermineVersionsResponse{
				Versions: []sdk.ArtifactVersion{
					{
						Kind:    sdk.ArtifactKindTerraformModule,
						Version: "v1-abc123",
						Name:    "/app",
						URL:     "/app",
					},
				},
			},
		},
		{
			name: "without commit hash",
			input: &sdk.DetermineVersionsInput[config.OpenTofuApplicationSpec]{
				Request: sdk.DetermineVersionsRequest[config.OpenTofuApplicationSpec]{
					DeploymentSource: sdk.DeploymentSource[config.OpenTofuApplicationSpec]{
						ApplicationDirectory: "/app",
					},
				},
			},
			expected: &sdk.DetermineVersionsResponse{
				Versions: []sdk.ArtifactVersion{
					{
						Kind:    sdk.ArtifactKindTerraformModule,
						Version: "v1",
						Name:    "/app",
						URL:     "/app",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, err := determineVersions(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestDetermineStrategy(t *testing.T) {
	t.Parallel()

	actual, err := determineStrategy()
	assert.NoError(t, err)
	assert.Equal(t, sdk.SyncStrategyQuickSync, actual.Strategy)
	assert.Equal(t, "Using QUICK_SYNC strategy for OpenTofu deployment", actual.Summary)
}

func TestFetchDefinedStages(t *testing.T) {
	t.Parallel()

	expected := []string{
		stageOpenTofuPlan,
		stageOpenTofuApply,
	}
	actual := fetchDefinedStages()
	assert.Equal(t, expected, actual)
}

func TestPlugin_executeOpenTofuPlanStage(t *testing.T) {
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
				Init: true,
			},
		},
	}

	// Prepare the input
	input := &sdk.ExecuteStageInput[config.OpenTofuApplicationSpec]{
		Request: sdk.ExecuteStageRequest[config.OpenTofuApplicationSpec]{
			StageName:   stageOpenTofuPlan,
			StageConfig: []byte(``),
			RunningDeploymentSource: sdk.DeploymentSource[config.OpenTofuApplicationSpec]{
				CommitHash: "", // Empty commit hash indicates no previous deployment
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

	status, err := plugin.executeStage(input, appCfg.Spec)
	require.NoError(t, err)
	assert.Equal(t, sdk.StageStatusSuccess, status)
}

func TestPlugin_executeOpenTofuPlanStage_withInvalidConfig(t *testing.T) {
	t.Parallel()

	// Initialize tool registry
	testRegistry := toolregistrytest.NewTestToolRegistry(t)

	// Create an invalid config
	invalidCfg := &config.OpenTofuApplicationSpec{
		Input: config.OpenTofuDeploymentInput{
			Version: "1.6.0",
			Config:  "nonexistent.tf",
		},
	}

	// Prepare the input
	input := &sdk.ExecuteStageInput[config.OpenTofuApplicationSpec]{
		Request: sdk.ExecuteStageRequest[config.OpenTofuApplicationSpec]{
			StageName:   stageOpenTofuPlan,
			StageConfig: []byte(``),
			RunningDeploymentSource: sdk.DeploymentSource[config.OpenTofuApplicationSpec]{
				CommitHash: "",
			},
			TargetDeploymentSource: sdk.DeploymentSource[config.OpenTofuApplicationSpec]{
				ApplicationDirectory:      filepath.Join("testdata", "simple"),
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

package deployment

import (
	"testing"

	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/config"
	"github.com/pipe-cd/pipecd/pkg/plugin/sdk"
	"github.com/stretchr/testify/assert"
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
					Name:        stageOpenTofuPlan,
					Description: "Plan OpenTofu changes",
					Rollback:    false,
					Metadata: map[string]string{
						"command": "plan",
					},
					AvailableOperation: sdk.ManualOperationNone,
				},
				{
					Name:        stageOpenTofuApply,
					Description: "Apply OpenTofu changes",
					Rollback:    false,
					Metadata: map[string]string{
						"command": "apply",
					},
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
		{
			name: "with rollback",
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
			rollback: true,
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
				{
					Index:              2,
					Name:               stageOpenTofuRollback,
					Rollback:           true,
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
		stageOpenTofuRollback,
	}
	actual := fetchDefinedStages()
	assert.Equal(t, expected, actual)
}

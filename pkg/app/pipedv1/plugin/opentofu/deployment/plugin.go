package deployment

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/config"
	"github.com/pipe-cd/pipecd/pkg/plugin/sdk"
)

// Plugin implements the sdk.DeploymentPlugin interface.
type Plugin struct{}

// NewPlugin creates a new OpenTofu plugin.
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Ensure the Plugin implements the DeploymentPlugin interface.
var _ sdk.DeploymentPlugin[sdk.ConfigNone, config.OpenTofuDeployTargetConfig, config.OpenTofuApplicationSpec] = (*Plugin)(nil)

func (p *Plugin) Name() string {
	return "opentofu"
}

func (p *Plugin) Version() string {
	return "v1.0.0"
}

func (p *Plugin) DetermineVersions(ctx context.Context, _ *sdk.ConfigNone, input *sdk.DetermineVersionsInput[config.OpenTofuApplicationSpec]) (*sdk.DetermineVersionsResponse, error) {
	return determineVersions(input)
}

func (p *Plugin) DetermineStrategy(ctx context.Context, _ *sdk.ConfigNone, input *sdk.DetermineStrategyInput[config.OpenTofuApplicationSpec]) (*sdk.DetermineStrategyResponse, error) {
	return determineStrategy()
}

func (p *Plugin) BuildQuickSyncStages(ctx context.Context, _ *sdk.ConfigNone, _ *sdk.BuildQuickSyncStagesInput) (*sdk.BuildQuickSyncStagesResponse, error) {
	return buildQuickSyncStages()
}

func (p *Plugin) BuildPipelineSyncStages(ctx context.Context, _ *sdk.ConfigNone, input *sdk.BuildPipelineSyncStagesInput) (*sdk.BuildPipelineSyncStagesResponse, error) {
	return buildPipelineSyncStages(input)
}

func (p *Plugin) FetchDefinedStages() []string {
	return fetchDefinedStages()
}

func (p *Plugin) ExecuteStage(
	ctx context.Context,
	_ *sdk.ConfigNone,
	_ []*sdk.DeployTarget[config.OpenTofuDeployTargetConfig],
	input *sdk.ExecuteStageInput[config.OpenTofuApplicationSpec],
) (*sdk.ExecuteStageResponse, error) {
	var spec config.OpenTofuApplicationSpec
	if err := json.Unmarshal(input.Request.StageConfig, &spec); err != nil {
		return nil, fmt.Errorf("failed to decode application spec: %w", err)
	}

	status, err := p.executeStage(input, &spec)
	if err != nil {
		return nil, err
	}

	return &sdk.ExecuteStageResponse{
		Status: status,
	}, nil
}


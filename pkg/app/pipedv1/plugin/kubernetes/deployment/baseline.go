// Copyright 2025 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deployment

import (
	"context"

	kubeconfig "github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/kubernetes/config"
	sdk "github.com/pipe-cd/piped-plugin-sdk-go"
)

func (p *Plugin) executeK8sBaselineRolloutStage(_ context.Context, input *sdk.ExecuteStageInput[kubeconfig.KubernetesApplicationSpec], _ []*sdk.DeployTarget[kubeconfig.KubernetesDeployTargetConfig]) sdk.StageStatus {
	input.Client.LogPersister().Error("Baseline rollout is not yet implemented")
	return sdk.StageStatusFailure
}

func (p *Plugin) executeK8sBaselineCleanStage(_ context.Context, input *sdk.ExecuteStageInput[kubeconfig.KubernetesApplicationSpec], _ []*sdk.DeployTarget[kubeconfig.KubernetesDeployTargetConfig]) sdk.StageStatus {
	input.Client.LogPersister().Error("Baseline clean is not yet implemented")
	return sdk.StageStatusFailure
}

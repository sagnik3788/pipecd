package config

// OpenTofuDeployTargetConfig represents PipedDeployTarget.Config for opentofu plugin.
type OpenTofuDeployTargetConfig struct {
	Version string `json:"version"` // e.g. "1.6.0"
}

// OpenTofuApplicationSpec defines app configuration for OpenTofu deployment.
type OpenTofuApplicationSpec struct {
	Input     OpenTofuDeploymentInput    `json:"input"`
	QuickSync OpenTofuDeployStageOptions `json:"quickSync"`
}

func (s *OpenTofuApplicationSpec) Validate() error {
	return nil
}

// OpenTofuDeploymentInput is the input for OpenTofu stages.
type OpenTofuDeploymentInput struct {
	Version    string   `json:"version"`
	Config     string   `json:"config"`
	WorkingDir string   `json:"workingDir"`
	Env        []string `json:"env"`
	Init       bool     `json:"init"`
}

// Options for quick sync stage
type OpenTofuDeployStageOptions struct {
	AutoApprove bool `json:"autoApprove"`
}

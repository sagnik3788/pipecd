package config

// OpenTofuDeployTargetConfig represents PipedDeployTarget.Config for opentofu plugin.
type OpenTofuDeployTargetConfig struct {
	Version string `json:"version"` // e.g. "1.6.0"
	// You can add backend, init flags etc. later
}

// OpenTofuApplicationSpec defines app configuration for OpenTofu deployment.
type OpenTofuApplicationSpec struct {
	Input     OpenTofuDeploymentInput    `json:"input"`
	QuickSync OpenTofuDeployStageOptions `json:"quickSync"`
}

func (s *OpenTofuApplicationSpec) Validate() error {
	// Add validations here if needed
	return nil
}

// OpenTofuDeploymentInput is the input for OpenTofu stages.
type OpenTofuDeploymentInput struct {
	// Version is the version of OpenTofu to use. e.g. "1.6.0"
	Version string `json:"version"`
	// Config is the path to the OpenTofu config file
	Config string `json:"config"`
	// WorkingDir is the working directory for OpenTofu commands
	WorkingDir string `json:"workingDir"`
	// Env is a list of environment variables to set for OpenTofu commands
	Env  []string `json:"env"`
	Init bool     `json:"init"`
}

// Options for quick sync stage
type OpenTofuDeployStageOptions struct {
	AutoApprove bool `json:"autoApprove"`
}

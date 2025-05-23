---
date: 2024-03-20
title: "Developing a PipeCD Plugin: A Practical Guide"
linkTitle: "Developing a PipeCD Plugin: A Practical Guide"
weight: 986
author: PipeCD Team
categories: ["Guide"]
tags: ["Plugin", "Development"]
---

In this article, we'll walk through the process of developing a PipeCD plugin from scratch. We'll focus on practical implementation and provide real-world examples to help you create your own plugin.

## Prerequisites

Before we begin, ensure you have:
- Go 1.21 or later installed
- Basic understanding of gRPC
- Familiarity with PipeCD concepts

## Plugin Structure

A PipeCD plugin consists of several key components:

1. **Plugin Definition**: The main plugin interface that defines your plugin's capabilities
2. **Pipeline Stages**: The deployment stages your plugin will handle
3. **Configuration**: How users configure your plugin

Let's explore each component in detail.

## 1. Plugin Definition

Your plugin needs to implement the `Plugin` interface from the PipeCD SDK:

```go
type Plugin interface {
    // GetPluginInfo returns the information about this plugin.
    GetPluginInfo() *PluginInfo
    // Register registers this plugin to the given registry.
    Register(registry *Registry) error
}
```

Here's a practical example:

```go
type MyPlugin struct {
    name    string
    version string
    config  *Config
}

func (p *MyPlugin) GetPluginInfo() *PluginInfo {
    return &PluginInfo{
        Name:    p.name,
        Version: p.version,
    }
}

func (p *MyPlugin) Register(registry *Registry) error {
    // Register your plugin's capabilities here
    return nil
}
```

## 2. Pipeline Stages

Your plugin needs to define the stages it will handle during deployment. The main stages are:

- `SYNC`: Deploy the application
- `ROLLBACK`: Rollback to a previous version
- `ANALYSIS`: Run analysis (if needed)

Here's a practical implementation of deployment stages:

```go
func (p *MyPlugin) Register(registry *Registry) error {
    // Register SYNC stage
    registry.RegisterDeploymentHandler(
        "my-platform",
        func(ctx context.Context, req *DeploymentRequest) (*DeploymentResponse, error) {
            // 1. Validate configuration
            if err := p.validateConfig(req.Config); err != nil {
                return nil, fmt.Errorf("invalid config: %w", err)
            }

            // 2. Execute deployment
            if err := p.executeDeployment(ctx, req); err != nil {
                return nil, fmt.Errorf("deployment failed: %w", err)
            }

            // 3. Return success response
            return &DeploymentResponse{
                Status: DeploymentStatus_SUCCESS,
                Metadata: map[string]string{
                    "deployment_id": req.DeploymentID,
                    "stage":        "SYNC",
                },
            }, nil
        },
    )

    // Register ROLLBACK stage
    registry.RegisterRollbackHandler(
        "my-platform",
        func(ctx context.Context, req *RollbackRequest) (*RollbackResponse, error) {
            // 1. Get previous commit
            prevCommit := req.PreviousCommit

            // 2. Execute rollback
            if err := p.executeRollback(ctx, prevCommit); err != nil {
                return nil, fmt.Errorf("rollback failed: %w", err)
            }

            // 3. Return success response
            return &RollbackResponse{
                Status: RollbackStatus_SUCCESS,
            }, nil
        },
    )

    return nil
}

// Helper functions
func (p *MyPlugin) validateConfig(config []byte) error {
    // Implement config validation
    return nil
}

func (p *MyPlugin) executeDeployment(ctx context.Context, req *DeploymentRequest) error {
    // Implement deployment logic
    return nil
}

func (p *MyPlugin) executeRollback(ctx context.Context, prevCommit string) error {
    // Implement rollback logic
    return nil
}
```

## 3. Configuration

Users configure your plugin through the `piped-config.yaml` file. Here's a practical configuration structure:

```go
type Config struct {
    // Platform-specific configuration
    Platform struct {
        Region    string `json:"region"`
        ProjectID string `json:"projectId"`
    } `json:"platform"`

    // Deployment configuration
    Deployment struct {
        Timeout    int    `json:"timeout"`
        RetryCount int    `json:"retryCount"`
        Strategy   string `json:"strategy"`
    } `json:"deployment"`

    // Custom configuration
    Custom map[string]interface{} `json:"custom"`
}
```

Example configuration in `piped-config.yaml`:

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Piped
spec:
  plugins:
    - name: my-plugin
      version: v0.1.0
      config:
        platform:
          region: "us-west1"
          projectId: "my-project"
        deployment:
          timeout: 300
          retryCount: 3
          strategy: "rolling"
        custom:
          featureFlags:
            enableNewUI: true
```

## Practical Example: Complete Plugin

Here's a complete example of a practical plugin:

```go
package main

import (
    "context"
    "fmt"
    "github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/sdk"
)

type MyPlugin struct {
    name    string
    version string
    config  *Config
}

func (p *MyPlugin) GetPluginInfo() *sdk.PluginInfo {
    return &sdk.PluginInfo{
        Name:    p.name,
        Version: p.version,
    }
}

func (p *MyPlugin) Register(registry *sdk.Registry) error {
    // Register deployment handler
    registry.RegisterDeploymentHandler(
        "my-platform",
        func(ctx context.Context, req *sdk.DeploymentRequest) (*sdk.DeploymentResponse, error) {
            // 1. Log deployment start
            sdk.LogInfo("Starting deployment", map[string]string{
                "deployment_id": req.DeploymentID,
                "commit":       req.Commit,
            })

            // 2. Execute deployment
            if err := p.executeDeployment(ctx, req); err != nil {
                sdk.LogError("Deployment failed", err, nil)
                return nil, err
            }

            // 3. Return success
            return &sdk.DeploymentResponse{
                Status: sdk.DeploymentStatus_SUCCESS,
                Metadata: map[string]string{
                    "deployment_id": req.DeploymentID,
                    "commit":       req.Commit,
                },
            }, nil
        },
    )

    // Register rollback handler
    registry.RegisterRollbackHandler(
        "my-platform",
        func(ctx context.Context, req *sdk.RollbackRequest) (*sdk.RollbackResponse, error) {
            // 1. Log rollback start
            sdk.LogInfo("Starting rollback", map[string]string{
                "deployment_id": req.DeploymentID,
                "commit":       req.PreviousCommit,
            })

            // 2. Execute rollback
            if err := p.executeRollback(ctx, req.PreviousCommit); err != nil {
                sdk.LogError("Rollback failed", err, nil)
                return nil, err
            }

            // 3. Return success
            return &sdk.RollbackResponse{
                Status: sdk.RollbackStatus_SUCCESS,
            }, nil
        },
    )

    return nil
}

func (p *MyPlugin) executeDeployment(ctx context.Context, req *sdk.DeploymentRequest) error {
    // Implement your deployment logic here
    // This is where you would:
    // 1. Parse configuration
    // 2. Execute deployment commands
    // 3. Wait for deployment to complete
    // 4. Verify deployment success
    return nil
}

func (p *MyPlugin) executeRollback(ctx context.Context, prevCommit string) error {
    // Implement your rollback logic here
    // This is where you would:
    // 1. Checkout previous commit
    // 2. Execute rollback commands
    // 3. Verify rollback success
    return nil
}

func main() {
    plugin := &MyPlugin{
        name:    "my-plugin",
        version: "v0.1.0",
    }
    sdk.Run(plugin)
}
```

## Best Practices

1. **Error Handling**
   - Always return meaningful error messages
   - Use proper error types from the SDK
   - Log important events using `sdk.LogInfo` and `sdk.LogError`

2. **Configuration**
   - Validate configuration early
   - Provide sensible defaults
   - Document all configuration options

3. **Testing**
   - Write unit tests for your plugin
   - Test different deployment scenarios
   - Test error cases

4. **Logging**
   - Log important events
   - Include relevant metadata in logs
   - Use appropriate log levels

## Documentation

Creating comprehensive documentation for your PipeCD plugin is essential to help users understand how to use it effectively. Here are some tips for writing good documentation:

1. **Overview**: Start with a brief overview of what your plugin does and why it's useful.

2. **Installation**: Provide clear instructions on how to install the plugin, including any prerequisites.

3. **Configuration**: Explain how to configure the plugin, including all available options and their meanings.

4. **Usage**: Include examples of how to use the plugin in different scenarios. Use code snippets to illustrate key points.

5. **Troubleshooting**: Provide a section for common issues and how to resolve them.

6. **Contributing**: If you want others to contribute to your plugin, include guidelines on how to do so.

7. **License**: Specify the license under which your plugin is released.

8. **Contact**: Provide a way for users to contact you for support or feedback.

By following these guidelines, you can create a user-friendly documentation that helps users get the most out of your plugin.

## Conclusion

Developing a PipeCD plugin is straightforward with the SDK. The key is to:
1. Implement the required interfaces
2. Handle deployment and rollback properly
3. Provide clear configuration options
4. Follow best practices for error handling and testing

Remember that your plugin should be:
- Reliable: Handle errors gracefully
- Configurable: Allow users to customize behavior
- Maintainable: Follow Go best practices
- Well-documented: Help users understand how to use it

For more information, check out:
- [PipeCD Plugin SDK Documentation](https://pipecd.dev/docs/plugin-sdk)
- [Example Plugins](https://github.com/pipe-cd/pipecd/tree/master/pkg/app/pipedv1/plugin)
- [Plugin Development Guide](https://pipecd.dev/docs/plugin-development)

If you have questions or need help, join our [community](https://pipecd.dev/community)! 
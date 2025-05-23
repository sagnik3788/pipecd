package main

import (
	"log"

	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/deployment"
	"github.com/pipe-cd/pipecd/pkg/app/pipedv1/plugin/opentofu/livestate"
	"github.com/pipe-cd/pipecd/pkg/plugin/sdk"
)

func main() {
	plugin, err := sdk.NewPlugin(
		"opentofu", "v1.0.0",
		sdk.WithDeploymentPlugin(&deployment.Plugin{}),
		sdk.WithLivestatePlugin(&livestate.Plugin{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}

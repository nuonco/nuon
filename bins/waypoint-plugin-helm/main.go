package main

import (
	"log"

	"github.com/go-playground/validator/v10"
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-helm/internal/platform/helm"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-helm/internal/registry"
)

func main() {
	v := validator.New()

	helmPlugin, err := helm.New(v)
	if err != nil {
		log.Fatalf("unable to create helm plugin: %s", err)
	}

	registryPlugin, err := registry.New(v)
	if err != nil {
		log.Fatalf("unable to initialize registry plugin: %s", err)
	}

	sdk.Main(sdk.WithComponents(
		helmPlugin,
		registryPlugin,
	),
	)
}

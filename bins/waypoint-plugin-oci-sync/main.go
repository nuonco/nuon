package main

import (
	"log"

	"github.com/go-playground/validator/v10"
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-oci-sync/internal/builder"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-oci-sync/internal/registry"
)

func main() {
	v := validator.New()

	buildPlugin, err := builder.New(v)
	if err != nil {
		log.Fatalf("unable to initialize build plugin: %s", err)
	}

	registryPlugin, err := registry.New(v)
	if err != nil {
		log.Fatalf("unable to initialize registry plugin: %s", err)
	}

	sdk.Main(sdk.WithComponents(
		buildPlugin,
		registryPlugin,
	),
	// NOTE(jm): eventually, we are going to move the oci plugin stuff into it's own, and expose an OCI artifact
	// properly.
	//sdk.WithMappers(registry.ImageMapper),
	)
}

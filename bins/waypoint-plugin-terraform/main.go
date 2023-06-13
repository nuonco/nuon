package main

import (
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-terraform/internal/builder"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-terraform/internal/platform"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-terraform/internal/registry"
	"oras.land/oras-go/v2/content/file"
)

const defaultStorePath string = "/tmp/plugin-store"

func getStore() (*file.Store, error) {
	fs, err := file.New(defaultStorePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file store: %w", err)
	}

	return fs, nil
}

func main() {
	v := validator.New()

	store, err := getStore()
	if err != nil {
		log.Fatalf("unable to get store: %s", err)
	}

	platformPlugin, err := platform.New(v, store)
	if err != nil {
		log.Fatalf("unable to initialize platform plugin: %s", err)
	}

	buildPlugin, err := builder.New(v, store)
	if err != nil {
		log.Fatalf("unable to initialize build plugin: %s", err)
	}

	registryPlugin, err := registry.New(v, store)
	if err != nil {
		log.Fatalf("unable to initialize registry plugin: %s", err)
	}

	sdk.Main(sdk.WithComponents(
		buildPlugin,
		registryPlugin,
		platformPlugin,
	),
	// NOTE(jm): eventually, we are going to move the oci plugin stuff into it's own, and expose an OCI artifact
	// properly.
	//sdk.WithMappers(registry.ImageMapper),
	)
}

package main

import (
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-terraform/internal/builder"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-terraform/internal/registry"
)

func main() {
	sdk.Main(sdk.WithComponents(
		&builder.Builder{},
		registry.New(),
	),
		sdk.WithMappers(registry.ImageMapper),
	)
}

package main

import (
	"log"

	"github.com/go-playground/validator/v10"
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-oci-sync/internal/builder"
)

func main() {
	v := validator.New()

	buildPlugin, err := builder.New(v)
	if err != nil {
		log.Fatalf("unable to initialize build plugin: %s", err)
	}

	sdk.Main(sdk.WithComponents(buildPlugin))
}

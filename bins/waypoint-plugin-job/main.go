package main

import (
	"log"

	"github.com/go-playground/validator/v10"
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-job/internal/platform"
)

func main() {
	v := validator.New()

	platformComponent, err := platform.New(v)
	if err != nil {
		log.Fatalf("unable to create job plugin: %s", err)
	}

	sdk.Main(sdk.WithComponents(
		platformComponent,
	))
}

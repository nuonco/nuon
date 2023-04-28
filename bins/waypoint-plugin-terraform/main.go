package main

import (
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-terraform/internal/builder"
)

func main() {
	sdk.Main(sdk.WithComponents(
		&builder.Builder{},
	))
}

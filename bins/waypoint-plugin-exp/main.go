// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-exp/internal/builder"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-exp/internal/platform"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-exp/internal/registry"
)

func main() {
	// sdk.Main allows you to register the components which should
	// be included in your plugin
	// Main sets up all the go-plugin requirements

	sdk.Main(sdk.WithComponents(
		// Comment out any components which are not
		// required for your plugin
		&builder.Builder{},
		&registry.Registry{},
		&platform.Platform{},
		// &platform.Mapper{},
	))
}

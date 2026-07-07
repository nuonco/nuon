// Command terraform-provider-nuon is the Nuon Terraform provider. It exposes
// the nuon_stack resource, which drives the stack SDK to provision install
// stacks locally (BYOC): fetching config from the Nuon control plane, running
// the install-stacks Terraform module against the customer's cloud account.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/nuonco/nuon/bins/terraform-provider-nuon/internal/provider"
)

// version is set at release time via -ldflags.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/nuonco/nuon",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

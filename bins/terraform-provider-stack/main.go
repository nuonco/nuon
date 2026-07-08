// Command terraform-provider-stack is the Nuon Terraform provider. It exposes
// the stack_config data source, which reads an install stack's rendered
// configuration from the Nuon control plane (keyed by phone_home_id) so an
// install-stacks Terraform module can consume it directly instead of tfvars.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/nuonco/nuon/bins/terraform-provider-stack/internal/provider"
)

// version is set at release time via -ldflags.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/nuonco/stack",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

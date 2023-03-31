package cmd

import (
	"context"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := &cobra.Command{
		Use: "nuonctl",
	}
	ctx := context.Background()
	c := &cli{
		v: validator.New(),
	}

	if err := c.init(rootCmd.Flags()); err != nil {
		log.Fatalf("unable to initialize cli: %s", err)
	}

	if err := c.registerDeployments(ctx, rootCmd); err != nil {
		log.Fatalf("unable to register deployments: %s", err)
	}
	if err := c.registerInstalls(ctx, rootCmd); err != nil {
		log.Fatalf("unable to register installs: %s", err)
	}
	if err := c.registerOrgs(ctx, rootCmd); err != nil {
		log.Fatalf("unable to register orgs: %s", err)
	}
	if err := c.registerApps(ctx, rootCmd); err != nil {
		log.Fatalf("unable to register apps: %s", err)
	}
	if err := c.registerGeneral(ctx, rootCmd); err != nil {
		log.Fatalf("unable to register general: %s", err)
	}
	if err := c.registerServices(ctx, rootCmd); err != nil {
		log.Fatalf("unable to register service commands: %s", err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

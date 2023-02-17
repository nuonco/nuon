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
	v := validator.New()

	if err := registerDeployments(ctx, v, rootCmd); err != nil {
		log.Fatalf("unable to register deployments: %s", err)
	}
	if err := registerInstalls(ctx, v, rootCmd); err != nil {
		log.Fatalf("unable to register installs: %s", err)
	}
	if err := registerOrgs(ctx, v, rootCmd); err != nil {
		log.Fatalf("unable to register orgs: %s", err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

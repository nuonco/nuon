package cmds

import (
	"context"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

func Execute() {
	v := validator.New()
	rootCmd := &cobra.Command{
		Use:          "nuonctl",
		SilenceUsage: true,
	}

	uiLog, err := ui.New(v)
	if err != nil {
		log.Fatalf("unable to initialize ui: %s", err)
	}

	ctx := context.Background()
	ctx = ui.WithContext(ctx, uiLog)
	c := &cli{
		v: v,
	}

	if err := c.init(rootCmd.Flags()); err != nil {
		log.Fatalf("unable to initialize cli: %s", err)
	}

	namespaces := map[string]func(context.Context, *cobra.Command) error{
		"all":     c.registerCtl,
		"version": c.registerVersion,
	}
	for ns, fn := range namespaces {
		if err := fn(ctx, rootCmd); err != nil {
			log.Fatalf("unable to initialize %s: %s", ns, err)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

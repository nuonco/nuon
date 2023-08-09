package cmds

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/bins/cli/internal/app"
	"github.com/spf13/cobra"
)

const (
	publicAPIAddress string = "https://api.prod.nuon.co/graphql"
)

type cli struct {
	v *validator.Validate
}

func (c *cli) registerCtl(ctx context.Context, rootCmd *cobra.Command) error {
	cmds, err := app.New(c.v,
		app.WithDefaultEnv(),
		app.WithAPIURL(publicAPIAddress),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize ctl: %w", err)
	}

	// register commands
	rootCmd.AddCommand(&cobra.Command{
		Use: "apps",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.ListApps(ctx)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use: "installs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.ListInstalls(ctx)
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use: "components",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.ListComponents(ctx)
		},
	})

	buildCmdOpts := &app.BuildOpts{}
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "build a component",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.Build(ctx, buildCmdOpts)
		},
	}
	buildCmd.PersistentFlags().StringVar(&buildCmdOpts.ComponentName, "component-name", "", "component name")
	buildCmd.PersistentFlags().StringVar(&buildCmdOpts.ComponentID, "component-id", "", "component id")
	buildCmd.PersistentFlags().StringVar(&buildCmdOpts.GitRef, "git-ref", "HEAD", "component id")
	rootCmd.AddCommand(buildCmd)

	deployOpts := &app.DeployOpts{}
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy to one or all installs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.Deploy(ctx, deployOpts)
		},
	}
	deployCmd.PersistentFlags().BoolVar(&deployOpts.All, "all", false, "all installs")
	deployCmd.PersistentFlags().StringVar(&deployOpts.ComponentID, "component-id", "", "component id")
	deployCmd.PersistentFlags().StringVar(&deployOpts.ComponentName, "component-name", "", "component name")
	deployCmd.PersistentFlags().StringVar(&deployOpts.BuildID, "build-id", "", "build id")
	deployCmd.PersistentFlags().BoolVar(&deployOpts.Latest, "latest", true, "latest")
	rootCmd.AddCommand(deployCmd)

	return nil
}

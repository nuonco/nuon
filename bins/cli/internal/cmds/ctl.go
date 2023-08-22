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

	rootCmd.AddCommand(&cobra.Command{
		Use:   "context",
		Short: "Get current org context",
		Long:  "Get the current org context",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.GetContext(ctx)
		},
	})

	// register commands
	appsCmd := &cobra.Command{
		Use:   "apps",
		Short: "View the apps in your org",
	}

	appsCmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all your apps",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.ListApps(ctx)
		},
	})

	appsCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get the current app",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.GetApp(ctx)
		},
	})

	rootCmd.AddCommand(appsCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use: "legacy-apps",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.LegacyListApps(ctx)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use: "legacy-installs",
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

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "get, view status or create an install",
	}

	createInstallOpts := &app.CreateInstallOpts{}
	createInstallCmd := &cobra.Command{
		Use:   "create",
		Short: "create a new install",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.CreateInstall(ctx, createInstallOpts)
		},
	}
	createInstallCmd.PersistentFlags().StringVar(&createInstallOpts.InstallName, "install-name", "", "install name")
	createInstallCmd.PersistentFlags().StringVar(&createInstallOpts.InstallRegion, "install-region", "", "install region")
	createInstallCmd.PersistentFlags().StringVar(&createInstallOpts.InstallARN, "install-arn", "", "install ARN")
	installCmd.AddCommand(createInstallCmd)

	getInstallOpts := &app.GetInstallOpts{}
	getInstallCmd := &cobra.Command{
		Use:   "get",
		Short: "get an install by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.GetInstall(ctx, getInstallOpts)
		},
	}
	getInstallCmd.PersistentFlags().StringVar(&getInstallOpts.InstallID, "install-id", "", "install ID")
	installCmd.AddCommand(getInstallCmd)

	var installID string
	getInstallStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "check install status by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.GetInstallStatus(ctx, installID)
		},
	}
	getInstallStatusCmd.PersistentFlags().StringVar(&installID, "install-id", "", "install ID")
	installCmd.AddCommand(getInstallStatusCmd)

	rootCmd.AddCommand(installCmd)

	return nil
}

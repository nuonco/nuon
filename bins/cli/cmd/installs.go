package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/installs"
)

func (c *cli) installsCmd() *cobra.Command {
	var (
		id               string
		name             string
		arn              string
		region           string
		appID            string
		deployID         string
		renderedVars     bool
		intermediateOnly bool
		jobConfig        bool
		runID            string
		inputs           []string
	)

	installsCmds := &cobra.Command{
		Use:               "installs",
		Short:             "Manage app installs",
		PersistentPreRunE: c.persistentPreRunE,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installs",
		Long:    "List all your app's installs",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.List(cmd.Context(), appID, PrintJSON)
		},
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter installs by")
	installsCmds.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an install",
		Long:  "Get an install by ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.Get(cmd.Context(), id, PrintJSON)
		},
	}
	getCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	getCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(getCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an install",
		Long:  "Create a new install of your app",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.Create(cmd.Context(), appID, name, region, arn, inputs, PrintJSON)
		},
	}
	createCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app to create this install for")
	createCmd.MarkFlagRequired("app-id")
	createCmd.Flags().StringVarP(&name, "name", "n", "", "The name you want to give this install")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().StringVarP(&arn, "role", "o", "", "The ARN of the IAM role to use to provision this install")
	createCmd.MarkFlagRequired("role")
	createCmd.Flags().StringVarP(&region, "region", "r", "", "The region to provision this install in")
	createCmd.MarkFlagRequired("region")
	createCmd.Flags().StringSliceVar(&inputs, "inputs", []string{}, "The app input values for the install")
	installsCmds.AddCommand(createCmd)

	confirmDelete := false
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete install",
		Long:  "Delete an install by ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.Delete(cmd.Context(), id, PrintJSON)
		},
	}
	deleteCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	deleteCmd.Flags().BoolVar(&confirmDelete, "confirm", false, "Confirm you want to delete the install")
	deleteCmd.MarkFlagRequired("install-id")
	deleteCmd.MarkFlagRequired("confirm")
	installsCmds.AddCommand(deleteCmd)

	componentsCmd := &cobra.Command{
		Use:   "components",
		Short: "Get install components",
		Long:  "Get all components on an install",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.Components(cmd.Context(), id, PrintJSON)
		},
	}
	componentsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	componentsCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(componentsCmd)

	getDeployCmd := &cobra.Command{
		Use:   "get-deploy",
		Short: "Get an install deploy",
		Long:  "Get an install deploy by ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.GetDeploy(cmd.Context(), id, deployID, PrintJSON)
		},
	}
	getDeployCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	getDeployCmd.Flags().StringVarP(&deployID, "deploy-id", "d", "", "The deploy ID for the deploy log you want to view")
	getDeployCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(getDeployCmd)

	printDeployPlan := &cobra.Command{
		Use:   "print-deploy-plan",
		Short: "Print install deploy plan",
		Long:  "Print install deploy plan as JSON",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.PrintDeployPlan(cmd.Context(), id, deployID, PrintJSON, renderedVars, intermediateOnly, jobConfig)
		},
	}
	printDeployPlan.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	printDeployPlan.MarkFlagRequired("install-id")
	printDeployPlan.Flags().StringVarP(&deployID, "deploy-id", "d", "", "The ID of the deploy you want to view")
	printDeployPlan.MarkFlagRequired("deploy-id")
	printDeployPlan.Flags().BoolVar(&renderedVars, "rendered-vars", false, "Print rendered variables from deploy plan")
	printDeployPlan.Flags().BoolVar(&intermediateOnly, "intermediate-only", false, "Print intermediate variables from deploy plan")
	printDeployPlan.Flags().BoolVar(&jobConfig, "print-job-config", false, "Print job config from deploy plan")
	printDeployPlan.MarkFlagsMutuallyExclusive("rendered-vars", "intermediate-only", "print-job-config")
	installsCmds.AddCommand(printDeployPlan)

	deployLogsCmd := &cobra.Command{
		Use:   "deploy-logs",
		Short: "View deploy logs",
		Long:  "View deploy logs by install and deploy ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.DeployLogs(cmd.Context(), id, deployID, PrintJSON)
		},
	}
	deployLogsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install whose deploy you want to view")
	deployLogsCmd.MarkFlagRequired("install-id")
	deployLogsCmd.Flags().StringVarP(&deployID, "deploy-id", "d", "", "The deploy ID for the deploy log you want to view")
	deployLogsCmd.MarkFlagRequired("deploy-id")
	installsCmds.AddCommand(deployLogsCmd)

	listDeploysCmd := &cobra.Command{
		Use:   "list-deploys",
		Short: "View all install deploys",
		Long:  "View all install deploys by install ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.ListDeploys(cmd.Context(), id, PrintJSON)
		},
	}
	listDeploysCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install whose deploy you want to view")
	listDeploysCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(listDeploysCmd)

	sandboxRunsCmd := &cobra.Command{
		Use:   "sandbox-runs",
		Short: "View sandbox runs",
		Long:  "View sandbox runs by install ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.SandboxRuns(cmd.Context(), id, PrintJSON)
		},
	}
	sandboxRunsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	sandboxRunsCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(sandboxRunsCmd)

	sandboxRunLogsCmd := &cobra.Command{
		Use:   "sandbox-run-logs",
		Short: "View sandbox run logs",
		Long:  "View sandbox run logs by run & install IDs",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.SandboxRunLogs(cmd.Context(), id, runID, PrintJSON)
		},
	}
	sandboxRunLogsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	sandboxRunLogsCmd.MarkFlagRequired("install-id")
	sandboxRunLogsCmd.Flags().StringVarP(&runID, "run-id", "r", "", "The ID of the run you want to view")
	sandboxRunLogsCmd.MarkFlagRequired("run-id")
	installsCmds.AddCommand(sandboxRunLogsCmd)

	currentInputs := &cobra.Command{
		Use:   "current-inputs",
		Short: "View current inputs",
		Long:  "View current set app inputs",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.CurrentInputs(cmd.Context(), id, PrintJSON)
		},
	}
	currentInputs.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	currentInputs.MarkFlagRequired("install-id")
	installsCmds.AddCommand(currentInputs)

	selectInstallCmd := &cobra.Command{
		Use:   "select",
		Short: "Select your current install",
		Long:  "Select your current install from a list or by install ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.Select(cmd.Context(), appID, id, PrintJSON)
		},
	}
	selectInstallCmd.Flags().StringVar(&id, "install", "", "The ID of the install you want to use")
	selectInstallCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter installs by")
	installsCmds.AddCommand(selectInstallCmd)

	reprovisionInstallCmd := &cobra.Command{
		Use:   "reprovision",
		Short: "Reproivision install",
		Long:  "Reprovision an install sandbox",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.Reprovision(cmd.Context(), id, PrintJSON)
		},
	}
	reprovisionInstallCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	reprovisionInstallCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(reprovisionInstallCmd)

	deprovisionInstallCmd := &cobra.Command{
		Use:   "deprovision",
		Short: "reprovision install",
		Long:  "Deprovision an install sandbox",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.Deprovision(cmd.Context(), id, PrintJSON)
		},
	}
	deprovisionInstallCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	deprovisionInstallCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(deprovisionInstallCmd)

	teardownInstallComponentsCmd := &cobra.Command{
		Use:   "teardown-components",
		Short: "Teardown components on install.",
		Long:  "Teardown all deployed components on an install",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.TeardownComponents(cmd.Context(), id, PrintJSON)
		},
	}
	teardownInstallComponentsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	teardownInstallComponentsCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(teardownInstallComponentsCmd)

	deployInstallComponentsCmd := &cobra.Command{
		Use:   "deploy-components",
		Short: "Deploy all components to an install.",
		Long:  "Deploy all components to an install.",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient, c.cfg)
			svc.DeployComponents(cmd.Context(), id, PrintJSON)
		},
	}
	deployInstallComponentsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	deployInstallComponentsCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(deployInstallComponentsCmd)

	return installsCmds
}

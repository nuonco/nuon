package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/installs"
)

func (c *cli) installsCmd() *cobra.Command {
	var (
		id            string
		workflowID    string
		name          string
		region        string
		appID         string
		deployID      string
		runID         string
		installCompID string
		componentID   string
		inputs        []string
		noSelect      bool
		deployDeps    bool
		offset        int
		limit         int
		planOnly      bool
		fileOrDir     string
		confirm       bool
		wait          bool
	)

	installsCmds := &cobra.Command{
		Use:               "installs",
		Short:             "Manage installs",
		Aliases:           []string{"i"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           InstallGroup.ID,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installs",
		Long:    "List all your app's installs",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.List(cmd.Context(), appID, offset, limit, PrintJSON)
		}),
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter installs by")
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum installs to return")
	installsCmds.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an install",
		Long:  "Get an install by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Get(cmd.Context(), id, PrintJSON)
		}),
	}
	getCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	getCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(getCmd)

	generateConfigCmd := &cobra.Command{
		Use:   "generate-config",
		Short: "Generate config for an existing install",
		Long:  "Generate config file for an existing install, to be used with a nuon app config",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.GenerateConfig(cmd.Context(), id)
		}),
	}
	generateConfigCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to import")
	generateConfigCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(generateConfigCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an install",
		Long:  "Create a new install of your app",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Create(cmd.Context(), appID, name, region, inputs, PrintJSON, noSelect)
		}),
	}
	createCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app to create this install for")
	createCmd.MarkFlagRequired("app-id")
	createCmd.Flags().StringVarP(&name, "name", "n", "", "The name you want to give this install")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().StringVarP(&region, "region", "r", "", "The region to provision this install in")
	createCmd.MarkFlagRequired("region")
	createCmd.Flags().StringSliceVar(&inputs, "inputs", []string{}, "The app input values for the install")
	createCmd.Flags().BoolVar(&noSelect, "no-select", false, "Do not automatically set the created install as the current install")
	installsCmds.AddCommand(createCmd)

	confirmDelete := false
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete install",
		Long:  "Delete an install by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Delete(cmd.Context(), id, PrintJSON)
		}),
	}
	deleteCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	deleteCmd.Flags().BoolVar(&confirmDelete, "confirm", false, "Confirm you want to delete the install")
	deleteCmd.MarkFlagRequired("install-id")
	deleteCmd.MarkFlagRequired("confirm")
	installsCmds.AddCommand(deleteCmd)

	confirmForget := false
	forgetCmd := &cobra.Command{
		Use:   "forget",
		Short: "Forget install",
		Long:  "Forget an install by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Forget(cmd.Context(), id, PrintJSON)
		}),
	}
	forgetCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to forget")
	forgetCmd.Flags().BoolVar(&confirmForget, "confirm", false, "Confirm you want to forget the install")
	forgetCmd.MarkFlagRequired("install-id")
	forgetCmd.MarkFlagRequired("confirm")
	installsCmds.AddCommand(forgetCmd)

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync install",
		Long:  "Sync install(s) with the help of config files",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Sync(cmd.Context(), fileOrDir, appID, confirm, wait)
		}),
	}
	syncCmd.Flags().StringVarP(&fileOrDir, "file", "d", "", "Path to an install config file or a directory with install config files to sync")
	syncCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the install belongs to")
	syncCmd.Flags().BoolVarP(&confirm, "yes", "y", false, "Set to automatically approve diffs and workflows for synced installs")
	syncCmd.Flags().BoolVarP(&wait, "wait", "w", false, "Set to wait for workflows to complete after syncing installs")
	syncCmd.MarkFlagRequired("file")
	syncCmd.MarkFlagRequired("app-id")
	installsCmds.AddCommand(syncCmd)

	componentsCmd := &cobra.Command{
		Use:   "components",
		Short: "Get install components",
		Long:  "Get all components on an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Components(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	componentsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	componentsCmd.MarkFlagRequired("install-id")
	componentsCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	componentsCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum components to return")
	installsCmds.AddCommand(componentsCmd)

	getDeployCmd := &cobra.Command{
		Use:   "get-deploy",
		Short: "Get an install deploy",
		Long:  "Get an install deploy by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.GetDeploy(cmd.Context(), id, deployID, PrintJSON)
		}),
	}
	getDeployCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	getDeployCmd.Flags().StringVarP(&deployID, "deploy-id", "d", "", "The deploy ID for the deploy log you want to view")
	getDeployCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(getDeployCmd)

	createDeployCmd := &cobra.Command{
		Use:   "create-deploy",
		Short: "Create an install deploy",
		Long:  "Create an install deploy by install ID and build ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.CreateDeploy(cmd.Context(), id, deployID, deployDeps, PrintJSON)
		}),
	}
	createDeployCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	createDeployCmd.MarkFlagRequired("install-id")
	createDeployCmd.Flags().StringVarP(&deployID, "build-id", "b", "", "The build ID for the deploy you want to create")
	createDeployCmd.MarkFlagRequired("build-id")
	createDeployCmd.Flags().BoolVar((&deployDeps), "dependents", false, "Deploy dependents")
	installsCmds.AddCommand(createDeployCmd)

	deployLogsCmd := &cobra.Command{
		Use:   "deploy-logs",
		Short: "View deploy logs",
		Long:  "View deploy logs by install and deploy ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.DeployLogs(cmd.Context(), id, deployID, installCompID, PrintJSON)
		}),
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
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.ListDeploys(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	listDeploysCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install whose deploy you want to view")
	listDeploysCmd.MarkFlagRequired("install-id")
	listDeploysCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listDeploysCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum deploys to return")
	installsCmds.AddCommand(listDeploysCmd)

	sandboxRunsCmd := &cobra.Command{
		Use:   "sandbox-runs",
		Short: "View sandbox runs",
		Long:  "View sandbox runs by install ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.SandboxRuns(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	sandboxRunsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	sandboxRunsCmd.MarkFlagRequired("install-id")
	sandboxRunsCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	sandboxRunsCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum runs to return")
	installsCmds.AddCommand(sandboxRunsCmd)

	sandboxRunLogsCmd := &cobra.Command{
		Use:   "sandbox-run-logs",
		Short: "View sandbox run logs",
		Long:  "View sandbox run logs by run & install IDs",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.SandboxRunLogs(cmd.Context(), id, runID, PrintJSON)
		}),
	}
	sandboxRunLogsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	sandboxRunLogsCmd.MarkFlagRequired("install-id")
	sandboxRunLogsCmd.Flags().StringVarP(&runID, "run-id", "r", "", "The ID of the run you want to view")
	sandboxRunLogsCmd.MarkFlagRequired("run-id")
	sandboxRunLogsCmd.Flags().StringVarP(&installCompID, "install-comp-id", "c", "", "The ID of the install component to view logs for")
	installsCmds.AddCommand(sandboxRunLogsCmd)

	currentInputs := &cobra.Command{
		Use:   "current-inputs",
		Short: "View current inputs",
		Long:  "View current set app inputs",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.CurrentInputs(cmd.Context(), id, PrintJSON)
		}),
	}
	currentInputs.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	currentInputs.MarkFlagRequired("install-id")
	installsCmds.AddCommand(currentInputs)

	selectInstallCmd := &cobra.Command{
		Use:   "select",
		Short: "Select your current install",
		Long:  "Select your current install from a list or by install ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Select(cmd.Context(), appID, id, PrintJSON)
		}),
	}
	selectInstallCmd.Flags().StringVar(&id, "install", "", "The ID of the install you want to use")
	selectInstallCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter installs by")
	installsCmds.AddCommand(selectInstallCmd)

	unsetCurrentInstallCmd := &cobra.Command{
		Use:   "unset-current",
		Short: "Unset your current install selection",
		Long:  "Unset your current install selection.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.UnsetCurrent(cmd.Context())
		}),
	}
	installsCmds.AddCommand(unsetCurrentInstallCmd)

	reprovisionInstallCmd := &cobra.Command{
		Use:   "reprovision",
		Short: "Reprovision install",
		Long:  "Reprovision an install sandbox",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Reprovision(cmd.Context(), id, PrintJSON)
		}),
	}
	reprovisionInstallCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	reprovisionInstallCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(reprovisionInstallCmd)

	deprovisionInstallCmd := &cobra.Command{
		Use:   "deprovision",
		Short: "Deprovision install",
		Long:  "Deprovision an install sandbox",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Deprovision(cmd.Context(), id, PrintJSON)
		}),
	}
	deprovisionInstallCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	deprovisionInstallCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(deprovisionInstallCmd)

	teardownInstallComponentsCmd := &cobra.Command{
		Use:   "teardown-components",
		Short: "Teardown components on install.",
		Long:  "Teardown all deployed components on an install (deprecated)",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.TeardownComponents(cmd.Context(), id, PrintJSON)
		}),
	}
	teardownInstallComponentsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	teardownInstallComponentsCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(teardownInstallComponentsCmd)

	teardownInstallComponentCmd := &cobra.Command{
		Use:   "teardown-component",
		Short: "Teardown component on install.",
		Long:  "Teardown all deployed components on an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.TeardownComponent(cmd.Context(), id, componentID, PrintJSON)
		}),
	}
	teardownInstallComponentCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	teardownInstallComponentCmd.MarkFlagRequired("install-id")
	teardownInstallComponentCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The ID of the component you want to teardown")
	teardownInstallComponentCmd.MarkFlagRequired("component-id")
	installsCmds.AddCommand(teardownInstallComponentCmd)

	deployInstallComponentsCmd := &cobra.Command{
		Use:   "deploy-components",
		Short: "Deploy all components to an install.",
		Long:  "Deploy all components to an install.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.DeployComponents(cmd.Context(), id, planOnly, PrintJSON)
		}),
	}
	deployInstallComponentsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	deployInstallComponentsCmd.MarkFlagRequired("install-id")
	deployInstallComponentsCmd.Flags().BoolVar(&planOnly, "plan-only", false, "Only plan, do not actually deploy")
	installsCmds.AddCommand(deployInstallComponentsCmd)

	updateInputCmd := &cobra.Command{
		Use:   "update-input",
		Short: "Update install input",
		Long:  "Update an install input value",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.UpdateInput(cmd.Context(), id, inputs, PrintJSON)
		}),
	}
	updateInputCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to update")
	updateInputCmd.MarkFlagRequired("install-id")
	updateInputCmd.Flags().StringSliceVar(&inputs, "inputs", []string{}, "The app input values for the install")
	updateInputCmd.MarkFlagRequired("inputs")
	installsCmds.AddCommand(updateInputCmd)

	deprovisionInstallSandboxCmd := &cobra.Command{
		Use:   "deprovision-sandbox",
		Short: "Deprovision install sandbox",
		Long:  "Deprovision an install sandbox",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.DeprovisionSandbox(cmd.Context(), id, PrintJSON)
		}),
	}
	deprovisionInstallSandboxCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	deprovisionInstallSandboxCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(deprovisionInstallSandboxCmd)

	workflowsCmd := &cobra.Command{
		Use:   "workflows",
		Short: "View workflows",
		Long:  "View workflows by install ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Workflows(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	workflowsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	workflowsCmd.MarkFlagRequired("install-id")
	workflowsCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	workflowsCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum workflows to return")
	installsCmds.AddCommand(workflowsCmd)

	// workflows get
	workflowGetCmd := &cobra.Command{
		Use:   "workflows-get",
		Short: "Get one workflows",
		Long:  "View one workflows by install ID and workflow ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.WorkflowGet(cmd.Context(), id, workflowID)
		}),
	}
	workflowGetCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the workflow you want to view")
	workflowGetCmd.MarkFlagRequired("install-id")
	workflowGetCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of the workflow you want to view")
	workflowGetCmd.MarkFlagRequired("workflow-id")
	installsCmds.AddCommand(workflowGetCmd)

	return installsCmds
}

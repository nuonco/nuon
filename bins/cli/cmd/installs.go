package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/installs"
)

func (c *cli) installsCmd() *cobra.Command {
	var (
		id            string
		workflowID    string
		stepID        string
		note          string
		name          string
		region        string
		appID         string
		deployID      string
		runID         string
		installCompID string
		componentID   string
		roleName      string
		inputs        []string
		noSelect      bool
		deployDeps    bool
		offset        int
		limit         int
		planOnly      bool
		fileOrDir     string
		confirm       bool
		wait          bool
		enable        bool
		disable       bool
		dryRun        bool
		skipConfirm   bool
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

	currentCmd := &cobra.Command{
		Use:        "current",
		Deprecated: "Use `nuon installs get` instead",
		Short:      "Get current install (deprecated)",
		Hidden:     true,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Get(cmd.Context(), c.cfg.GetString("install_id"), PrintJSON)
		}),
	}
	installsCmds.AddCommand(currentCmd)

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
		Use:         "create",
		Short:       "Create an install",
		Long:        "Create a new install of your app",
		Annotations: tuiAnnotation(TUIAltScreen),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Create(cmd.Context(), appID, name, region, inputs, PrintJSON, noSelect)
		}),
	}
	createCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app to create this install for")
	createCmd.Flags().StringVarP(&name, "name", "n", "", "The name you want to give this install")

	if !c.cfg.Preview {
		createCmd.MarkFlagRequired("name")
	}
	createCmd.Flags().StringVarP(&region, "region", "r", "", "The region to provision this install in")
	if !c.cfg.Preview {
		createCmd.MarkFlagRequired("region")
	}
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
			return svc.Sync(cmd.Context(), fileOrDir, appID, confirm, wait, dryRun)
		}),
	}
	syncCmd.Flags().StringVarP(&fileOrDir, "file", "d", "", "Path to an install config file or a directory with install config files to sync")
	syncCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the install belongs to")
	syncCmd.Flags().BoolVarP(&confirm, "yes", "y", false, "Set to automatically approve diffs and workflows for synced installs")
	syncCmd.Flags().BoolVarP(&wait, "wait", "w", false, "Set to wait for workflows to complete after syncing installs")
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "If set the changes will not be applied, only the diffs will be shown")
	syncCmd.MarkFlagRequired("file")
	syncCmd.MarkFlagRequired("app-id")
	installsCmds.AddCommand(syncCmd)

	toggleSyncCmd := &cobra.Command{
		Use:   "toggle-sync",
		Short: "Enable/disable install config sync",
		Long:  "Toggle syncing of install using a config file",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.ToggleSync(cmd.Context(), id, enable, disable)
		}),
	}
	toggleSyncCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to toggle config file syncing for")
	toggleSyncCmd.Flags().BoolVar(&enable, "enable", false, "Set to explicitly enable config file syncing for an install")
	toggleSyncCmd.Flags().BoolVar(&disable, "disable", false, "Set to explicitly disable config file syncing for an install")
	toggleSyncCmd.MarkFlagRequired("install-id")
	toggleSyncCmd.MarkFlagsMutuallyExclusive("enable", "disable")
	installsCmds.AddCommand(toggleSyncCmd)

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
		Use:         "deploy-logs",
		Short:       "View deploy logs",
		Long:        "View deploy logs by install and deploy ID",
		Annotations: tuiAnnotation(TUIAltScreen),
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
		Use:         "select",
		Short:       "Select your current install",
		Long:        "Select your current install from a list or by install ID",
		Annotations: tuiAnnotation(TUIContextual),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Select(cmd.Context(), appID, id, PrintJSON)
		}),
	}
	selectInstallCmd.Flags().StringVar(&id, "install", "", "The ID of the install you want to use")
	selectInstallCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter installs by")
	installsCmds.AddCommand(selectInstallCmd)

	deselectInstallCmd := &cobra.Command{
		Use:   "deselect",
		Short: "Deselect your current install",
		Long:  "Deselect your current install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Deselect(cmd.Context())
		}),
	}
	installsCmds.AddCommand(deselectInstallCmd)

	unsetCurrentInstallCmd := &cobra.Command{
		Use:        "unset-current",
		Deprecated: "Use `nuon installs deselect` instead",
		Short:      "Unset your current install selection (deprecated)",
		Hidden:     true,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Deselect(cmd.Context())
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
			return svc.TeardownComponent(cmd.Context(), id, componentID, roleName, PrintJSON)
		}),
	}
	teardownInstallComponentCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	teardownInstallComponentCmd.MarkFlagRequired("install-id")
	teardownInstallComponentCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The ID of the component you want to teardown")
	teardownInstallComponentCmd.MarkFlagRequired("component-id")
	teardownInstallComponentCmd.Flags().StringVar(&roleName, "role-name", "", "IAM role name to use for component teardown")
	installsCmds.AddCommand(teardownInstallComponentCmd)

	deployInstallComponentsCmd := &cobra.Command{
		Use:   "deploy-components",
		Short: "Deploy all components to an install.",
		Long:  "Deploy all components to an install.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.DeployComponents(cmd.Context(), id, roleName, planOnly, PrintJSON)
		}),
	}
	deployInstallComponentsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	deployInstallComponentsCmd.MarkFlagRequired("install-id")
	deployInstallComponentsCmd.Flags().BoolVar(&planOnly, "plan-only", false, "Only plan, do not actually deploy")
	deployInstallComponentsCmd.Flags().StringVar(&roleName, "role-name", "", "IAM role name to use for component deploys")
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

	var skipComponents bool
	reprovisionInstallSandboxCmd := &cobra.Command{
		Use:         "reprovision-sandbox",
		Short:       "Reprovision install sandbox [preview]",
		Long:        "Reprovision an install sandbox",
		Annotations: tuiAnnotation(TUIAltScreen),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := c.persistentPreRunE(cmd, args); err != nil {
				return err
			}
			if !c.cfg.Preview {
				return fmt.Errorf("[NUON_PREVIEW=false] reprovision-sandbox is a preview feature, set NUON_PREVIEW=true to enable")
			}
			return nil
		},
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.ReprovisionSandbox(cmd.Context(), id, skipComponents, PrintJSON)
		}),
	}
	reprovisionInstallSandboxCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use (shows selector if omitted)")
	reprovisionInstallSandboxCmd.Flags().BoolVar(&skipComponents, "skip-components", false, "Skip deploying components after reprovisioning the sandbox")
	installsCmds.AddCommand(reprovisionInstallSandboxCmd)

	var autoRetry bool
	workflowsCmd := &cobra.Command{
		Use:   "workflows",
		Short: "Manage workflows",
		Long: `Manage and view workflows by install ID.

By default, launches an interactive TUI to view workflows.`,
		Args:        cobra.NoArgs,
		Annotations: tuiAnnotation(TUIAltScreen),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.WorkflowsTUI(cmd.Context(), id, workflowID, PrintJSON, autoRetry)
		}),
	}
	workflowsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	workflowsCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of a specific workflow to view")
	workflowsCmd.Flags().BoolVarP(&autoRetry, "auto-retry", "r", false, "Automatically retry failed steps")
	installsCmds.AddCommand(workflowsCmd)

	workflowsListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List workflows",
		Long:    "List all workflows for an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.WorkflowsList(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	workflowsListCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	workflowsListCmd.MarkFlagRequired("install-id")
	workflowsListCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	workflowsListCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum workflows to return")
	workflowsCmd.AddCommand(workflowsListCmd)

	workflowsGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a workflow",
		Long:  "Get workflow details including steps summary",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			wfID := workflowID
			if wfID == "" {
				wfID = svc.GetWorkflowID()
			}
			if wfID == "" {
				return fmt.Errorf("workflow-id is required, use --workflow-id or 'workflows select' to set one")
			}
			return svc.WorkflowsGet(cmd.Context(), wfID, PrintJSON)
		}),
	}
	workflowsGetCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of the workflow (uses selected workflow if not provided)")
	workflowsCmd.AddCommand(workflowsGetCmd)

	workflowsSelectCmd := &cobra.Command{
		Use:         "select",
		Short:       "Select a workflow",
		Long:        "Select a workflow to use as default for subsequent commands",
		Annotations: tuiAnnotation(TUIContextual),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.WorkflowsSelect(cmd.Context(), id, workflowID, offset, limit, PrintJSON)
		}),
	}
	workflowsSelectCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	workflowsSelectCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of the workflow to select directly")
	workflowsSelectCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	workflowsSelectCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum workflows to return")
	workflowsCmd.AddCommand(workflowsSelectCmd)

	workflowsDeselectCmd := &cobra.Command{
		Use:   "deselect",
		Short: "Deselect the current workflow",
		Long:  "Clear the currently selected workflow",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.WorkflowsDeselect(cmd.Context())
		}),
	}
	workflowsCmd.AddCommand(workflowsDeselectCmd)

	workflowsWatchCmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch workflows in a full-screen TUI",
		Long: `Launch a full-screen TUI to watch all workflows for an install.

The TUI displays a list of workflows with auto-refresh every 5 seconds.
Select a workflow to view details, and press 'o' to open in browser.

Exit codes:
  0 - Success (user quit normally)
  1 - Error
  130 - Interrupted (ctrl+c)

Examples:
  # Watch workflows for an install
  nuon installs workflows watch -i myinstall

  # Watch using a workflow ID (resolves install from workflow)
  nuon installs workflows watch -w wfl123abc

  # Uses selected workflow from 'workflows select' if no flags provided
  nuon installs workflows watch`,
		Annotations: tuiAnnotation(TUIAltScreen),
		Run: c.wrapCmdWithExitCode(func(cmd *cobra.Command, _ []string) (int, error) {
			svc := installs.New(c.apiClient, c.cfg)

			// Try to get workflow ID from flag or config
			wfID := workflowID
			if wfID == "" {
				wfID = svc.GetWorkflowID()
			}

			return svc.WorkflowsWatchTUI(cmd.Context(), id, wfID)
		}),
	}
	workflowsWatchCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	workflowsWatchCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of a workflow (resolves install automatically)")
	workflowsCmd.AddCommand(workflowsWatchCmd)

	stepsCmd := &cobra.Command{
		Use:   "steps",
		Short: "Manage workflow steps",
		Long:  "View and manage workflow steps",
	}
	workflowsCmd.AddCommand(stepsCmd)

	// Helper to get workflow ID from flag or config
	getWorkflowID := func(svc *installs.Service) (string, error) {
		wfID := workflowID
		if wfID == "" {
			wfID = svc.GetWorkflowID()
		}
		if wfID == "" {
			return "", fmt.Errorf("workflow-id is required, use --workflow-id or 'workflows select' to set one")
		}
		return wfID, nil
	}

	stepsListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List workflow steps",
		Long:    "List all steps for a workflow",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			wfID, err := getWorkflowID(svc)
			if err != nil {
				return err
			}
			return svc.WorkflowStepsList(cmd.Context(), wfID, PrintJSON)
		}),
	}
	stepsListCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of the workflow (uses selected workflow if not provided)")
	stepsCmd.AddCommand(stepsListCmd)

	stepsGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a workflow step",
		Long:  "Get detailed information about a workflow step",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			wfID, err := getWorkflowID(svc)
			if err != nil {
				return err
			}
			stepID, _ := cmd.Flags().GetString("step-id")
			return svc.WorkflowStepsGet(cmd.Context(), wfID, stepID, PrintJSON)
		}),
	}
	stepsGetCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of the workflow (uses selected workflow if not provided)")
	stepsGetCmd.Flags().StringP("step-id", "s", "", "The ID of the step (defaults to latest)")
	stepsCmd.AddCommand(stepsGetCmd)

	stepsPlanCmd := &cobra.Command{
		Use:   "plan",
		Short: "View step plan",
		Long:  "View the deploy plan for a workflow step",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			wfID, err := getWorkflowID(svc)
			if err != nil {
				return err
			}
			return svc.WorkflowStepPlan(cmd.Context(), id, wfID, stepID, PrintJSON)
		}),
	}
	stepsPlanCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	stepsPlanCmd.MarkFlagRequired("install-id")
	stepsPlanCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of the workflow (uses selected workflow if not provided)")
	stepsPlanCmd.Flags().StringVarP(&stepID, "step-id", "s", "", "The ID of the step (defaults to latest)")
	stepsCmd.AddCommand(stepsPlanCmd)

	var (
		logsFollow    bool
		logsRaw       bool
		logsBrowser   bool
		logsLimit     int
		logsFilter    string
		logsSeverity  []string
		logsService   []string
		logsSortOrder string
	)
	stepsLogsCmd := &cobra.Command{
		Use:   "logs",
		Short: "View step logs",
		Long: `View execution logs for a workflow step. Supports deploy, action workflow run, and sandbox run steps.

Filtering examples:
  # Show only error logs
  nuon installs workflows steps logs -i myinstall --severity Error

  # Show only runner service logs at warn or error level
  nuon installs workflows steps logs -i myinstall --severity Warn,Error --service runner

  # Search for a keyword, sorted oldest first
  nuon installs workflows steps logs -i myinstall --filter "timeout" --sort asc

Available severity levels: Trace, Debug, Info, Warn, Error, Fatal
Available service names: api, runner (or any service name present in the logs)`,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			wfID, err := getWorkflowID(svc)
			if err != nil {
				return err
			}
			return svc.WorkflowStepLogs(cmd.Context(), id, wfID, stepID, PrintJSON, installs.WorkflowStepLogsOptions{
				Follow:    logsFollow,
				Raw:       logsRaw,
				Browser:   logsBrowser,
				Limit:     logsLimit,
				Filter:    logsFilter,
				Severity:  logsSeverity,
				Service:   logsService,
				SortOrder: logsSortOrder,
			})
		}),
	}
	stepsLogsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	stepsLogsCmd.MarkFlagRequired("install-id")
	stepsLogsCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of the workflow (uses selected workflow if not provided)")
	stepsLogsCmd.Flags().StringVarP(&stepID, "step-id", "s", "", "The ID of the step (defaults to latest)")
	stepsLogsCmd.Flags().BoolVarP(&logsFollow, "tail", "t", false, "Tail logs in real-time (stream until log stream closes)")
	stepsLogsCmd.Flags().BoolVar(&logsRaw, "raw", false, "Print plain text log lines (useful for piping)")
	stepsLogsCmd.Flags().BoolVar(&logsBrowser, "browser", false, "Open logs in the dashboard UI instead")
	stepsLogsCmd.Flags().IntVarP(&logsLimit, "limit", "n", 0, "Maximum number of log lines to display (0 for all)")
	stepsLogsCmd.Flags().StringVar(&logsFilter, "filter", "", "Filter log lines by substring match on the log body")
	stepsLogsCmd.Flags().StringSliceVar(&logsSeverity, "severity", nil, "Filter by severity level (Trace, Debug, Info, Warn, Error, Fatal)")
	stepsLogsCmd.Flags().StringSliceVar(&logsService, "service", nil, "Filter by service name (e.g., api, runner)")
	stepsLogsCmd.Flags().StringVar(&logsSortOrder, "sort", "", "Sort order by timestamp: asc (oldest first) or desc (newest first)")
	stepsCmd.AddCommand(stepsLogsCmd)

	stepsApproveCmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve a step",
		Long:  "Approve a waiting workflow step. If step-id is not provided, uses the latest step and prompts for confirmation.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			wfID, err := getWorkflowID(svc)
			if err != nil {
				return err
			}
			return svc.WorkflowStepApprove(cmd.Context(), id, wfID, stepID, note, skipConfirm, PrintJSON)
		}),
	}
	stepsApproveCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install (used for plan display)")
	stepsApproveCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of the workflow (uses selected workflow if not provided)")
	stepsApproveCmd.Flags().StringVarP(&stepID, "step-id", "s", "", "The ID of the step (defaults to latest)")
	stepsApproveCmd.Flags().StringVarP(&note, "note", "n", "", "Optional note for the approval")
	stepsApproveCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt when using latest step")
	stepsCmd.AddCommand(stepsApproveCmd)

	stepsRejectCmd := &cobra.Command{
		Use:   "reject",
		Short: "Reject a step",
		Long:  "Reject a waiting workflow step. If step-id is not provided, uses the latest step and prompts for confirmation.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			wfID, err := getWorkflowID(svc)
			if err != nil {
				return err
			}
			return svc.WorkflowStepReject(cmd.Context(), id, wfID, stepID, note, skipConfirm, PrintJSON)
		}),
	}
	stepsRejectCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install (used for plan display)")
	stepsRejectCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of the workflow (uses selected workflow if not provided)")
	stepsRejectCmd.Flags().StringVarP(&stepID, "step-id", "s", "", "The ID of the step (defaults to latest)")
	stepsRejectCmd.Flags().StringVarP(&note, "note", "n", "", "Optional note for the rejection")
	stepsRejectCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt when using latest step")
	stepsCmd.AddCommand(stepsRejectCmd)

	stepsRetryCmd := &cobra.Command{
		Use:   "retry",
		Short: "Retry a step",
		Long:  "Retry a failed workflow step. If step-id is not provided, uses the latest step and prompts for confirmation.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			wfID, err := getWorkflowID(svc)
			if err != nil {
				return err
			}
			return svc.WorkflowStepRetry(cmd.Context(), id, wfID, stepID, skipConfirm, PrintJSON)
		}),
	}
	stepsRetryCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install (used for plan display)")
	stepsRetryCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of the workflow (uses selected workflow if not provided)")
	stepsRetryCmd.Flags().StringVarP(&stepID, "step-id", "s", "", "The ID of the step (defaults to latest)")
	stepsRetryCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt when using latest step")
	stepsCmd.AddCommand(stepsRetryCmd)

	approveAll := false
	promptApproval := false
	setApprovalOptionCmd := &cobra.Command{
		Use:   "set-approval-option",
		Short: "Set workflow approval option",
		Long:  "Set the approval option for a workflow (auto-approve all steps or prompt for each)",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			wfID, err := getWorkflowID(svc)
			if err != nil {
				return err
			}
			return svc.WorkflowSetApprovalOption(cmd.Context(), wfID, approveAll, promptApproval, PrintJSON)
		}),
	}
	setApprovalOptionCmd.Flags().StringVarP(&workflowID, "workflow-id", "w", "", "The ID of the workflow (uses selected workflow if not provided)")
	setApprovalOptionCmd.Flags().BoolVar(&approveAll, "approve-all", false, "Auto-approve all steps in the workflow")
	setApprovalOptionCmd.Flags().BoolVar(&promptApproval, "prompt", false, "Prompt for approval on each step")
	setApprovalOptionCmd.MarkFlagsMutuallyExclusive("approve-all", "prompt")
	workflowsCmd.AddCommand(setApprovalOptionCmd)

	runnerCmd := &cobra.Command{
		Use:   "runner",
		Short: "Manage install runner",
		Long:  "Manage the runner process for an install",
	}
	installsCmds.AddCommand(runnerCmd)

	runnerGetCmd := &cobra.Command{
		Use:         "get",
		Short:       "Get install runner info",
		Long:        "Get runner information for an install",
		Annotations: tuiAnnotation(TUIContextual),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.RunnerGet(cmd.Context(), id, PrintJSON)
		}),
	}
	runnerGetCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	runnerCmd.AddCommand(runnerGetCmd)

	runnerRestartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the install runner",
		Long:  "Restart the runner process for an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.RunnerRestart(cmd.Context(), id)
		}),
	}
	runnerRestartCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	runnerRestartCmd.MarkFlagRequired("install-id")
	runnerCmd.AddCommand(runnerRestartCmd)

	runnerShutdownVMCmd := &cobra.Command{
		Use:   "shutdown-vm",
		Short: "Shut down the install runner VM",
		Long:  "Shut down the VM running the install runner",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.RunnerVMShutDown(cmd.Context(), id)
		}),
	}
	runnerShutdownVMCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	runnerShutdownVMCmd.MarkFlagRequired("install-id")
	runnerCmd.AddCommand(runnerShutdownVMCmd)

	runnerShutdownCmd := &cobra.Command{
		Use:    "shutdown",
		Short:  "Shut down the install runner mng process",
		Long:   "Shut down the mng process for an install runner (does not shut down the runner process)",
		Hidden: true,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.RunnerShutDown(cmd.Context(), id)
		}),
	}
	runnerShutdownCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	runnerShutdownCmd.MarkFlagRequired("install-id")
	runnerCmd.AddCommand(runnerShutdownCmd)

	// NOTE(fd): this may not be the place where this ends up living
	actionsCmd := &cobra.Command{
		Use:         "actions",
		Short:       "View actions",
		Long:        "View actions by install ID",
		Annotations: tuiAnnotation(TUIAltScreen),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Actions(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	actionsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	actionsCmd.MarkFlagRequired("install-id")
	actionsCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	actionsCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum actions to return")
	installsCmds.AddCommand(actionsCmd)

	return installsCmds
}

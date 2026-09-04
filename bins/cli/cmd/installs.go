package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/installs"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (c *cli) installsCmd() *cobra.Command {
	var (
		id                  string
		workflowID          string
		actionWorkflowID    string
		stepID              string
		note                string
		name                string
		region              string
		awsAccountID        string
		azureSubscriptionID string
		gcpProjectID        string
		appID               string
		deployID            string
		runID               string
		installCompID       string
		componentID         string
		roleName            string
		inputs              []string
		labelArgs           []string
		noSelect            bool
		deployDeps          bool
		deployDependents    bool
		deployDependencies  bool
		stackOnly           bool
		inputsOnly          bool
		offset              int
		limit               int
		planOnly            bool
		fileOrDir           string
		confirm             bool
		syncApproveAll      bool
		deprecatedYes       bool
		wait                bool
		enable              bool
		disable             bool
		dryRun              bool
		skipConfirm         bool
	)

	installsCmds := &cobra.Command{
		Use:               "installs",
		Short:             "Manage installs",
		Aliases:           []string{"i"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           InstallGroup.ID,
	}

	installOpsGroup := cobra.Group{ID: "install-ops", Title: "Install Commands"}
	installConfigGroup := cobra.Group{ID: "install-config", Title: "Install Config Commands"}
	installResourceGroup := cobra.Group{ID: "install-resources", Title: "Install Child Commands"}
	installsCmds.AddGroup(&installOpsGroup, &installConfigGroup, &installResourceGroup)

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installs",
		Long: `List all your app's installs.

Use --labels (repeatable, format key=value) to filter installs by labels. All
provided labels must match (AND semantics):

  nuon installs list --labels env=prod --labels team=platform`,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.List(cmd.Context(), appID, offset, limit, labelArgs, PrintJSON)
		}),
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter installs by")
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum installs to return")
	listCmd.Flags().StringSliceVar(&labelArgs, "labels", []string{}, "Filter installs by labels (repeatable, format: key=value). All labels must match.")
	installsCmds.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an install",
		Long:  "Get an install by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.Get(cmd.Context(), id, PrintJSON)
		}),
	}
	getCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	getCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(getCmd)

	currentCmd := &cobra.Command{
		Use:    "current",
		Short:  "Get current install (deprecated)",
		Hidden: true,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			printDeprecatedCommandWarning(cmd, "Use `nuon installs get` instead")

			svc := c.installs
			return svc.Get(cmd.Context(), c.cfg.GetString("install_id"), PrintJSON)
		}),
	}
	installsCmds.AddCommand(currentCmd)

	generateConfigCmd := &cobra.Command{
		Use:   "generate-config",
		Short: "Generate config for an existing install",
		Long:  "Generate config file for an existing install, to be used with a nuon app config",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.GenerateConfig(cmd.Context(), id, PrintJSON)
		}),
	}
	generateConfigCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to import")
	generateConfigCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(generateConfigCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an install",
		Long: `Create a new install of your app.

--region is required for every cloud. It is the AWS or GCP region, or the Azure location.

Use --label (repeatable, format key=value) to attach labels at creation time:

  nuon installs create -a my-app -n my-install -r us-west-2 \
    --label env=prod --label team=platform

Use --stack-only to provision the stack and runner and stop there, leaving the
sandbox and components unprovisioned:

  nuon installs create -a my-app -n my-install -r us-west-2 --stack-only
  nuon installs inputs set bootstrap_token=... -i my-install --inputs-only
  nuon installs reprovision-sandbox -i my-install`,
		Annotations: tuiAnnotation(TUIAltScreen),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.Create(cmd.Context(), appID, name, region, installs.TargetAccount{
				AWSAccountID:        awsAccountID,
				AzureSubscriptionID: azureSubscriptionID,
				GCPProjectID:        gcpProjectID,
			}, inputs, labelArgs, PrintJSON, noSelect, stackOnly)
		}),
	}
	createCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app to create this install for")
	createCmd.Flags().StringVarP(&name, "name", "n", "", "The name you want to give this install")

	if !c.cfg.Preview {
		createCmd.MarkFlagRequired("name")
	}
	createCmd.Flags().StringVarP(&region, "region", "r", "", "The AWS or GCP region, or Azure location (required)")
	createCmd.Flags().StringVar(&awsAccountID, "aws-account-id", "", "The AWS account ID this install targets (required when phone home authentication is enabled for your org; immutable after creation)")
	createCmd.Flags().StringVar(&azureSubscriptionID, "azure-subscription-id", "", "The Azure subscription ID this install targets (required when phone home authentication is enabled for your org; immutable after creation)")
	createCmd.Flags().StringVar(&gcpProjectID, "gcp-project-id", "", "The GCP project ID this install targets (required when phone home authentication is enabled for your org; immutable after creation)")
	createCmd.Flags().StringSliceVar(&inputs, "inputs", []string{}, "The app input values for the install")
	createCmd.Flags().StringSliceVar(&labelArgs, "label", []string{}, "Labels to set on the install (repeatable, format: key=value). Example: --label env=prod --label team=platform")
	createCmd.Flags().BoolVar(&noSelect, "no-select", false, "Do not automatically set the created install as the current install")
	createCmd.Flags().BoolVar(&stackOnly, "stack-only", false, "Provision the install stack and runner only, stopping before the sandbox and components")
	installsCmds.AddCommand(createCmd)

	confirmDelete := false
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete install",
		Long:  "Delete an install by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
			if deprecatedYes {
				confirm = true
				syncApproveAll = true
			}
			return svc.Sync(cmd.Context(), fileOrDir, appID, confirm, syncApproveAll, wait, dryRun, PrintJSON)
		}),
	}
	syncCmd.Flags().StringVarP(&fileOrDir, "file", "d", "", "Path to an install config file or a directory with install config files to sync")
	syncCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the install belongs to")
	syncCmd.Flags().BoolVar(&confirm, "confirm", false, "Set to skip the diff confirmation prompt for synced installs")
	syncCmd.Flags().BoolVar(&syncApproveAll, "approve-all", false, "Set to approve all steps in the workflows triggered by the sync, overriding each install's configured approval_option")
	syncCmd.Flags().BoolVarP(&deprecatedYes, "yes", "y", false, "Set to automatically approve diffs and workflows for synced installs")
	syncCmd.Flags().MarkDeprecated("yes", "use --confirm to skip the diff prompt and --approve-all to approve triggered workflows")
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
			svc := c.installs
			return svc.ToggleSync(cmd.Context(), id, enable, disable, PrintJSON)
		}),
	}
	toggleSyncCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to toggle config file syncing for")
	toggleSyncCmd.Flags().BoolVar(&enable, "enable", false, "Set to explicitly enable config file syncing for an install")
	toggleSyncCmd.Flags().BoolVar(&disable, "disable", false, "Set to explicitly disable config file syncing for an install")
	toggleSyncCmd.MarkFlagRequired("install-id")
	toggleSyncCmd.MarkFlagsMutuallyExclusive("enable", "disable")
	installsCmds.AddCommand(toggleSyncCmd)

	var healthAppID, healthLabels string
	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "Show component health across installs",
		Long: "Show the component health rollup for every install, optionally narrowed by app or install labels.\n\n" +
			"Intended for gating a rollout: poll this with --output agent and continue only once all_healthy is true.",
		Args: cobra.NoArgs,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.Health(cmd.Context(), healthAppID, healthLabels, PrintJSON)
		}),
	}
	healthCmd.Flags().StringVar(&healthAppID, "app-id", "", "Only include installs of this app")
	healthCmd.Flags().StringVar(&healthLabels, "labels", "", "Only include installs matching these labels (key:value,key:value)")
	installsCmds.AddCommand(healthCmd)

	var compOffset, compLimit int
	componentsCmd := &cobra.Command{
		Use:   "components",
		Short: "Manage install components",
		Long:  "Manage the components on an install. With no subcommand, lists the components on an install.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.Components(cmd.Context(), id, compOffset, compLimit, false, PrintJSON)
		}),
	}
	componentsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	componentsCmd.MarkFlagRequired("install-id")
	componentsCmd.Flags().IntVarP(&compOffset, "offset", "o", 0, "Offset for pagination")
	componentsCmd.Flags().IntVarP(&compLimit, "limit", "l", 0, "Maximum components to return (0 returns all)")

	componentsOutputsCmd := &cobra.Command{
		Use:         "outputs",
		Short:       "Get component outputs",
		Long:        "Fetch the latest outputs for an install component",
		Args:        cobra.NoArgs,
		Annotations: previewAnnotation(),
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if !c.cfg.Preview {
				return fmt.Errorf("[NUON_PREVIEW=false] installs components outputs is a preview feature, set NUON_PREVIEW=true to enable")
			}
			return nil
		},
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if componentID == "" {
				ui.PrintWarning("missing --component-id; pass a component ID to fetch outputs")
				return nil
			}

			svc := c.installs
			return svc.ComponentOutputs(cmd.Context(), id, componentID, PrintJSON)
		}),
	}
	componentsOutputsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	componentsOutputsCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The component ID on the install")
	componentsOutputsCmd.MarkFlagRequired("install-id")
	componentsCmd.AddCommand(componentsOutputsCmd)

	componentsToggleCmd := &cobra.Command{
		Use:   "toggle",
		Short: "Toggle a component on an install",
		Long:  "Enable or disable a toggleable component on an install. If neither --enable nor --disable is passed, an interactive prompt will ask. The resulting workflow is shown in the TUI unless -j is passed.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if enable && disable {
				return fmt.Errorf("only one of --enable or --disable can be set")
			}
			svc := c.installs
			return svc.ToggleComponent(cmd.Context(), id, componentID, enable, disable, planOnly, PrintJSON)
		}),
	}
	componentsToggleCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	componentsToggleCmd.MarkFlagRequired("install-id")
	componentsToggleCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The component ID or name to toggle")
	componentsToggleCmd.MarkFlagRequired("component-id")
	componentsToggleCmd.Flags().BoolVar(&enable, "enable", false, "Enable the component")
	componentsToggleCmd.Flags().BoolVar(&disable, "disable", false, "Disable the component")
	componentsToggleCmd.Flags().BoolVar(&planOnly, "plan-only", false, "Only plan the resulting deploy or teardown, do not apply it")
	componentsCmd.AddCommand(componentsToggleCmd)

	var compShowAll bool
	componentsListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List install components",
		Long:    "List all components on an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.Components(cmd.Context(), id, compOffset, compLimit, compShowAll, PrintJSON)
		}),
	}
	componentsListCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	componentsListCmd.MarkFlagRequired("install-id")
	componentsListCmd.Flags().IntVarP(&compOffset, "offset", "o", 0, "Offset for pagination")
	componentsListCmd.Flags().IntVarP(&compLimit, "limit", "l", 0, "Maximum components to return (0 returns all)")
	componentsListCmd.Flags().BoolVar(&compShowAll, "all", false, "Show all components including those removed from the current config")
	componentsCmd.AddCommand(componentsListCmd)

	componentsDeploysCmd := &cobra.Command{
		Use:     "deploys",
		Aliases: []string{"ls"},
		Short:   "List deploys for an install component",
		Long:    "List the deploys for a single component on an install (alias for `nuon installs deploys list`)",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.ComponentDeploysList(cmd.Context(), id, componentID, offset, limit, PrintJSON)
		}),
	}
	componentsDeploysCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	componentsDeploysCmd.MarkFlagRequired("install-id")
	componentsDeploysCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The component ID or name")
	componentsDeploysCmd.MarkFlagRequired("component-id")
	componentsDeploysCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	componentsDeploysCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum deploys to return")
	componentsCmd.AddCommand(componentsDeploysCmd)

	componentsDeployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy an install component",
		Long:  "Deploy a single component on an install (alias for `nuon installs deploys create --type=deploy`). Uses the component's latest build unless --build-id is given.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.ComponentDeployCreate(cmd.Context(), id, componentID, deployID, deployDeps, deployDependencies, PrintJSON)
		}),
	}
	componentsDeployCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	componentsDeployCmd.MarkFlagRequired("install-id")
	componentsDeployCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The component ID or name to deploy")
	componentsDeployCmd.MarkFlagRequired("component-id")
	componentsDeployCmd.Flags().StringVarP(&deployID, "build-id", "b", "", "The build ID to deploy (defaults to the component's latest build)")
	componentsDeployCmd.Flags().BoolVar(&deployDeps, "dependents", false, "Trigger a deploy for any component that depends on this component")
	componentsDeployCmd.Flags().BoolVar(&deployDependencies, "dependency-images", false, "Sync any images that this component depends on")
	componentsCmd.AddCommand(componentsDeployCmd)

	componentsDeployAllCmd := &cobra.Command{
		Use:   "deploy-all",
		Short: "Deploy all components to an install",
		Long:  "Deploy all components to an install.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.DeployComponents(cmd.Context(), id, roleName, planOnly, PrintJSON)
		}),
	}
	componentsDeployAllCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	componentsDeployAllCmd.MarkFlagRequired("install-id")
	componentsDeployAllCmd.Flags().BoolVar(&planOnly, "plan-only", false, "Only plan, do not actually deploy")
	componentsDeployAllCmd.Flags().StringVar(&roleName, "role-name", "", "IAM role name to use for component deploys")
	componentsCmd.AddCommand(componentsDeployAllCmd)

	componentsTeardownCmd := &cobra.Command{
		Use:   "teardown",
		Short: "Teardown a component on an install",
		Long:  "Teardown a deployed component on an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.TeardownComponent(cmd.Context(), id, componentID, roleName, PrintJSON)
		}),
	}
	componentsTeardownCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	componentsTeardownCmd.MarkFlagRequired("install-id")
	componentsTeardownCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The ID of the component you want to teardown")
	componentsTeardownCmd.MarkFlagRequired("component-id")
	componentsTeardownCmd.Flags().StringVar(&roleName, "role-name", "", "IAM role name to use for component teardown")
	componentsCmd.AddCommand(componentsTeardownCmd)

	var recoverAutoApprove bool
	componentsRecoverHelmReleaseCmd := &cobra.Command{
		Use:   "recover-helm-release",
		Short: "Recover a stuck helm release for a component on an install",
		Long: "Recover a Helm release that was left part-way through an operation, so the component can be deployed again.\n\n" +
			"Helm marks a release pending before it starts changing the cluster and clears that when the operation finishes. " +
			"A release left pending is a rollout whose runner went away, and Helm refuses every further operation on it until " +
			"it is recovered.\n\n" +
			"If an earlier revision rolled out successfully, the release is rolled back to it. If none ever did, the stuck " +
			"release is removed so the next deploy can start clean. Nothing is deployed either way — deploy the component " +
			"afterwards to roll out the version you want.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.RecoverHelmRelease(cmd.Context(), id, componentID, roleName, recoverAutoApprove, PrintJSON)
		}),
	}
	componentsRecoverHelmReleaseCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	componentsRecoverHelmReleaseCmd.MarkFlagRequired("install-id")
	componentsRecoverHelmReleaseCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The ID of the component whose helm release you want to recover")
	componentsRecoverHelmReleaseCmd.MarkFlagRequired("component-id")
	componentsRecoverHelmReleaseCmd.Flags().StringVar(&roleName, "role-name", "", "IAM role name to use for the recovery")
	componentsRecoverHelmReleaseCmd.Flags().BoolVarP(&recoverAutoApprove, "yes", "y", false, "Skip the confirmation prompt")
	componentsCmd.AddCommand(componentsRecoverHelmReleaseCmd)

	componentsTeardownAllCmd := &cobra.Command{
		Use:   "teardown-all",
		Short: "Teardown all components on an install",
		Long:  "Teardown all deployed components on an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.TeardownComponents(cmd.Context(), id, PrintJSON)
		}),
	}
	componentsTeardownAllCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	componentsTeardownAllCmd.MarkFlagRequired("install-id")
	componentsCmd.AddCommand(componentsTeardownAllCmd)

	componentsForgetCmd := &cobra.Command{
		Use:   "forget",
		Short: "Forget a component on an install",
		Long:  "Remove a component from Nuon's tracking without destroying its underlying infrastructure. The component must first be removed from the app config (via nuon apps sync). This is irreversible via the API.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.ForgetComponent(cmd.Context(), id, componentID, skipConfirm, PrintJSON)
		}),
	}
	componentsForgetCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	componentsForgetCmd.MarkFlagRequired("install-id")
	componentsForgetCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The component ID or name to forget")
	componentsForgetCmd.MarkFlagRequired("component-id")
	componentsForgetCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip the confirmation prompt")
	componentsCmd.AddCommand(componentsForgetCmd)

	installsCmds.AddCommand(componentsCmd)

	var deployType string
	deploysCmd := &cobra.Command{
		Use:   "deploys",
		Short: "Manage install component deploys",
		Long:  "Manage the deploys for components on an install",
	}

	deploysListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List deploys for an install component",
		Long:    "List the deploys for a single component on an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.ComponentDeploysList(cmd.Context(), id, componentID, offset, limit, PrintJSON)
		}),
	}
	deploysListCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	deploysListCmd.MarkFlagRequired("install-id")
	deploysListCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The component ID or name")
	deploysListCmd.MarkFlagRequired("component-id")
	deploysListCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	deploysListCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum deploys to return")
	deploysCmd.AddCommand(deploysListCmd)

	deploysGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an install deploy",
		Long:  "Get an install deploy by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.GetDeploy(cmd.Context(), id, deployID, PrintJSON)
		}),
	}
	deploysGetCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	deploysGetCmd.MarkFlagRequired("install-id")
	deploysGetCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The component ID or name")
	deploysGetCmd.Flags().StringVarP(&deployID, "deploy-id", "d", "", "The deploy ID")
	deploysGetCmd.MarkFlagRequired("deploy-id")
	deploysCmd.AddCommand(deploysGetCmd)

	deploysCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a deploy for an install component",
		Long: `Create a deploy for a single component on an install.

--type selects the operation:
  --type=deploy     deploy the component (uses its latest build unless --build-id is given)
  --type=teardown   teardown the component`,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			switch deployType {
			case "deploy":
				return svc.ComponentDeployCreate(cmd.Context(), id, componentID, deployID, deployDeps, deployDependencies, PrintJSON)
			case "teardown":
				return svc.TeardownComponent(cmd.Context(), id, componentID, roleName, PrintJSON)
			default:
				return fmt.Errorf("invalid --type %q: must be one of deploy, teardown", deployType)
			}
		}),
	}
	deploysCreateCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	deploysCreateCmd.MarkFlagRequired("install-id")
	deploysCreateCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The component ID or name")
	deploysCreateCmd.MarkFlagRequired("component-id")
	deploysCreateCmd.Flags().StringVar(&deployType, "type", "", "The deploy type: deploy or teardown")
	deploysCreateCmd.MarkFlagRequired("type")
	deploysCreateCmd.Flags().StringVarP(&deployID, "build-id", "b", "", "The build ID to deploy (defaults to the component's latest build; --type=deploy only)")
	deploysCreateCmd.Flags().BoolVar(&deployDeps, "dependents", false, "Trigger a deploy for any component that depends on this component (--type=deploy only)")
	deploysCreateCmd.Flags().BoolVar(&deployDependencies, "dependency-images", false, "Sync any images that this component depends on (--type=deploy only)")
	deploysCreateCmd.Flags().StringVar(&roleName, "role-name", "", "IAM role name to use (--type=teardown only)")
	deploysCmd.AddCommand(deploysCreateCmd)

	deploysCancelCmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel an install deploy",
		Long:  "Cancel an in-flight install deploy by cancelling its workflow",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.DeployCancel(cmd.Context(), id, deployID, PrintJSON)
		}),
	}
	deploysCancelCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	deploysCancelCmd.MarkFlagRequired("install-id")
	deploysCancelCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The component ID or name")
	deploysCancelCmd.Flags().StringVarP(&deployID, "deploy-id", "d", "", "The deploy ID to cancel")
	deploysCancelCmd.MarkFlagRequired("deploy-id")
	deploysCmd.AddCommand(deploysCancelCmd)

	deploysLogsCmd := &cobra.Command{
		Use:         "logs",
		Short:       "View deploy logs",
		Long:        "View deploy logs by install and deploy ID",
		Annotations: tuiAnnotation(TUIAltScreen),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.DeployLogs(cmd.Context(), id, deployID, installCompID, PrintJSON)
		}),
	}
	deploysLogsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	deploysLogsCmd.MarkFlagRequired("install-id")
	deploysLogsCmd.Flags().StringVarP(&deployID, "deploy-id", "d", "", "The deploy ID for the deploy log you want to view")
	deploysLogsCmd.MarkFlagRequired("deploy-id")
	deploysCmd.AddCommand(deploysLogsCmd)

	installsCmds.AddCommand(deploysCmd)

	getDeployCmd := &cobra.Command{
		Use:        "get-deploy",
		Deprecated: "use `nuon installs deploys get` instead",
		Hidden:     true,
		Short:      "Get an install deploy",
		Long:       "Get an install deploy by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.GetDeploy(cmd.Context(), id, deployID, PrintJSON)
		}),
	}
	getDeployCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	getDeployCmd.Flags().StringVarP(&deployID, "deploy-id", "d", "", "The deploy ID for the deploy log you want to view")
	getDeployCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(getDeployCmd)

	createDeployCmd := &cobra.Command{
		Use:        "create-deploy",
		Deprecated: "use `nuon installs components deploy` instead",
		Hidden:     true,
		Short:      "Create an install deploy",
		Long:       "Create an install deploy by install ID and build ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.CreateDeploy(cmd.Context(), id, deployID, deployDeps, deployDependencies, PrintJSON)
		}),
	}
	createDeployCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	createDeployCmd.MarkFlagRequired("install-id")
	createDeployCmd.Flags().StringVarP(&deployID, "build-id", "b", "", "The build ID for the deploy you want to create")
	createDeployCmd.MarkFlagRequired("build-id")
	createDeployCmd.Flags().BoolVar((&deployDeps), "dependents", false, "Trigger a deploy for any component that depends on this component")
	createDeployCmd.Flags().BoolVar((&deployDependencies), "dependency-images", false, "Sync any images that this component depends on")
	installsCmds.AddCommand(createDeployCmd)

	deployLogsCmd := &cobra.Command{
		Use:         "deploy-logs",
		Deprecated:  "use `nuon installs deploys logs` instead",
		Hidden:      true,
		Short:       "View deploy logs",
		Long:        "View deploy logs by install and deploy ID",
		Annotations: tuiAnnotation(TUIAltScreen),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.DeployLogs(cmd.Context(), id, deployID, installCompID, PrintJSON)
		}),
	}
	deployLogsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install whose deploy you want to view")
	deployLogsCmd.MarkFlagRequired("install-id")
	deployLogsCmd.Flags().StringVarP(&deployID, "deploy-id", "d", "", "The deploy ID for the deploy log you want to view")
	deployLogsCmd.MarkFlagRequired("deploy-id")
	installsCmds.AddCommand(deployLogsCmd)

	listDeploysCmd := &cobra.Command{
		Use:        "list-deploys",
		Deprecated: "use `nuon installs deploys list` instead",
		Hidden:     true,
		Short:      "View all install deploys",
		Long:       "View all install deploys by install ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.ListDeploys(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	listDeploysCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install whose deploy you want to view")
	listDeploysCmd.MarkFlagRequired("install-id")
	listDeploysCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listDeploysCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum deploys to return")
	installsCmds.AddCommand(listDeploysCmd)

	var (
		sandboxRunType        string
		sandboxSkipComponents bool
	)
	sandboxRunsCmd := &cobra.Command{
		Use:   "sandbox-runs",
		Short: "Manage install sandbox runs",
		Long:  "Manage the sandbox runs for an install",
	}

	sandboxRunsListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List install sandbox runs",
		Long:    "List the sandbox runs for an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.SandboxRuns(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	sandboxRunsListCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	sandboxRunsListCmd.MarkFlagRequired("install-id")
	sandboxRunsListCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	sandboxRunsListCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum runs to return")
	sandboxRunsCmd.AddCommand(sandboxRunsListCmd)

	sandboxRunsGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an install sandbox run",
		Long:  "Get an install sandbox run by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.SandboxRunGet(cmd.Context(), id, runID, PrintJSON)
		}),
	}
	sandboxRunsGetCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	sandboxRunsGetCmd.MarkFlagRequired("install-id")
	sandboxRunsGetCmd.Flags().StringVarP(&runID, "run-id", "r", "", "The ID of the sandbox run")
	sandboxRunsGetCmd.MarkFlagRequired("run-id")
	sandboxRunsCmd.AddCommand(sandboxRunsGetCmd)

	sandboxRunsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an install sandbox run",
		Long: `Create a sandbox run for an install.

--type selects the operation:
  --type=reprovision   reprovision the install sandbox
  --type=deprovision   deprovision the install sandbox`,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			switch sandboxRunType {
			case "reprovision":
				return svc.ReprovisionSandbox(cmd.Context(), id, sandboxSkipComponents, PrintJSON)
			case "deprovision":
				return svc.DeprovisionSandbox(cmd.Context(), id, PrintJSON)
			default:
				return fmt.Errorf("invalid --type %q: must be one of reprovision, deprovision", sandboxRunType)
			}
		}),
	}
	sandboxRunsCreateCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	sandboxRunsCreateCmd.MarkFlagRequired("install-id")
	sandboxRunsCreateCmd.Flags().StringVar(&sandboxRunType, "type", "", "The run type: reprovision or deprovision")
	sandboxRunsCreateCmd.MarkFlagRequired("type")
	sandboxRunsCreateCmd.Flags().BoolVar(&sandboxSkipComponents, "skip-components", false, "Skip deploying components after reprovisioning (--type=reprovision only)")
	sandboxRunsCmd.AddCommand(sandboxRunsCreateCmd)

	sandboxRunsCancelCmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel an install sandbox run",
		Long:  "Cancel an in-flight install sandbox run by cancelling its workflow",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.SandboxRunCancel(cmd.Context(), id, runID, PrintJSON)
		}),
	}
	sandboxRunsCancelCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	sandboxRunsCancelCmd.MarkFlagRequired("install-id")
	sandboxRunsCancelCmd.Flags().StringVarP(&runID, "run-id", "r", "", "The ID of the sandbox run to cancel")
	sandboxRunsCancelCmd.MarkFlagRequired("run-id")
	sandboxRunsCmd.AddCommand(sandboxRunsCancelCmd)

	sandboxRunsLogsCmd := &cobra.Command{
		Use:   "logs",
		Short: "View sandbox run logs",
		Long:  "View sandbox run logs by run & install IDs",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.SandboxRunLogs(cmd.Context(), id, runID, PrintJSON)
		}),
	}
	sandboxRunsLogsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	sandboxRunsLogsCmd.MarkFlagRequired("install-id")
	sandboxRunsLogsCmd.Flags().StringVarP(&runID, "run-id", "r", "", "The ID of the run you want to view")
	sandboxRunsLogsCmd.MarkFlagRequired("run-id")
	sandboxRunsLogsCmd.Flags().StringVarP(&installCompID, "install-comp-id", "c", "", "The ID of the install component to view logs for")
	sandboxRunsCmd.AddCommand(sandboxRunsLogsCmd)

	installsCmds.AddCommand(sandboxRunsCmd)

	sandboxCmd := &cobra.Command{
		Use:   "sandbox",
		Short: "Manage an install sandbox",
		Long:  "Manage an install sandbox and its runs",
	}

	sandboxRunsPorcelainCmd := &cobra.Command{
		Use:     "runs",
		Aliases: []string{"ls"},
		Short:   "List install sandbox runs",
		Long:    "List the sandbox runs for an install (alias for `nuon installs sandbox-runs list`)",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.SandboxRuns(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	sandboxRunsPorcelainCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	sandboxRunsPorcelainCmd.MarkFlagRequired("install-id")
	sandboxRunsPorcelainCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	sandboxRunsPorcelainCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum runs to return")
	sandboxCmd.AddCommand(sandboxRunsPorcelainCmd)

	sandboxReprovisionCmd := &cobra.Command{
		Use:   "reprovision",
		Short: "Reprovision an install sandbox",
		Long:  "Reprovision an install sandbox (alias for `nuon installs sandbox-runs create --type=reprovision`)",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.ReprovisionSandbox(cmd.Context(), id, sandboxSkipComponents, PrintJSON)
		}),
	}
	sandboxReprovisionCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	sandboxReprovisionCmd.MarkFlagRequired("install-id")
	sandboxReprovisionCmd.Flags().BoolVar(&sandboxSkipComponents, "skip-components", false, "Skip deploying components after reprovisioning the sandbox")
	sandboxCmd.AddCommand(sandboxReprovisionCmd)

	sandboxDeprovisionCmd := &cobra.Command{
		Use:   "deprovision",
		Short: "Deprovision an install sandbox",
		Long:  "Deprovision an install sandbox (alias for `nuon installs sandbox-runs create --type=deprovision`)",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.DeprovisionSandbox(cmd.Context(), id, PrintJSON)
		}),
	}
	sandboxDeprovisionCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	sandboxDeprovisionCmd.MarkFlagRequired("install-id")
	sandboxCmd.AddCommand(sandboxDeprovisionCmd)

	installsCmds.AddCommand(sandboxCmd)

	sandboxRunLogsCmd := &cobra.Command{
		Use:        "sandbox-run-logs",
		Deprecated: "use `nuon installs sandbox-runs logs` instead",
		Hidden:     true,
		Short:      "View sandbox run logs",
		Long:       "View sandbox run logs by run & install IDs",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.SandboxRunLogs(cmd.Context(), id, runID, PrintJSON)
		}),
	}
	sandboxRunLogsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	sandboxRunLogsCmd.MarkFlagRequired("install-id")
	sandboxRunLogsCmd.Flags().StringVarP(&runID, "run-id", "r", "", "The ID of the run you want to view")
	sandboxRunLogsCmd.MarkFlagRequired("run-id")
	sandboxRunLogsCmd.Flags().StringVarP(&installCompID, "install-comp-id", "c", "", "The ID of the install component to view logs for")
	installsCmds.AddCommand(sandboxRunLogsCmd)

	sandboxOutputsCmd := &cobra.Command{
		Use:        "sandbox-outputs",
		Deprecated: "Use `nuon installs outputs --sandbox` instead",
		Hidden:     true,
		Short:      "View sandbox outputs (deprecated)",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.Outputs(cmd.Context(), id, installs.OutputsOptions{SandboxOnly: true}, PrintJSON)
		}),
	}
	sandboxOutputsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	sandboxOutputsCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(sandboxOutputsCmd)

	var (
		outputsStack   bool
		outputsSandbox bool
	)
	outputsCmd := &cobra.Command{
		Use:   "outputs",
		Short: "View install outputs",
		Long: `View install outputs across stack, sandbox, and components.

By default, all sections are shown. Use a filter flag to scope the output to
a single section:

  --stack                       show only the install stack outputs
  --sandbox                     show only sandbox outputs
  --component-id <id-or-name>   show outputs for a single component

The --stack, --sandbox, and --component-id flags are mutually exclusive.`,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.Outputs(cmd.Context(), id, installs.OutputsOptions{
				StackOnly:   outputsStack,
				SandboxOnly: outputsSandbox,
				ComponentID: componentID,
			}, PrintJSON)
		}),
	}
	outputsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	outputsCmd.MarkFlagRequired("install-id")
	outputsCmd.Flags().BoolVar(&outputsStack, "stack", false, "Show only the install stack outputs")
	outputsCmd.Flags().BoolVar(&outputsSandbox, "sandbox", false, "Show only sandbox outputs")
	outputsCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The ID or name of a component to show outputs for")
	outputsCmd.MarkFlagsMutuallyExclusive("stack", "sandbox", "component-id")
	installsCmds.AddCommand(outputsCmd)

	currentInputs := &cobra.Command{
		Use:        "current-inputs",
		Deprecated: "use `nuon installs inputs get` instead",
		Hidden:     true,
		Short:      "View current inputs",
		Long:       "View current set app inputs",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.CurrentInputs(cmd.Context(), id, PrintJSON)
		}),
	}
	currentInputs.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	currentInputs.MarkFlagRequired("install-id")
	installsCmds.AddCommand(currentInputs)

	inputsCmd := &cobra.Command{
		Use:   "inputs",
		Short: "Manage install inputs",
		Long:  "View and set the app input values for an install",
	}
	inputsCmd.PersistentFlags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")

	inputsGetCmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"list"},
		Short:   "View install inputs",
		Long:    "View the install's current input values alongside their declared defaults",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.GetInputs(cmd.Context(), id, PrintJSON)
		}),
	}
	inputsCmd.AddCommand(inputsGetCmd)

	inputsSetCmd := &cobra.Command{
		Use:   "set [key=value ...]",
		Short: "Set install inputs",
		Long: `Set one or more install input values.

Accepts a list of inputs in key=value format, e.g.:

  nuon installs inputs set foo=bar baz=qux

The current inputs are fetched first so changed values can be shown. Setting an
input that is not declared on the app raises an error.

Use --inputs-only to save the values without deploying components or
reprovisioning the sandbox.`,
		Args: cobra.MinimumNArgs(1),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := c.installs
			return svc.SetInputs(cmd.Context(), id, args, deployDependents, inputsOnly, PrintJSON)
		}),
	}
	inputsSetCmd.Flags().BoolVar(&deployDependents, "deploy-dependents", true, "Deploy components that depend on the updated inputs")
	inputsSetCmd.Flags().BoolVar(&inputsOnly, "inputs-only", false, "Record the new input values without deploying components or reprovisioning the sandbox")
	inputsCmd.AddCommand(inputsSetCmd)

	// `inputs edit` is a preview feature, gated behind NUON_PREVIEW.
	if c.cfg.Preview {
		inputsEditCmd := &cobra.Command{
			Use:         "edit",
			Short:       "Edit install inputs",
			Long:        "Edit an install's inputs in an interactive TUI form pre-filled with the current values",
			Annotations: annotations(tuiAnnotation(TUIAltScreen), outputsAnnotation(OutputTable)),
			Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
				svc := c.installs
				return svc.EditInputs(cmd.Context(), id, deployDependents)
			}),
		}
		inputsEditCmd.Flags().BoolVar(&deployDependents, "deploy-dependents", true, "Deploy components that depend on the updated inputs")
		inputsCmd.AddCommand(inputsEditCmd)
	}

	installsCmds.AddCommand(inputsCmd)

	selectInstallCmd := &cobra.Command{
		Use:         "select",
		Short:       "Select your current install",
		Long:        "Select your current install from a list or by install ID",
		Annotations: tuiAnnotation(TUIContextual),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
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
			svc := c.installs
			return svc.Deselect(cmd.Context(), PrintJSON)
		}),
	}
	installsCmds.AddCommand(deselectInstallCmd)

	unsetCurrentInstallCmd := &cobra.Command{
		Use:        "unset-current",
		Deprecated: "Use `nuon installs deselect` instead",
		Short:      "Unset your current install selection (deprecated)",
		Hidden:     true,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.Deselect(cmd.Context(), PrintJSON)
		}),
	}
	installsCmds.AddCommand(unsetCurrentInstallCmd)

	var reprovisionStackOnly, reprovisionSkipComponents bool
	reprovisionInstallCmd := &cobra.Command{
		Use:   "reprovision",
		Short: "Reprovision install",
		Long: `Reprovision an install: the stack, then the sandbox, then all components.

With --stack-only, only the stack is reprovisioned — the runner and its
infrastructure are recreated and the sandbox is left alone. This is the same as
` + "`nuon installs stacks reprovision`" + `.`,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			if reprovisionSkipComponents && !reprovisionStackOnly {
				return ui.PrintError(&ui.CLIUserError{Msg: "--skip-components is only supported with --stack-only"})
			}
			if reprovisionStackOnly {
				return svc.ReprovisionStack(cmd.Context(), id, reprovisionSkipComponents, PrintJSON)
			}
			return svc.Reprovision(cmd.Context(), id, PrintJSON)
		}),
	}
	reprovisionInstallCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	reprovisionInstallCmd.MarkFlagRequired("install-id")
	reprovisionInstallCmd.Flags().BoolVar(&reprovisionStackOnly, "stack-only", false, "Only reprovision the install stack, leaving the sandbox untouched")
	reprovisionInstallCmd.Flags().BoolVar(&reprovisionSkipComponents, "skip-components", false, "Skip deploying components after reprovisioning the stack (--stack-only only)")
	installsCmds.AddCommand(reprovisionInstallCmd)

	deprovisionInstallCmd := &cobra.Command{
		Use:   "deprovision",
		Short: "Deprovision install",
		Long:  "Deprovision an install sandbox",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.Deprovision(cmd.Context(), id, PrintJSON)
		}),
	}
	deprovisionInstallCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	deprovisionInstallCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(deprovisionInstallCmd)

	teardownInstallComponentsCmd := &cobra.Command{
		Use:        "teardown-components",
		Deprecated: "use `nuon installs components teardown-all` instead",
		Hidden:     true,
		Short:      "Teardown components on install.",
		Long:       "Teardown all deployed components on an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.TeardownComponents(cmd.Context(), id, PrintJSON)
		}),
	}
	teardownInstallComponentsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	teardownInstallComponentsCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(teardownInstallComponentsCmd)

	teardownInstallComponentCmd := &cobra.Command{
		Use:        "teardown-component",
		Deprecated: "use `nuon installs components teardown` instead",
		Hidden:     true,
		Short:      "Teardown component on install.",
		Long:       "Teardown a deployed component on an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.TeardownComponent(cmd.Context(), id, componentID, roleName, PrintJSON)
		}),
	}
	teardownInstallComponentCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	teardownInstallComponentCmd.MarkFlagRequired("install-id")
	teardownInstallComponentCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The ID of the component you want to teardown")
	teardownInstallComponentCmd.MarkFlagRequired("component-id")
	teardownInstallComponentCmd.Flags().StringVar(&roleName, "role-name", "", "IAM role name to use for component teardown")
	installsCmds.AddCommand(teardownInstallComponentCmd)

	forgetInstallComponentCmd := &cobra.Command{
		Use:        "forget-component",
		Deprecated: "use `nuon installs components forget` instead",
		Hidden:     true,
		Short:      "Forget a component on an install.",
		Long:       "Remove a component from Nuon's tracking without destroying its underlying infrastructure. The component must first be removed from the app config (via nuon apps sync). This is irreversible via the API.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.ForgetComponent(cmd.Context(), id, componentID, skipConfirm, PrintJSON)
		}),
	}
	forgetInstallComponentCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	forgetInstallComponentCmd.MarkFlagRequired("install-id")
	forgetInstallComponentCmd.Flags().StringVarP(&componentID, "component-id", "c", "", "The component ID or name to forget")
	forgetInstallComponentCmd.MarkFlagRequired("component-id")
	forgetInstallComponentCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip the confirmation prompt")
	installsCmds.AddCommand(forgetInstallComponentCmd)

	deployInstallComponentsCmd := &cobra.Command{
		Use:        "deploy-components",
		Deprecated: "use `nuon installs components deploy-all` instead",
		Hidden:     true,
		Short:      "Deploy all components to an install.",
		Long:       "Deploy all components to an install.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.DeployComponents(cmd.Context(), id, roleName, planOnly, PrintJSON)
		}),
	}
	deployInstallComponentsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	deployInstallComponentsCmd.MarkFlagRequired("install-id")
	deployInstallComponentsCmd.Flags().BoolVar(&planOnly, "plan-only", false, "Only plan, do not actually deploy")
	deployInstallComponentsCmd.Flags().StringVar(&roleName, "role-name", "", "IAM role name to use for component deploys")
	installsCmds.AddCommand(deployInstallComponentsCmd)

	updateInputCmd := &cobra.Command{
		Use:        "update-input",
		Deprecated: "use `nuon installs inputs set` instead",
		Hidden:     true,
		Short:      "Update install input",
		Long:       "Update an install input value",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.UpdateInput(cmd.Context(), id, inputs, deployDependents, PrintJSON)
		}),
	}
	updateInputCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to update")
	updateInputCmd.MarkFlagRequired("install-id")
	updateInputCmd.Flags().StringSliceVar(&inputs, "inputs", []string{}, "The app input values for the install")
	updateInputCmd.MarkFlagRequired("inputs")
	updateInputCmd.Flags().BoolVar(&deployDependents, "deploy-dependents", true, "Deploy components that depend on the updated inputs")
	installsCmds.AddCommand(updateInputCmd)

	deprovisionInstallSandboxCmd := &cobra.Command{
		Use:        "deprovision-sandbox",
		Deprecated: "use `nuon installs sandbox deprovision` instead",
		Hidden:     true,
		Short:      "Deprovision install sandbox",
		Long:       "Deprovision an install sandbox",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.DeprovisionSandbox(cmd.Context(), id, PrintJSON)
		}),
	}
	deprovisionInstallSandboxCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to use")
	deprovisionInstallSandboxCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(deprovisionInstallSandboxCmd)

	var skipComponents bool
	reprovisionInstallSandboxCmd := &cobra.Command{
		Use:         "reprovision-sandbox",
		Deprecated:  "use `nuon installs sandbox reprovision` instead",
		Hidden:      true,
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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
			return svc.WorkflowsDeselect(cmd.Context(), PrintJSON)
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
			svc := c.installs

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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
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
			svc := c.installs
			return svc.RunnerRestart(cmd.Context(), id, PrintJSON)
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
			svc := c.installs
			return svc.RunnerVMShutDown(cmd.Context(), id, PrintJSON)
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
			svc := c.installs
			return svc.RunnerShutDown(cmd.Context(), id, PrintJSON)
		}),
	}
	runnerShutdownCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	runnerShutdownCmd.MarkFlagRequired("install-id")
	runnerCmd.AddCommand(runnerShutdownCmd)

	stacksCmd := &cobra.Command{
		Use:   "stacks",
		Short: "Manage install stacks",
		Long:  "View install stacks and stack versions, and reprovision an install stack",
	}
	installsCmds.AddCommand(stacksCmd)

	stacksListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List install stack versions",
		Long:    "List all stack versions for an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.StacksList(cmd.Context(), id, PrintJSON)
		}),
	}
	stacksListCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	stacksListCmd.MarkFlagRequired("install-id")
	stacksCmd.AddCommand(stacksListCmd)

	var installStackID string
	stacksGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an install stack",
		Long:  "Get an install stack by stack ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.StacksGet(cmd.Context(), installStackID, PrintJSON)
		}),
	}
	stacksGetCmd.Flags().StringVar(&installStackID, "install-stack-id", "", "The ID of the install stack")
	stacksGetCmd.MarkFlagRequired("install-stack-id")
	stacksCmd.AddCommand(stacksGetCmd)

	stacksLatestCmd := &cobra.Command{
		Use:   "latest",
		Short: "Get the latest install stack version",
		Long:  "Get the latest stack version for an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.StacksLatest(cmd.Context(), id, PrintJSON)
		}),
	}
	stacksLatestCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	stacksLatestCmd.MarkFlagRequired("install-id")
	stacksCmd.AddCommand(stacksLatestCmd)

	var stackSkipComponents bool
	stacksReprovisionCmd := &cobra.Command{
		Use:   "reprovision",
		Short: "Reprovision an install stack",
		Long:  "Reprovision an install stack, recreating the runner and its infrastructure",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.ReprovisionStack(cmd.Context(), id, stackSkipComponents, PrintJSON)
		}),
	}
	stacksReprovisionCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	stacksReprovisionCmd.MarkFlagRequired("install-id")
	stacksReprovisionCmd.Flags().BoolVar(&stackSkipComponents, "skip-components", false, "Skip deploying components after reprovisioning the stack")
	stacksCmd.AddCommand(stacksReprovisionCmd)

	// NOTE(fd): this may not be the place where this ends up living
	actionsCmd := &cobra.Command{
		Use:   "actions",
		Short: "Manage install actions [preview]",
		Long: `Manage and view install actions.

By default, launches an interactive TUI to browse and execute actions.`,
		Args:        cobra.NoArgs,
		Annotations: annotations(tuiAnnotation(TUIAltScreen), previewAnnotation()),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := c.persistentPreRunE(cmd, args); err != nil {
				return err
			}
			if !c.cfg.Preview {
				return fmt.Errorf("[NUON_PREVIEW=false] installs actions is a preview feature, set NUON_PREVIEW=true to enable")
			}
			return nil
		},
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.Actions(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	actionsCmd.PersistentFlags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	actionsCmd.MarkPersistentFlagRequired("install-id")
	actionsCmd.PersistentFlags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	actionsCmd.PersistentFlags().IntVarP(&limit, "limit", "l", 20, "Maximum actions to return")
	actionsCmd.PersistentFlags().StringVar(&actionWorkflowID, "action-workflow-id", "", "The action workflow ID or slug")

	actionsListCmd := &cobra.Command{
		Use:         "list",
		Aliases:     []string{"ls"},
		Short:       "List install actions",
		Long:        "List action workflows available for an install",
		Args:        cobra.NoArgs,
		Annotations: previewAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.ActionsList(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	actionsCmd.AddCommand(actionsListCmd)

	actionsOutputsCmd := &cobra.Command{
		Use:         "outputs",
		Short:       "Get action outputs",
		Long:        "Fetch the latest outputs for an install action workflow",
		Args:        cobra.NoArgs,
		Annotations: previewAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if actionWorkflowID == "" {
				ui.PrintWarning("missing --action-workflow-id; pass an action workflow ID or slug to fetch outputs")
				return nil
			}

			svc := c.installs
			return svc.ActionOutputs(cmd.Context(), id, actionWorkflowID, PrintJSON)
		}),
	}
	actionsCmd.AddCommand(actionsOutputsCmd)

	installsCmds.AddCommand(actionsCmd)

	labelsCmd := &cobra.Command{
		Use:   "labels",
		Short: "List, set, or unset labels on an install",
		Long: `List, set, or unset labels on an install.

Labels are key-value strings used to organize and filter installs.`,
	}

	labelsListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the labels on an install",
		Long: `List the labels on an install.

Examples:
  nuon installs labels list --install-id inst_abc`,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.LabelsList(cmd.Context(), id, PrintJSON)
		}),
	}
	labelsListCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	labelsListCmd.MarkFlagRequired("install-id")
	labelsCmd.AddCommand(labelsListCmd)

	labelsSetCmd := &cobra.Command{
		Use:   "set",
		Short: "Set (add or overwrite) labels on an install",
		Long: `Set labels on an install. Pass kubectl-style positional args:
  KEY=VALUE   add or overwrite a label

Examples:
  nuon installs labels set --install-id inst_abc env=prod team=platform`,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := c.installs
			return svc.LabelsSet(cmd.Context(), id, args, PrintJSON)
		}),
	}
	labelsSetCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	labelsSetCmd.MarkFlagRequired("install-id")
	labelsCmd.AddCommand(labelsSetCmd)

	labelsUnsetCmd := &cobra.Command{
		Use:   "unset",
		Short: "Unset (remove) labels on an install",
		Long: `Unset labels on an install. Pass one or more bare keys to remove:
  KEY   remove a label

Examples:
  nuon installs labels unset --install-id inst_abc env team`,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := c.installs
			return svc.LabelsUnset(cmd.Context(), id, args, PrintJSON)
		}),
	}
	labelsUnsetCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install")
	labelsUnsetCmd.MarkFlagRequired("install-id")
	labelsCmd.AddCommand(labelsUnsetCmd)

	installsCmds.AddCommand(labelsCmd)

	resourceCommands := map[string]struct{}{
		"components":   {},
		"deploys":      {},
		"inputs":       {},
		"outputs":      {},
		"sandbox":      {},
		"sandbox-runs": {},
		"workflows":    {},
		"stacks":       {},
		"actions":      {},
		"labels":       {},
		"runner":       {},
	}
	configCommands := map[string]struct{}{
		"generate-config": {},
		"sync":            {},
		"toggle-sync":     {},
	}
	for _, sub := range installsCmds.Commands() {
		if sub.Hidden {
			continue
		}
		if _, ok := resourceCommands[sub.Name()]; ok {
			sub.GroupID = installResourceGroup.ID
		} else if _, ok := configCommands[sub.Name()]; ok {
			sub.GroupID = installConfigGroup.ID
		} else {
			sub.GroupID = installOpsGroup.ID
		}
	}

	return installsCmds
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/bins/cli/internal/services/apps"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	initui "github.com/nuonco/nuon/bins/cli/internal/ui/v3/init"
	"github.com/spf13/cobra"
)

func initRootParams(cmd *cobra.Command) apps.ConfigGenParams {
	params := apps.ConfigGenParams{}
	params.Path, _ = cmd.Flags().GetString("path")
	params.EnableDefaults, _ = cmd.Flags().GetBool("enable-defaults")
	params.EnableComments, _ = cmd.Flags().GetBool("enable-comments")
	params.Overwrite, _ = cmd.Flags().GetBool("overwrite")
	params.SkipNonRequired, _ = cmd.Flags().GetBool("skip-non-required")

	return params
}

func successMesssage(path string, configType string) {
	fmt.Printf("%s\n", bubbles.SuccessStyle.Render(fmt.Sprintf("Successfully initialized %s in %s", configType, path)))
	fmt.Printf("%s\n", bubbles.InfoStyle.Render("Next steps:"))
	fmt.Printf("  1. Review and edit the generated configuration file\n")
	fmt.Printf("  2. Run 'nuon apps sync' to sync your configuration to Nuon\n")
}

func (c *cli) initCmd() *cobra.Command {
	// nuon apps init command
	var (
		initPath           string
		initEnableDefaults bool
		initEnableComments bool
		initOverwrite      bool
		interactive        bool
		prebuildTemplate   string
		skipNonRequired    bool
	)

	// Parent init command
	initCmd := &cobra.Command{
		Use:               "init",
		Short:             "Initialize app configuration",
		Long:              "Generate app configuration files. Use subcommands to generate specific config files, or run without subcommands to generate all files.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil }, // Skip auth for local init
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			params := apps.InitParams{
				PrebuiltTemplate: prebuildTemplate,
			}
			genParams := apps.ConfigGenParams{
				Path:            initPath,
				EnableDefaults:  initEnableDefaults,
				EnableComments:  initEnableComments,
				Overwrite:       initOverwrite,
				SkipNonRequired: skipNonRequired,
			}
			sampleComponentPArams := apps.SampleComponentParams{}
			sampleActionsParams := apps.SampleActionsParams{}

			ui := initui.NewConfigInit()
			if interactive {
				_, err := ui.RunInitMenu(&params)
				if err != nil {
					return errors.Wrap(err, "unable to get init parameters")
				}

				if params.PrebuiltTemplate != "" {
					sampleComponentPArams := apps.SampleComponentParams{}
					err = ui.RunComponentsMenu(&sampleComponentPArams)
					if err != nil {
						return errors.Wrap(err, "unable to get sample components parameters")
					}
					err = ui.RunActionsMenu(&sampleActionsParams)
					if err != nil {
						return errors.Wrap(err, "unable to get sample actions parameters")
					}

				}

				err = ui.RunGeneratorConfigMenu(&genParams)
				if err != nil {
					return errors.Wrap(err, "unable to get sample actions parameters")
				}
			}

			// run config gens here
			err := svc.Init(cmd.Context(), genParams, &params)
			if err != nil {
				return errors.Wrap(err, "unable to init app config")
			}
			if sampleComponentPArams.EnableSampleComponents {
				gparams := genParams
				gparams.Overwrite = true
				err = svc.InitSampleComponents(cmd.Context(), genParams, sampleComponentPArams)
				if err != nil {
					return errors.Wrap(err, "unable to init app config")
				}
			}
			if sampleActionsParams.EnableSampleActions {
				gparams := genParams
				gparams.Overwrite = true
				err = svc.InitSampleActions(cmd.Context(), genParams, sampleActionsParams)
				if err != nil {
					return errors.Wrap(err, "unable to init app config")
				}
			}

			successMesssage(genParams.Path, "app config")
			return nil
		}),
	}

	initCmd.PersistentFlags().StringVarP(&initPath, "path", "p", "./app-config", "path to create the app config directory")
	initCmd.PersistentFlags().StringVarP(&prebuildTemplate, "prebuild-template", "t", "", "prebuild sample apps, aws-eks aws-ecs")
	initCmd.PersistentFlags().BoolVar(&initEnableDefaults, "enable-defaults", false, "include default values from schema")
	initCmd.PersistentFlags().BoolVar(&skipNonRequired, "skip-non-required", false, "skip parameters which are not required")
	initCmd.PersistentFlags().BoolVar(&initEnableComments, "enable-comments", false, "include comments with field descriptions")
	initCmd.PersistentFlags().BoolVar(&initOverwrite, "overwrite", false, "overwrite existing directory contents")
	initCmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "interactive session")

	// Helper function to create config-specific subcommands
	// createConfigCmd := func(commandName, configFileName, description string) *cobra.Command {
	// 	return &cobra.Command{
	// 		Use:               commandName,
	// 		Short:             fmt.Sprintf("Initialize %s", description),
	// 		Long:              fmt.Sprintf("Generate the %s configuration file", configFileName),
	// 		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	// 		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
	// 			svc := apps.New(c.v, c.apiClient, c.cfg)
	// 			return svc.InitConfigFile(cmd.Context(), initPath, configFileName, initEnableDefaults, initEnableComments, initOverwrite)
	// 		}),
	// 	}
	// }

	// Add subcommands for each config type
	// initCmd.AddCommand(createConfigCmd("inputs", "inputs.toml", "inputs configuration"))
	// initCmd.AddCommand(createConfigCmd("installer", "installer.toml", "installer configuration"))
	initCmd.AddCommand(c.initSandboxCmd())
	initCmd.AddCommand(c.initRunnerCmd())
	initCmd.AddCommand(c.initStackCmd())
	// initCmd.AddCommand(createConfigCmd("secrets", "secrets.toml", "secrets configuration"))
	// initCmd.AddCommand(createConfigCmd("break-glass", "break_glass.toml", "break glass configuration"))
	// initCmd.AddCommand(createConfigCmd("policies", "policies.toml", "policies configuration"))
	initCmd.AddCommand(c.initComponentCmd())

	return initCmd
}

func (c *cli) initSandboxCmd() *cobra.Command {
	var (
		// Sandbox-specific flags
		terraformVersion string
		publicRepo       string
		publicRepoDir    string
		publicRepoBranch string
		connectedRepo    string
		connectedRepoDir string
		connectedBranch  string
		driftSchedule    string
		envVars          []string
		vars             []string
		varFiles         []string
	)

	sandboxCmd := &cobra.Command{
		Use:               "sandbox",
		Short:             "Initialize sandbox configuration",
		Long:              "Generate the sandbox.toml configuration file with custom parameters",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			genParams := initRootParams(cmd)

			params := apps.SandboxParams{
				TerraformVersion: terraformVersion,
				PublicRepo:       publicRepo,
				PublicRepoDir:    publicRepoDir,
				PublicRepoBranch: publicRepoBranch,
				ConnectedRepo:    connectedRepo,
				ConnectedRepoDir: connectedRepoDir,
				ConnectedBranch:  connectedBranch,
				DriftSchedule:    driftSchedule,
				EnvVars:          parseKeyValuePairs(envVars),
				Vars:             parseKeyValuePairs(vars),
				VarFiles:         varFiles,
			}

			svc := apps.New(c.v, c.apiClient, c.cfg)
			err := svc.InitSandboxConfig(cmd.Context(), genParams, params)
			if err != nil {
				return err
			}

			successMesssage(genParams.Path, "sandbox.toml")
			return nil
		}),
	}

	// Terraform version
	sandboxCmd.Flags().StringVar(&terraformVersion, "terraform-version", "1.11.3", "Terraform version to use")

	// Public repo flags
	sandboxCmd.Flags().StringVar(&publicRepo, "public-repo", "", "Public repository (e.g., 'nuonco/aws-eks-sandbox')")
	sandboxCmd.Flags().StringVar(&publicRepoDir, "public-repo-dir", ".", "Directory within the public repository")
	sandboxCmd.Flags().StringVar(&publicRepoBranch, "public-repo-branch", "main", "Branch of the public repository")

	// Connected repo flags
	sandboxCmd.Flags().StringVar(&connectedRepo, "connected-repo", "", "Connected repository")
	sandboxCmd.Flags().StringVar(&connectedRepoDir, "connected-repo-dir", ".", "Directory within the connected repository")
	sandboxCmd.Flags().StringVar(&connectedBranch, "connected-branch", "main", "Branch of the connected repository")

	// Drift schedule
	sandboxCmd.Flags().StringVar(&driftSchedule, "drift-schedule", "", "Cron expression for drift detection")

	// Environment variables and Terraform variables
	sandboxCmd.Flags().StringArrayVar(&envVars, "env-var", []string{}, "Environment variable in key=value format (can be specified multiple times)")
	sandboxCmd.Flags().StringArrayVar(&vars, "var", []string{}, "Terraform variable in key=value format (can be specified multiple times)")
	sandboxCmd.Flags().StringArrayVar(&varFiles, "var-file", []string{}, "Terraform variable file path (can be specified multiple times)")

	return sandboxCmd
}

func (c *cli) initStackCmd() *cobra.Command {
	var (
		// Stack-specific flags
		stackType               string
		stackName               string
		stackDescription        string
		vpcNestedTemplateURL    string
		runnerNestedTemplateURL string
	)

	stackCmd := &cobra.Command{
		Use:               "stack",
		Short:             "Initialize stack configuration",
		Long:              "Generate the stack.toml configuration file with custom parameters",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			genParams := initRootParams(cmd)

			params := apps.StackParams{
				Type:                    stackType,
				Name:                    stackName,
				Description:             stackDescription,
				VPCNestedTemplateURL:    vpcNestedTemplateURL,
				RunnerNestedTemplateURL: runnerNestedTemplateURL,
			}

			svc := apps.New(c.v, c.apiClient, c.cfg)
			err := svc.InitStackConfig(cmd.Context(), genParams, params)
			if err != nil {
				return err
			}

			successMesssage(genParams.Path, "stack.toml")
			return nil
		}),
	}

	// Stack-specific flags
	stackCmd.Flags().StringVar(&stackType, "type", "aws-cloudformation", "Type of infrastructure stack")
	stackCmd.Flags().StringVar(&stackName, "name", "", "Name of the CloudFormation stack (required)")
	stackCmd.Flags().StringVar(&stackDescription, "description", "", "Description of the stack (required)")
	stackCmd.Flags().StringVar(&vpcNestedTemplateURL, "vpc-template-url", "", "URL to the CloudFormation nested template for VPC resources")
	stackCmd.Flags().StringVar(&runnerNestedTemplateURL, "runner-template-url", "", "URL to the CloudFormation nested template for runner infrastructure")

	// Mark required flags
	stackCmd.MarkFlagRequired("name")
	stackCmd.MarkFlagRequired("description")

	return stackCmd
}

func (c *cli) initRunnerCmd() *cobra.Command {
	var (
		// Runner-specific flags
		runnerType    string
		envVars       []string
		helmDriver    string
		initScriptURL string
	)

	runnerCmd := &cobra.Command{
		Use:               "runner",
		Short:             "Initialize runner configuration",
		Long:              "Generate the runner.toml configuration file with custom parameters",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			genParams := initRootParams(cmd)

			params := apps.RunnerParams{
				RunnerType:    runnerType,
				EnvVars:       parseKeyValuePairs(envVars),
				HelmDriver:    helmDriver,
				InitScriptURL: initScriptURL,
			}

			svc := apps.New(c.v, c.apiClient, c.cfg)
			err := svc.InitRunnerConfig(cmd.Context(), genParams, params)
			if err != nil {
				return err
			}

			successMesssage(genParams.Path, "runner.toml")
			return nil
		}),
	}

	// Runner-specific flags
	runnerCmd.Flags().StringVar(&runnerType, "runner-type", "kubernetes", "Type of runner (kubernetes, docker, vm)")
	runnerCmd.Flags().StringArrayVar(&envVars, "env-var", []string{}, "Environment variable in key=value format (can be specified multiple times)")
	runnerCmd.Flags().StringVar(&helmDriver, "helm-driver", "", "Helm backend driver (e.g., 'configmap', 'secret')")
	runnerCmd.Flags().StringVar(&initScriptURL, "init-script-url", "", "URL to initialization script")

	// Mark required flags
	runnerCmd.MarkFlagRequired("runner-type")

	return runnerCmd
}

func (c *cli) initComponentTerraformModuleCmd() *cobra.Command {
	var (
		// Component flags
		componentName string
		varName       string
		dependencies  []string

		// Terraform module flags
		terraformVersion string
		envVars          []string
		vars             []string
		varFiles         []string
		publicRepo       string
		publicRepoDir    string
		publicRepoBranch string
		connectedRepo    string
		connectedRepoDir string
		connectedBranch  string
		driftSchedule    string
	)

	cmd := &cobra.Command{
		Use:               "terraform-module",
		Short:             "Initialize Terraform module component",
		Long:              "Generate a Terraform module component configuration file",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			genParams := initRootParams(cmd)

			params := apps.TerraformModuleComponentParams{
				Name:             componentName,
				VarName:          varName,
				Dependencies:     dependencies,
				TerraformVersion: terraformVersion,
				EnvVars:          parseKeyValuePairs(envVars),
				Vars:             parseKeyValuePairs(vars),
				VarFiles:         varFiles,
				PublicRepo:       publicRepo,
				PublicRepoDir:    publicRepoDir,
				PublicRepoBranch: publicRepoBranch,
				ConnectedRepo:    connectedRepo,
				ConnectedRepoDir: connectedRepoDir,
				ConnectedBranch:  connectedBranch,
				DriftSchedule:    driftSchedule,
			}

			svc := apps.New(c.v, c.apiClient, c.cfg)
			err := svc.InitTerraformModuleComponentConfig(cmd.Context(), genParams, params)
			if err != nil {
				return err
			}

			successMesssage(genParams.Path, fmt.Sprintf("components/%s.toml", params.Name))
			return nil
		}),
	}

	// Component flags
	cmd.Flags().StringVar(&componentName, "name", "", "Component name (required)")
	cmd.Flags().StringVar(&varName, "var-name", "", "Variable name for component output")
	cmd.Flags().StringArrayVar(&dependencies, "dependency", []string{}, "Component dependencies (can be specified multiple times)")

	// Terraform module flags
	cmd.Flags().StringVar(&terraformVersion, "terraform-version", "1.11.3", "Terraform version")
	cmd.Flags().StringArrayVar(&envVars, "env-var", []string{}, "Environment variable in key=value format (can be specified multiple times)")
	cmd.Flags().StringArrayVar(&vars, "var", []string{}, "Terraform variable in key=value format (can be specified multiple times)")
	cmd.Flags().StringArrayVar(&varFiles, "var-file", []string{}, "Terraform variable file path (can be specified multiple times)")
	cmd.Flags().StringVar(&publicRepo, "public-repo", "", "Public repository")
	cmd.Flags().StringVar(&publicRepoDir, "public-repo-dir", ".", "Directory within the public repository")
	cmd.Flags().StringVar(&publicRepoBranch, "public-repo-branch", "main", "Branch of the public repository")
	cmd.Flags().StringVar(&connectedRepo, "connected-repo", "", "Connected repository")
	cmd.Flags().StringVar(&connectedRepoDir, "connected-repo-dir", ".", "Directory within the connected repository")
	cmd.Flags().StringVar(&connectedBranch, "connected-branch", "main", "Branch of the connected repository")
	cmd.Flags().StringVar(&driftSchedule, "drift-schedule", "", "Cron expression for drift detection")

	// Mark required flags
	cmd.MarkFlagRequired("name")

	return cmd
}

func (c *cli) initComponentHelmChartCmd() *cobra.Command {
	var (
		// Component flags
		componentName string
		varName       string
		dependencies  []string

		// Helm chart flags
		chartName        string
		values           []string
		valuesFiles      []string
		publicRepo       string
		publicRepoDir    string
		publicRepoBranch string
		connectedRepo    string
		connectedRepoDir string
		connectedBranch  string
		helmRepoURL      string
		helmChart        string
		helmVersion      string
		namespace        string
		storageDriver    string
		takeOwnership    bool
		driftSchedule    string
	)

	cmd := &cobra.Command{
		Use:               "helm-chart",
		Short:             "Initialize Helm chart component",
		Long:              "Generate a Helm chart component configuration file",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			genParams := initRootParams(cmd)

			params := apps.HelmChartComponentParams{
				Name:             componentName,
				VarName:          varName,
				Dependencies:     dependencies,
				ChartName:        chartName,
				Values:           parseKeyValuePairs(values),
				ValuesFiles:      valuesFiles,
				PublicRepo:       publicRepo,
				PublicRepoDir:    publicRepoDir,
				PublicRepoBranch: publicRepoBranch,
				ConnectedRepo:    connectedRepo,
				ConnectedRepoDir: connectedRepoDir,
				ConnectedBranch:  connectedBranch,
				HelmRepoURL:      helmRepoURL,
				HelmChart:        helmChart,
				HelmVersion:      helmVersion,
				Namespace:        namespace,
				StorageDriver:    storageDriver,
				TakeOwnership:    takeOwnership,
				DriftSchedule:    driftSchedule,
			}

			svc := apps.New(c.v, c.apiClient, c.cfg)
			err := svc.InitHelmChartComponentConfig(cmd.Context(), genParams, params)
			if err != nil {
				return err
			}

			successMesssage(genParams.Path, fmt.Sprintf("components/%s.toml", params.Name))
			return nil
		}),
	}

	// Component flags
	cmd.Flags().StringVar(&componentName, "name", "", "Component name (required)")
	cmd.Flags().StringVar(&varName, "var-name", "", "Variable name for component output")
	cmd.Flags().StringArrayVar(&dependencies, "dependency", []string{}, "Component dependencies (can be specified multiple times)")

	// Helm chart flags
	cmd.Flags().StringVar(&chartName, "chart-name", "", "Helm chart name (required)")
	cmd.Flags().StringArrayVar(&values, "value", []string{}, "Helm value in key=value format (can be specified multiple times)")
	cmd.Flags().StringArrayVar(&valuesFiles, "values-file", []string{}, "Helm values file path (can be specified multiple times)")
	cmd.Flags().StringVar(&publicRepo, "public-repo", "", "Public repository with helm chart")
	cmd.Flags().StringVar(&publicRepoDir, "public-repo-dir", ".", "Directory within the public repository")
	cmd.Flags().StringVar(&publicRepoBranch, "public-repo-branch", "main", "Branch of the public repository")
	cmd.Flags().StringVar(&connectedRepo, "connected-repo", "", "Connected repository with helm chart")
	cmd.Flags().StringVar(&connectedRepoDir, "connected-repo-dir", ".", "Directory within the connected repository")
	cmd.Flags().StringVar(&connectedBranch, "connected-branch", "main", "Branch of the connected repository")
	cmd.Flags().StringVar(&helmRepoURL, "helm-repo-url", "", "Helm repository URL")
	cmd.Flags().StringVar(&helmChart, "helm-chart", "", "Helm chart name in repository")
	cmd.Flags().StringVar(&helmVersion, "helm-version", "", "Helm chart version")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace")
	cmd.Flags().StringVar(&storageDriver, "storage-driver", "", "Helm storage driver (configmap, secret)")
	cmd.Flags().BoolVar(&takeOwnership, "take-ownership", false, "Take ownership of existing releases")
	cmd.Flags().StringVar(&driftSchedule, "drift-schedule", "", "Cron expression for drift detection")

	// Mark required flags
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("chart-name")

	return cmd
}

func (c *cli) initComponentKubernetesManifestCmd() *cobra.Command {
	var (
		// Component flags
		componentName string
		varName       string
		dependencies  []string

		// Kubernetes manifest flags
		manifest      string
		namespace     string
		driftSchedule string
	)

	cmd := &cobra.Command{
		Use:               "kubernetes-manifest",
		Short:             "Initialize Kubernetes manifest component",
		Long:              "Generate a Kubernetes manifest component configuration file",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			genParams := initRootParams(cmd)

			params := apps.KubernetesManifestComponentParams{
				Name:          componentName,
				VarName:       varName,
				Dependencies:  dependencies,
				Manifest:      manifest,
				Namespace:     namespace,
				DriftSchedule: driftSchedule,
			}

			svc := apps.New(c.v, c.apiClient, c.cfg)
			err := svc.InitKubernetesManifestComponentConfig(cmd.Context(), genParams, params)
			if err != nil {
				return err
			}

			successMesssage(genParams.Path, fmt.Sprintf("components/%s.toml", params.Name))
			return nil
		}),
	}

	// Component flags
	cmd.Flags().StringVar(&componentName, "name", "", "Component name (required)")
	cmd.Flags().StringVar(&varName, "var-name", "", "Variable name for component output")
	cmd.Flags().StringArrayVar(&dependencies, "dependency", []string{}, "Component dependencies (can be specified multiple times)")

	// Kubernetes manifest flags
	cmd.Flags().StringVar(&manifest, "manifest", "", "Kubernetes manifest YAML content (required)")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace (required)")
	cmd.Flags().StringVar(&driftSchedule, "drift-schedule", "", "Cron expression for drift detection")

	// Mark required flags
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("manifest")
	cmd.MarkFlagRequired("namespace")

	return cmd
}

func (c *cli) initComponentCmd() *cobra.Command {
	componentCmd := &cobra.Command{
		Use:   "component",
		Short: "Initialize component configuration",
		Long:  "Generate component configuration files for terraform modules, helm charts, or kubernetes manifests",
	}

	// Add subcommands for each component type
	componentCmd.AddCommand(c.initComponentTerraformModuleCmd())
	componentCmd.AddCommand(c.initComponentHelmChartCmd())
	componentCmd.AddCommand(c.initComponentKubernetesManifestCmd())

	return componentCmd
}

func splitKeyValue(s string) []string {
	return strings.SplitN(s, "=", 2)
}

// parseKeyValuePairs string slice in key=value format to map[string]string
func parseKeyValuePairs(pairs []string) map[string]string {
	result := make(map[string]string)
	for _, pair := range pairs {
		parts := splitKeyValue(pair)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

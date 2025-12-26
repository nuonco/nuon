package init

import (
	"log"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/nuonco/nuon/bins/cli/internal/services/apps"
	"github.com/pkg/errors"
)

type ConfigInit struct{}

func NewConfigInit() *ConfigInit {
	return &ConfigInit{}
}

// handle error in all functions
func (c *ConfigInit) RunInitMenu(params *apps.InitParams) (*apps.InitParams, error) {
	var configType string
	var prebuiltTemplate string
	var appName string
	var stackType string
	var runnerType string
	var componentTypes []string
	var enableActionsAddition bool

	// Initial question: custom or prebuilt
	initialForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Configuration Type").
				Description("Choose how you want to initialize your app config").
				Options(
					huh.NewOption("Prebuilt Template", "prebuilt"),
					huh.NewOption("Custom Configuration", "custom"),
				).
				Value(&configType),
		),
	)

	err := initialForm.Run()
	if err != nil {
		log.Fatal(err)
	}

	// If prebuilt, show template options
	if configType == "prebuilt" {
		templateForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select Template").
					Description("Choose a prebuilt template for your app").
					Options(
						huh.NewOption("AWS EKS (Elastic Kubernetes Service)", "aws-eks"),
						huh.NewOption("AWS ECS (Elastic Container Service)", "aws-ecs"),
					).
					Value(&prebuiltTemplate),
			),
		)

		err = templateForm.Run()
		if err != nil {
			return nil, err
		}

		params.PrebuiltTemplate = prebuiltTemplate
		return params, nil
	}

	// If custom, show all configuration options
	customForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("App Name").
				Description("Enter a name for your application").
				Value(&appName).
				Validate(func(str string) error {
					if strings.TrimSpace(str) == "" {
						return errors.New("app name cannot be empty")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Stack Type").
				Description("Choose your cloud provider").
				Options(
					huh.NewOption("AWS", "aws"),
					huh.NewOption("Azure", "azure"),
				).
				Value(&stackType),

			huh.NewSelect[string]().
				Title("Runner Type").
				Description("Choose your runner deployment type").
				Options(
					huh.NewOption("ECS", "ecs"),
					huh.NewOption("EKS", "eks"),
					huh.NewOption("AKS", "aks"),
					huh.NewOption("VM", "vm"),
				).
				Value(&runnerType),
		),

		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Component Types").
				Description("Select component types to include in your app").
				Options(
					huh.NewOption("Terraform Module", "terraform-module"),
					huh.NewOption("Helm Chart", "helm-chart"),
					huh.NewOption("Kubernetes Manifest", "kubernetes-manifest"),
				).
				Value(&componentTypes),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Actions Configuration").
				Description("Include actions config sample?").
				Value(&enableActionsAddition),
		),
	)

	err = customForm.Run()
	if err != nil {
		return nil, err
	}

	params.AppName = appName
	params.StackType = stackType
	params.RunnerType = runnerType
	params.ComponentTypes = componentTypes
	params.Actions = append(params.Actions, "sample_action")

	return params, nil
}

func (c *ConfigInit) RunGeneratorConfigMenu(params *apps.ConfigGenParams) error {
	var enableDefaults bool
	var enableComments bool
	customForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable default values?").
				Description("Include default values in generated config files").
				Value(&enableDefaults),

			huh.NewConfirm().
				Title("Enable comments?").
				Description("Include helpful comments in generated config files").
				Value(&enableComments),
		),
	)

	err := customForm.Run()
	if err != nil {
		return err
	}

	params.EnableDefaults = enableDefaults
	params.EnableComments = enableComments

	return nil
}

func (c *ConfigInit) RunComponentsMenu(params *apps.SampleComponentParams) error {
	var componentTypes []string
	var enableComponentAddition bool

	// Initial question: custom or prebuilt
	initialForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Component Configuration").
				Description("Include component config samples?").
				Value(&enableComponentAddition),
		),
	)

	err := initialForm.Run()
	if err != nil {
		return errors.Wrap(err, "unable to render component selection form")
	}

	if !enableComponentAddition {
		return nil
	}

	// If custom, show all configuration options
	componentForm := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Component Types").
				Description("Select component types to include in your app").
				Options(
					huh.NewOption("Terraform Module", "terraform-module"),
					huh.NewOption("Helm Chart", "helm-chart"),
					huh.NewOption("Kubernetes Manifest", "kubernetes-manifest"),
				).
				Value(&componentTypes),
		),
	)

	err = componentForm.Run()
	if err != nil {
		log.Fatal(err)
	}

	params.ComponentTypes = componentTypes

	return nil
}

func (c *ConfigInit) RunActionsMenu(params *apps.SampleActionsParams) error {
	// var actions []string
	var enableActionsAdditions bool

	// Initial question: custom or prebuilt
	initialForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Actions Configuration").
				Description("Include actions config sample?").
				Value(&enableActionsAdditions),
		),
	)

	err := initialForm.Run()
	if err != nil {
		log.Fatal(err)
	}

	if !enableActionsAdditions {
		return nil
	}

	// If custom, show all configuration options
	componentForm := huh.NewForm(
	// huh.NewGroup(
	// 	huh.NewMultiSelect[string]().
	// 		Title("Actions").
	// 		Description("Select actions to include in app config").
	// 		Options(
	// 			huh.NewOption("Test action", "test"),
	// 		).
	// 		Value(&actions),
	// ),
	)

	err = componentForm.Run()
	if err != nil {
		log.Fatal(err)
	}

	params.Actions = make([]string, 1)

	return nil
}

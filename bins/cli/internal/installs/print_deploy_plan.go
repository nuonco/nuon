package installs

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

type DeployPlanActual struct {
	Waypoint_Plan struct {
		Variables struct {
			Intermediate_Data interface{}
			Variables         []struct {
				Actual *struct {
					TerraformVariable *struct {
						Name  string
						Value string
					}
					HelmValue *struct {
						Name  string
						Value string
					}
				}
			}
		}
		Waypoint_Job struct {
			Hcl_Config string
		}
	}
}

func (s *Service) PrintDeployPlan(ctx context.Context, installID, deployID string, asJSON, renderedVars, intermediateOnly, jobConfig bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	plan, err := s.api.GetInstallDeployPlan(ctx, installID, deployID)
	if err != nil {
		view.Error(err)
		return
	}

	if renderedVars || intermediateOnly || jobConfig {
		var p DeployPlanActual
		err = mapstructure.Decode(plan.Actual, &p)
		if err != nil {
			ui.PrintError(err)
			return
		}

		if renderedVars {
			if asJSON {
				ui.PrintJSON(p.Waypoint_Plan.Variables.Variables)
				return
			}

			data := [][]string{{
				"TYPE",
				"NAME",
				"VALUE",
			}}

			for _, v := range p.Waypoint_Plan.Variables.Variables {
				if v.Actual.TerraformVariable != nil {
					data = append(data, []string{
						"Terraform",
						v.Actual.TerraformVariable.Name,
						v.Actual.TerraformVariable.Value,
					})
				}

				if v.Actual.HelmValue != nil {
					data = append(data, []string{
						"Helm",
						v.Actual.HelmValue.Name,
						v.Actual.HelmValue.Value,
					})
				}
			}

			view.Render(data)
			return
		}

		if intermediateOnly {
			ui.PrintJSON(p.Waypoint_Plan.Variables.Intermediate_Data)
			return
		}

		if jobConfig {
			if asJSON {
				ui.PrintJSON(p.Waypoint_Plan.Waypoint_Job)
				return
			}

			fmt.Printf("%s", p.Waypoint_Plan.Waypoint_Job.Hcl_Config)
			return
		}
	}

	ui.PrintJSON(plan)
}

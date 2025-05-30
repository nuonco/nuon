package helpers

import (
	"github.com/powertoolsdev/mono/pkg/config/refs"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// NOTE(jm): this is a function, because it is used directly within a workflow.
func GetComponentReferences(cfg *app.AppConfig, comp string) []refs.Ref {
	rfs := make([]refs.Ref, 0)
	for _, cc := range cfg.ComponentConfigConnections {
		for _, ref := range cc.Refs {
			if ref.Type == refs.RefTypeComponents && ref.Name == comp {
				rfs = append(rfs, ref)
			}
		}
	}

	for _, cc := range cfg.ActionWorkflowConfigs {
		for _, ref := range cc.Refs {
			if ref.Type == refs.RefTypeComponents && ref.Name == comp {
				rfs = append(rfs, ref)
			}
		}
	}

	return rfs
}

func GetActionReferences(cfg *app.AppConfig, action string) []refs.Ref {
	rfs := make([]refs.Ref, 0)
	for _, cc := range cfg.ComponentConfigConnections {
		for _, ref := range cc.Refs {
			if ref.Type == refs.RefTypeActions && ref.Name == action {
				rfs = append(rfs, ref)
			}
		}
	}
	for _, cc := range cfg.ActionWorkflowConfigs {
		for _, ref := range cc.Refs {
			if ref.Type == refs.RefTypeActions && ref.Name == action {
				rfs = append(rfs, ref)
			}
		}
	}

	return rfs
}

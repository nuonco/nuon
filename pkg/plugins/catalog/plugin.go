package catalog

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	imageURLTemplate string = "public.ecr.aws/p7e3r5y0/%s"
)

func ToPluginType(val string) (PluginType, error) {
	for k, v := range pluginTypeNames {
		if val == v {
			return k, nil
		}
	}

	return PluginTypeUnknown, fmt.Errorf("unable to find plugin type for %s", val)
}

var pluginTypeNames map[PluginType]string = map[PluginType]string{
	PluginTypeDefault:   "default",
	PluginTypeTerraform: "terraform",
	PluginTypeHelm:      "helm",
	PluginTypeExp:       "exp",
	PluginTypeOci:       "oci",
	PluginTypeOciSync:   "oci-sync",
	PluginTypeDev:       "dev",
	PluginTypeNoop:      "noop",
}

type PluginType int

const (
	PluginTypeDefault PluginType = iota + 1
	PluginTypeTerraform
	PluginTypeExp
	PluginTypeDev
	PluginTypeUnknown
	PluginTypeNoop
	PluginTypeOci
	PluginTypeOciSync
	PluginTypeHelm
)

// ImageURL returns the image url to access this using our default, public ecr repo
func (p PluginType) ImageURL() string {
	return fmt.Sprintf(imageURLTemplate, p.RepositoryName())
}

func (p PluginType) String() string {
	return pluginTypeNames[p]
}

// RepositoryName is used to return the name of a repository. We hard code these here, to prevent having to configure
// them and/or describe the repository in the public ecr.
func (p PluginType) RepositoryName() string {
	switch p {
	case PluginTypeDefault:
		return "waypoint-odr"
	case PluginTypeTerraform:
		return "waypoint-plugin-terraform"
	case PluginTypeDev:
		// TODO(jm): rename this
		return "dev-public"
	case PluginTypeExp:
		return "waypoint-plugin-exp"
	case PluginTypeNoop:
		return "waypoint-plugin-noop"
	case PluginTypeHelm:
		return "waypoint-plugin-helm"
	case PluginTypeOci:
		return "waypoint-plugin-oci"
	case PluginTypeOciSync:
		return "waypoint-plugin-oci-sync"
	default:
	}
	return ""
}

// Plugin is used to return the information needed to use a plugin for an ODR
type Plugin struct {
	ImageURL       string    `validate:"required"`
	Tag            string    `validate:"required"`
	RepositoryName string    `validate:"required"`
	CreatedAt      time.Time `validate:"required"`
}

func (p Plugin) Validate(v validator.Validate) error {
	if err := v.Struct(p); err != nil {
		return fmt.Errorf("unable to validate plugin: %w", err)
	}

	return nil
}

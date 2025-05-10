package cloudformation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type TemplateInput struct {
	Install                    *app.Install             `validate:"required"`
	CloudFormationStackVersion *app.InstallStackVersion `validate:"required"`
	InstallState               *state.State             `validate:"required"`
	AppCfg                     *app.AppConfig           `validate:"required"`

	Runner   *app.Runner              `validate:"required"`
	Settings *app.RunnerGroupSettings `validate:"required"`
	APIToken string                   `validate:"required"`
}

func (t *Templates) Template(inputs *TemplateInput) (*cloudformation.Template, string, error) {
	tmpl, err := t.getAWSTemplate(inputs)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to create aws-eks template")
	}

	// Marshal the template to JSON
	jsonBytes, err := json.Marshal(tmpl)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to marshal template to JSON")
	}

	hash := sha256.Sum256(jsonBytes)
	checksum := hex.EncodeToString(hash[:])

	return tmpl, checksum, nil
}

package stack

import (
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
	"github.com/nuonco/nuon/sdks/stack/models"
)

// configFromModel converts the generated wire type into the SDK's working
// Config. Both are generated from, or asserted against, the same json tags, so
// re-encoding copies every field without naming any of them — a field added to
// the spec arrives without a code change here. TestWireFieldsHaveCoreFields
// fails if a tag ever stops lining up.
//
// Config is a strict superset of the wire type: the provisioning-method fields
// (Method, Terraform*, TerraformBackend) and the customer-supplied GCP inputs
// have no wire representation and are filled from Options at provision time.
func configFromModel(m *models.AppInstallerSDKConfig) (*core.Config, error) {
	if m == nil {
		return nil, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("encode stack config: %w", err)
	}

	var cfg core.Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("decode stack config: %w", err)
	}

	return &cfg, nil
}

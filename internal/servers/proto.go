package servers

import (
	"encoding/json"

	"github.com/powertoolsdev/api/internal/models"
	orgv1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
)

// Org model to proto converts org domain model into org proto message
func OrgModelToProto(org *models.Org) (*orgv1.Org, error) {
	orgJSON, err := json.Marshal(org)
	if err != nil {
		return nil, err
	}

	orgProto := orgv1.Org{}

	if err := json.Unmarshal(orgJSON, &orgProto); err != nil {
		return nil, err
	}

	return &orgProto, nil
}

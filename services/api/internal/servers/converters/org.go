package converters

import (
	orgv1 "github.com/powertoolsdev/mono/pkg/types/api/org/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

// Org model to proto converts org domain model into org proto message
func OrgModelToProto(org *models.Org) *orgv1.Org {
	return &orgv1.Org{
		Id:        org.ID.String(),
		Name:      org.Name,
		OwnerId:   org.CreatedByID,
		UpdatedAt: TimeToDatetime(org.UpdatedAt),
		CreatedAt: TimeToDatetime(org.CreatedAt),
	}
}

// OrgModelsToProtos converts a slice of org models to protos
func OrgModelsToProtos(orgs []*models.Org) []*orgv1.Org {
	protos := make([]*orgv1.Org, len(orgs))
	for idx, org := range orgs {
		protos[idx] = OrgModelToProto(org)
	}

	return protos
}

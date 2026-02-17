package plantypes

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func PlanAuthFromSDK(planAuth *models.PlantypesPlanAuth) (*plantypes.PlanAuth, error) {
	var auth plantypes.PlanAuth
	planAuthBytes, err := json.Marshal(planAuth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build plan auth")
	}

	err = json.Unmarshal(planAuthBytes, &auth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build plan auth")
	}

	return &auth, nil
}

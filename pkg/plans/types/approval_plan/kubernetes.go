package approvalplan

import (
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/types/components/plan"
)

type KubernetesApprovalPlan struct {
	PlanJSON []byte `json:"plan_json"`
}

func NewKubernetesApprovalPlan(planJSON []byte) *KubernetesApprovalPlan {
	return &KubernetesApprovalPlan{
		PlanJSON: planJSON,
	}
}

func (k *KubernetesApprovalPlan) IsNoop() (bool, error) {
	kplan := &plan.KubernetesManifestPlanContents{}
	err := json.Unmarshal(k.PlanJSON, kplan)
	if err != nil {
		return false, fmt.Errorf("unable to unmarshal kubernetes plan json: %w", err)
	}

	if len(kplan.ContentDiff) == 0 {
		return true, nil
	}

	return false, nil
}

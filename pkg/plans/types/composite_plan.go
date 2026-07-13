package plantypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/kube"
)

type CompositePlan struct {
	BuildPlan              *BuildPlan              `json:"build_plan,omitempty"`
	DeployPlan             *DeployPlan             `json:"deploy_plan,omitempty"`
	ActionWorkflowRunPlan  *ActionWorkflowRunPlan  `json:"action_workflow_run_plan,omitempty"`
	SyncSecretsPlan        *SyncSecretsPlan        `json:"sync_secrets_plan,omitempty"`
	SyncOCIPlan            *SyncOCIPlan            `json:"sync_oci_plan,omitempty"`
	FetchImageMetadataPlan *FetchImageMetadataPlan `json:"fetch_image_metadata_plan,omitempty"`
	SandboxRunPlan         *SandboxRunPlan         `json:"sandbox_run_plan,omitempty"`
}

// CompositePlanFromAny converts any composite-plan-shaped value (e.g. a
// generated SDK model) into a CompositePlan via a JSON round-trip. Both the SDK
// models and this type derive from the same schema, so the tags line up.
func CompositePlanFromAny(v any) (*CompositePlan, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var cp CompositePlan
	if err := json.Unmarshal(bytes, &cp); err != nil {
		return nil, err
	}

	cp.pruneEmptyClusterAuth()
	return &cp, nil
}

// pruneEmptyClusterAuth clears cluster-info auth configs that decoded as empty
// objects. go-swagger generates ClusterInfo.aws_auth as an inline struct value
// (not a pointer), so a null aws_auth from the API re-marshals as {} during the
// round-trip above and would otherwise decode into a non-nil empty config,
// routing kube auth down the AWS path on non-AWS installs.
func (cp *CompositePlan) pruneEmptyClusterAuth() {
	for _, ci := range cp.clusterInfos() {
		if ci == nil {
			continue
		}
		if ci.AWSAuth != nil && *ci.AWSAuth == (awscredentials.Config{}) {
			ci.AWSAuth = nil
		}
	}
}

func (cp *CompositePlan) clusterInfos() []*kube.ClusterInfo {
	infos := make([]*kube.ClusterInfo, 0, 6)
	if dp := cp.DeployPlan; dp != nil {
		if dp.HelmDeployPlan != nil {
			infos = append(infos, dp.HelmDeployPlan.ClusterInfo)
		}
		if dp.TerraformDeployPlan != nil {
			infos = append(infos, dp.TerraformDeployPlan.ClusterInfo)
		}
		if dp.KubernetesManifestDeployPlan != nil {
			infos = append(infos, dp.KubernetesManifestDeployPlan.ClusterInfo)
		}
		if dp.PulumiDeployPlan != nil {
			infos = append(infos, dp.PulumiDeployPlan.ClusterInfo)
		}
	}
	if cp.ActionWorkflowRunPlan != nil {
		infos = append(infos, cp.ActionWorkflowRunPlan.ClusterInfo)
	}
	if cp.SyncSecretsPlan != nil {
		infos = append(infos, cp.SyncSecretsPlan.ClusterInfo)
	}
	return infos
}

// Inner returns the single populated sub-plan, or nil when the composite plan is
// empty. Exactly one sub-plan is set per job.
func (cp *CompositePlan) Inner() any {
	switch {
	case cp.BuildPlan != nil:
		return cp.BuildPlan
	case cp.DeployPlan != nil:
		return cp.DeployPlan
	case cp.ActionWorkflowRunPlan != nil:
		return cp.ActionWorkflowRunPlan
	case cp.SyncSecretsPlan != nil:
		return cp.SyncSecretsPlan
	case cp.SyncOCIPlan != nil:
		return cp.SyncOCIPlan
	case cp.FetchImageMetadataPlan != nil:
		return cp.FetchImageMetadataPlan
	case cp.SandboxRunPlan != nil:
		return cp.SandboxRunPlan
	default:
		return nil
	}
}

func (cp CompositePlan) Value() (driver.Value, error) {
	if cp.IsEmpty() {
		return nil, nil
	}
	return json.Marshal(cp)
}

func (cp CompositePlan) IsEmpty() bool {
	return cp.BuildPlan == nil &&
		cp.DeployPlan == nil &&
		cp.ActionWorkflowRunPlan == nil &&
		cp.SyncSecretsPlan == nil &&
		cp.SyncOCIPlan == nil &&
		cp.FetchImageMetadataPlan == nil &&
		cp.SandboxRunPlan == nil
}

func (cp *CompositePlan) Scan(value any) error {
	if value == nil {
		*cp = CompositePlan{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into CompositePlan", value)
	}

	if len(bytes) == 0 {
		*cp = CompositePlan{}
		return nil
	}

	return json.Unmarshal(bytes, cp)
}

// GormDataType tells GORM what database type to use
func (CompositePlan) GormDataType() string {
	return "jsonb"
}

// GormDBDataType returns the database data type based on the current using database
func (CompositePlan) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Name() {
	case "postgres":
		return "JSONB"
	case "mysql":
		return "JSON"
	case "sqlite":
		return "TEXT"
	default:
		return "TEXT"
	}
}

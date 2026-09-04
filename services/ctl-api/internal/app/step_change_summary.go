package app

type StepChangePlanType string

const (
	StepChangePlanTypeTerraform  StepChangePlanType = "terraform_plan"
	StepChangePlanTypePulumi     StepChangePlanType = "pulumi_plan"
	StepChangePlanTypeHelm       StepChangePlanType = "helm_approval"
	StepChangePlanTypeKubernetes StepChangePlanType = "kubernetes_manifest_approval"
	StepChangePlanTypeAppBranch  StepChangePlanType = "app_branch_plan"
	StepChangePlanTypeInstall    StepChangePlanType = "install_creation"
)

type StepChangeStatus string

const (
	StepChangeStatusPendingApproval StepChangeStatus = "pending-approval"
	StepChangeStatusApproved        StepChangeStatus = "approved"
	StepChangeStatusDenied          StepChangeStatus = "denied"
	StepChangeStatusApplied         StepChangeStatus = "applied"
	StepChangeStatusGenerating      StepChangeStatus = "generating"
	StepChangeStatusError           StepChangeStatus = "error"
)

type StepChangeCounts struct {
	Create  int `json:"create" binding:"required"`
	Update  int `json:"update" binding:"required"`
	Delete  int `json:"delete" binding:"required"`
	Replace int `json:"replace" binding:"required"`
	Noop    int `json:"noop" binding:"required"`
}

type StepChangeSummary struct {
	StepID        string             `json:"step_id" binding:"required"`
	StepName      string             `json:"step_name" binding:"required"`
	ApprovalID    string             `json:"approval_id" binding:"required"`
	ComponentName string             `json:"component_name,omitempty"`
	PlanType      StepChangePlanType `json:"plan_type" binding:"required"`
	Status        StepChangeStatus   `json:"status" binding:"required"`
	Counts        StepChangeCounts   `json:"counts" binding:"required"`
	HasDetail     bool               `json:"has_detail" binding:"required"`
}

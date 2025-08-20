package approvalplan

type ApprovalPlan interface {
	IsNoop() (bool, error)
}

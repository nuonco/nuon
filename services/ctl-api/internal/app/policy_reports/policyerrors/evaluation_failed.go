package policyerrors

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"

const EvaluationFailedErrorType compositeerrors.Type = "policy.evaluation_failed"

type EvaluationFailureStage string

const (
	EvaluationFailureStagePreparation EvaluationFailureStage = "preparation"
	EvaluationFailureStageEvaluation  EvaluationFailureStage = "evaluation"
)

type EvaluationFailedError struct {
	Stage EvaluationFailureStage `json:"stage"`
}

var _ compositeerrors.CompositeError = (*EvaluationFailedError)(nil)

func (*EvaluationFailedError) Error() string {
	return "Policy evaluation could not be completed"
}

func (*EvaluationFailedError) Type() compositeerrors.Type {
	return EvaluationFailedErrorType
}

func (*EvaluationFailedError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityWarning
}

func (e *EvaluationFailedError) Sections() []compositeerrors.Section {
	whatHappened := "Nuon could not complete policy enforcement for this plan. The workflow continued without a policy result."
	switch e.Stage {
	case EvaluationFailureStagePreparation:
		whatHappened = "Nuon could not prepare the plan and configured policies for evaluation. The workflow continued without a policy result."
	case EvaluationFailureStageEvaluation:
		whatHappened = "Nuon prepared the policy inputs but could not finish evaluating the configured policies. The workflow continued without a policy result."
	}

	return []compositeerrors.Section{
		compositeerrors.MarkdownSection("What happened", whatHappened),
		compositeerrors.MarkdownSection("How to continue", "Review the plan carefully before approving it. Retry the plan to attempt policy evaluation again, and contact support if this warning persists."),
	}
}

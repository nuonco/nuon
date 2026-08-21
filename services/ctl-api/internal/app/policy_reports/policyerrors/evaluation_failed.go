package policyerrors

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"

const EvaluationFailedErrorType compositeerrors.Type = "policy.evaluation_failed"

type EvaluationFailureStage string

const (
	EvaluationFailureStagePreparation EvaluationFailureStage = "preparation"
	EvaluationFailureStageEvaluation  EvaluationFailureStage = "evaluation"
	EvaluationFailureStagePersistence EvaluationFailureStage = "persistence"
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
	whatHappened := "Nuon could not complete policy enforcement. The operation continued without a policy result."
	switch e.Stage {
	case EvaluationFailureStagePreparation:
		whatHappened = "Nuon could not prepare the input and configured policies for evaluation. The operation continued without a policy result."
	case EvaluationFailureStageEvaluation:
		whatHappened = "Nuon prepared the policy inputs but could not finish evaluating the configured policies. The operation continued without a policy result."
	case EvaluationFailureStagePersistence:
		whatHappened = "Nuon evaluated the configured policies but could not save the policy report. The operation continued without a durable policy result."
	}

	return []compositeerrors.Section{
		compositeerrors.MarkdownSection("What happened", whatHappened),
		compositeerrors.MarkdownSection("How to continue", "Review the resulting infrastructure or artifact carefully. Retry the operation to attempt policy evaluation again, and contact support if this warning persists."),
	}
}

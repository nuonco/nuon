package credentialcleanup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	acctcleanup "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/credentialcleanup"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = acctcleanup.SignalType

type Signal struct {
	ServiceAccountSubject string `json:"service_account_subject,omitempty"`
	AccountID             string `json:"account_id,omitempty"`
	OrgID                 string `json:"org_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(workflow.Context) error {
	hasSubject := s.ServiceAccountSubject != ""
	hasUser := s.AccountID != "" || s.OrgID != ""
	if hasSubject == hasUser {
		return fmt.Errorf("credential cleanup requires a service account subject or an account and org")
	}
	if hasUser && (s.AccountID == "" || s.OrgID == "") {
		return fmt.Errorf("credential cleanup for a user requires an account and org")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	if err := activities.AwaitCleanupCredentials(ctx, activities.CleanupCredentialsRequest{
		ServiceAccountSubject: s.ServiceAccountSubject,
		AccountID:             s.AccountID,
		OrgID:                 s.OrgID,
	}); err != nil {
		return fmt.Errorf("cleanup credentials: %w", err)
	}
	return nil
}

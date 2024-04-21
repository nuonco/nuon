package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
)

type SendEmailRequest struct {
	TransactionalEmailID string
	Email                string
	Variables            map[string]string
}

func (a *Activities) SendEmail(ctx context.Context, req SendEmailRequest) error {
	vars := generics.ToIntMap(req.Variables)

	err := a.loopsClient.SendTransactionalEmail(ctx, req.Email, req.TransactionalEmailID, vars)
	if err != nil {
		return fmt.Errorf("unable to send email: %w", err)
	}

	return nil
}

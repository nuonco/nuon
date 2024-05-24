package notifications

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (n *Notifications) sendEmailNotification(ctx context.Context, typ Type, vars map[string]string) error {
	if typ.EmailTemplateID() == "" {
		return nil
	}

	email, ok := vars["email"]
	if !ok || email == "" {
		return fmt.Errorf("email must be set as one field in vars")
	}

	intVars := generics.ToIntMap(vars)

	err := n.Loops.SendTransactionalEmail(ctx, email, typ.EmailTemplateID(), intVars)
	if err != nil {
		return fmt.Errorf("unable to send email: %w", err)
	}

	return nil
}

package context

import (
	"fmt"

	"github.com/nuonco/nuon-go/models"
)

// Context is an interface that requires a Value method. This is the only context method this package requires.
// This allows ContextWriter to work with the standard context.Context, Temporal's workflow.Context, and any other type that provides a Value method.
type Context interface {
	Value(any) any
}

const (
	orgCtxKey   string = "org"
	orgIDCtxKey string = "org_id"

	accountIDCtxKey string = "account_id"
	accountCtxKey   string = "account"
)

func AccountFromContext(ctx Context) (*models.AppAccount, error) {
	acct := ctx.Value(accountCtxKey)
	if acct == nil {
		return nil, fmt.Errorf("org was not set on context")
	}

	return acct.(*models.AppAccount), nil
}

func OrgFromContext(ctx Context) (*models.AppOrg, error) {
	org := ctx.Value(orgCtxKey)
	if org == nil {
		return nil, fmt.Errorf("org was not set on context")
	}

	return org.(*models.AppOrg), nil
}

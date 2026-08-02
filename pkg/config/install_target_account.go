package config

import (
	"errors"
	"fmt"
)

// CheckImmutableTargetAccount refuses a sync that would change the cloud account an
// existing install targets.
//
// The account is fixed at creation: UpdateInstallRequest carries no field for it, and
// the phone-home secret's cross-account grant names a role in that specific account.
// Without this guard the edit would either show as permanent drift or silently do
// nothing, depending on which interface made it.
//
// It lives here rather than in a caller because the rule has to hold across every
// interface that can sync an install — the CLI, and the server-side install config
// syncer — and both already reduce their inputs to a pair of config.Install values.
//
// upstream is nil on a create, where there is nothing to protect yet.
func (i *Install) CheckImmutableTargetAccount(upstream *Install) error {
	if i == nil || upstream == nil {
		return nil
	}

	type immutableField struct {
		key      string
		desired  string
		upstream string
	}
	var fields []immutableField

	if i.AWSAccount != nil && upstream.AWSAccount != nil {
		fields = append(fields, immutableField{
			"aws_account.account_id", i.AWSAccount.AccountID, upstream.AWSAccount.AccountID,
		})
	}
	if i.AzureAccount != nil && upstream.AzureAccount != nil {
		fields = append(fields, immutableField{
			"azure_account.subscription_id", i.AzureAccount.SubscriptionID, upstream.AzureAccount.SubscriptionID,
		})
	}
	if i.GCPAccount != nil && upstream.GCPAccount != nil {
		fields = append(fields, immutableField{
			"gcp_account.project_id", i.GCPAccount.ProjectID, upstream.GCPAccount.ProjectID,
		})
	}

	var errs []error
	for _, f := range fields {
		// An unset value in the config is "don't care", not "clear it".
		if f.desired == "" || f.desired == f.upstream {
			continue
		}
		errs = append(errs, fmt.Errorf(
			"refusing to change %s on existing install from %q to %q: the target cloud account is immutable after creation",
			f.key, f.upstream, f.desired))
	}
	if len(errs) > 0 {
		return fmt.Errorf("\n%w", errors.Join(errs...))
	}

	return nil
}

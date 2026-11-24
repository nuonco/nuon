package validation

import (
	"github.com/nuonco/nuon/pkg/urlpolicy"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

const urlNotAllowedCode = "url_not_allowed"

func CheckURLPolicy(policy *urlpolicy.Policy, violations urlpolicy.Violations) error {
	if len(violations) == 0 {
		return nil
	}
	if policy.Mode() == urlpolicy.ModeWarn {
		policy.ReportWarnings(violations)
		return nil
	}
	return stderr.ErrUser{
		Err:         violations,
		Code:        urlNotAllowedCode,
		Description: violations.Error(),
	}
}

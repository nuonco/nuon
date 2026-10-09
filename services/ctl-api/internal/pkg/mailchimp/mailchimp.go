package mailchimp

import (
	"github.com/go-playground/validator/v10"
	mailchimpclient "github.com/nuonco/nuon/pkg/mailchimp"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func New(v *validator.Validate, cfg *internal.Config) (mailchimpclient.Client, error) {
	if cfg.MailchimpAPIKey == "" || cfg.MailchimpServerPrefix == "" || cfg.MailchimpAudienceID == "" {
		return mailchimpclient.NewDisabled(), nil
	}

	return mailchimpclient.New(v,
		mailchimpclient.WithAPIKey(cfg.MailchimpAPIKey),
		mailchimpclient.WithServerPrefix(cfg.MailchimpServerPrefix),
		mailchimpclient.WithAudienceID(cfg.MailchimpAudienceID),
	)
}

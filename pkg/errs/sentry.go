package errs

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/report"
	"github.com/getsentry/sentry-go"
)

const (
	SentryMainDSN string = "https://a0546c06ff00cb18c7867c5783f96763@o4507623795523584.ingest.us.sentry.io/4507623799193600"
)

type SentryTagger interface {
	error
	ErrorTags() map[string]string
}

type SentryErrOptions struct {
	Tags   map[string]string
	UserID string
}

func ReportToSentry(err error, opt *SentryErrOptions) string {
	event, extraDetails := report.BuildSentryReport(err)

	if hints := errors.GetAllHints(err); len(hints) > 0 {
		event.Tags["user_facing"] = "yes"
		if len(event.Exception) > 0 && len(event.Exception[0].Value) > 0 {
			event.Exception[0].Value = strings.SplitN(hints[0], "\n", 1)[0]
		}
	} else {
		event.Tags["user_facing"] = "no"
	}

	// TODO(sdboyer) decide on how to populate the Level field

	for extraKey, extraValue := range extraDetails {
		event.Extra[extraKey] = extraValue
	}

	// why: Avoid leaking the machine's hostname by injecting the literal "<redacted>".
	// Otherwise, sentry.Client.Capture will see an empty ServerName field and
	// automatically fill in the machine's hostname.
	event.ServerName = "<redacted>"

	tags := make(map[string]string)
	visitAllMultiPostOrder(err, func(c error) {
		if t, ok := c.(SentryTagger); ok {
			for key, value := range t.ErrorTags() {
				tags[key] = value
			}
		}
	})

	event.Tags["report_type"] = "error"
	for k, v := range tags {
		if _, has := event.Tags[k]; !has {
			event.Tags[k] = v
		}
	}

	if opt != nil {
		for k, v := range opt.Tags {
			if _, has := event.Tags[k]; !has {
				event.Tags[k] = v
			}
		}
	}

	if opt != nil && opt.UserID != "" {
		event.User.ID = opt.UserID
	}

	res := sentry.CaptureEvent(event)
	if res != nil {
		return string(*res)
	}
	return ""
}

package errs

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/report"
	"github.com/getsentry/sentry-go"
)

const (
	// As with any analytics system, Sentry DSNs are a public endpoints that needn't be secret - worst case,
	// an attacker tries to undermine our remediation capabilities by injecting noise into our telemetry data.

	// DSN for the "main" sentry project. Everything goes in the same project, until we have a good reason to deviate.
	SentryMainDSN string = "https://a0546c06ff00cb18c7867c5783f96763@o4507623795523584.ingest.us.sentry.io/4507623799193600"
)

// ReportToSentry reports an error to sentry, populating the data in a standardized manner for all Nuon errors.
func ReportToSentry(err error) string {
	event, extraDetails := report.BuildSentryReport(err)

	if hints := errors.GetAllHints(err); len(hints) > 0 {
		// If hints have been provided, we want them specially included in sentry output
		event.Tags["user_facing"] = "yes"
		if len(event.Exception) > 0 && len(event.Exception[0].Value) > 0 {
			// Inject the first line of the human-facing string into the exception message
			event.Exception[0].Value = strings.SplitN(hints[0], "\n", 1)[0]

			// This injects the first line of hint text into the rest of the info already there. Might be preferable?
			// event.Exception[0].Value = strings.Replace(event.Exception[0].Value, ":", fmt.Sprintf(": %s", strings.SplitN(hints[0], "\n", 1)[0]), 1)
		}
	} else {
		event.Tags["user_facing"] = "no"
	}

	// TODO(sdboyer) decide on how to populate the Level field

	for extraKey, extraValue := range extraDetails {
		event.Extra[extraKey] = extraValue
	}

	// Avoid leaking the machine's hostname by injecting the literal "<redacted>".
	// Otherwise, sentry.Client.Capture will see an empty ServerName field and
	// automatically fill in the machine's hostname.
	event.ServerName = "<redacted>"

	tags := map[string]string{
		"report_type": "error",
	}
	for key, value := range tags {
		event.Tags[key] = value
	}

	res := sentry.CaptureEvent(event)
	if res != nil {
		return string(*res)
	}
	return ""
}

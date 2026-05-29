package client

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ResolveSiteURL maps a stored site value to its base API URL. Accepts
// either:
//
//   - A known regional key (us1, us3, us5, eu1, ap1, gov) → returns the
//     canonical DD API hostname for that region.
//   - A full https URL → returned verbatim (after trimming trailing /).
//
// Anything else is rejected upstream by app.validateDatadogSite; this
// function falls back to the us1 hostname only as a defensive default
// for already-persisted rows that somehow slip through validation. The
// model's BeforeSave invariant makes that branch effectively unreachable
// in practice.
//
// Keeping this in the client package rather than on the model lets the
// HTTP client be the single owner of "what hostname do I hit" — the model
// only cares about validation + storage shape.
func ResolveSiteURL(site string) string {
	site = strings.TrimSpace(site)
	if strings.HasPrefix(site, "https://") {
		return strings.TrimRight(site, "/")
	}
	switch site {
	case app.DatadogSiteUS1:
		return "https://api.datadoghq.com"
	case app.DatadogSiteUS3:
		return "https://api.us3.datadoghq.com"
	case app.DatadogSiteUS5:
		return "https://api.us5.datadoghq.com"
	case app.DatadogSiteEU1:
		return "https://api.datadoghq.eu"
	case app.DatadogSiteAP1:
		return "https://api.ap1.datadoghq.com"
	case app.DatadogSiteGov:
		return "https://api.ddog-gov.com"
	default:
		// Defensive fallback: should be impossible past validateDatadogSite.
		return "https://api.datadoghq.com"
	}
}

// AppURLForSite returns the DD web UI base URL for the same site. Used to
// deep-link from Nuon dashboards into the DD event stream / monitor page.
// The mapping is identical to ResolveSiteURL except the `api.` subdomain
// is dropped (DD uses app.datadoghq.com, app.us3.datadoghq.com, etc.).
func AppURLForSite(site string) string {
	site = strings.TrimSpace(site)
	if strings.HasPrefix(site, "https://") {
		// For custom URLs we have no app/api distinction — return the
		// URL verbatim and let the caller decide.
		return strings.TrimRight(site, "/")
	}
	switch site {
	case app.DatadogSiteUS1:
		return "https://app.datadoghq.com"
	case app.DatadogSiteUS3:
		return "https://app.us3.datadoghq.com"
	case app.DatadogSiteUS5:
		return "https://app.us5.datadoghq.com"
	case app.DatadogSiteEU1:
		return "https://app.datadoghq.eu"
	case app.DatadogSiteAP1:
		return "https://app.ap1.datadoghq.com"
	case app.DatadogSiteGov:
		return "https://app.ddog-gov.com"
	default:
		return "https://app.datadoghq.com"
	}
}

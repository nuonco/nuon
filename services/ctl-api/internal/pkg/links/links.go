package links

import (
	"context"
	"fmt"
	"net/url"
)

func buildUILink(ctx context.Context, pieces ...string) string {
	cfg := configFromContext(ctx)
	if cfg == nil {
		return ""
	}

	link, err := url.JoinPath(cfg.AppURL, pieces...)
	if err == nil {
		return link
	}

	return ""
}

func buildAPILink(ctx context.Context, pieces ...string) string {
	cfg := configFromContext(ctx)
	if cfg == nil {
		return ""
	}

	link, err := url.JoinPath(cfg.PublicAPIURL, pieces...)
	if err == nil {
		return link
	}

	return ""
}

func buildAdminAPILink(ctx context.Context, pieces ...string) string {
	cfg := configFromContext(ctx)
	if cfg == nil {
		return ""
	}

	link, err := url.JoinPath(cfg.PublicAPIURL, pieces...)
	if err == nil {
		return link
	}

	return ""
}

func eventLoopLink(ctx context.Context, namespace, id string) string {
	cfg := configFromContext(ctx)
	if cfg == nil {
		return ""
	}

	link, err := url.JoinPath(cfg.TemporalUIURL,
		"namespaces",
		namespace,
		"workflows",
		"event-loop-"+id)
	if err == nil {
		return link
	}

	return ""
}

func eventLoopSignalLink(ctx context.Context, namespace, id string, sig string) string {
	cfg := configFromContext(ctx)
	if cfg == nil {
		return ""
	}

	link, err := url.JoinPath(cfg.TemporalUIURL,
		"namespaces",
		namespace,
		"workflows",
		fmt.Sprintf("sig-%s-%s", id, sig),
	)
	if err == nil {
		return link
	}

	return ""
}

package links

import (
	"fmt"
	"net/url"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func eventLoopLink(cfg *internal.Config, namespace, id string) string {
	link, err := url.JoinPath(cfg.TemporalUIURL,
		"namespaces",
		namespace,
		"workflows",
		"event-loop-"+id)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

func eventLoopSignalLink(cfg *internal.Config, namespace, id string, sig string) string {
	link, err := url.JoinPath(cfg.TemporalUIURL,
		"namespaces",
		namespace,
		"workflows",
		fmt.Sprintf("sig-%s-%s", id, sig),
	)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

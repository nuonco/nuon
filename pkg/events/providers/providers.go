// Package providers registers the built-in inbound event providers. Plain
// providers are declarative provider.Base values; protocol-heavy providers
// (Slack, Azure Event Grid) override the default behavior they differ in.
package providers

import (
	"sort"

	"github.com/nuonco/nuon/pkg/events/envelope"
	"github.com/nuonco/nuon/pkg/events/provider"
)

var defaultProvider = provider.Base{}

var registry = func() map[string]provider.Provider {
	all := []provider.Provider{
		provider.Base{ProviderName: "github", ProviderDefaults: provider.Defaults{
			Auth:       provider.AuthHMAC,
			AuthConfig: provider.AuthConfig{Header: "X-Hub-Signature-256", Prefix: "sha256=", Algorithm: "sha256", Encoding: "hex"},
			TypeFrom:   envelope.FieldSelector{Header: "X-GitHub-Event"},
			IDFrom:     envelope.FieldSelector{Header: "X-GitHub-Delivery"},
		}},
		provider.Base{ProviderName: "gitlab", ProviderDefaults: provider.Defaults{
			Auth:       provider.AuthAPIKey,
			AuthConfig: provider.AuthConfig{Header: "X-Gitlab-Token"},
			TypeFrom:   envelope.FieldSelector{Header: "X-Gitlab-Event"},
			IDFrom:     envelope.FieldSelector{Header: "X-Gitlab-Event-UUID"},
		}},
		provider.Base{ProviderName: "bitbucket", ProviderDefaults: provider.Defaults{
			Auth:       provider.AuthHMAC,
			AuthConfig: provider.AuthConfig{Header: "X-Hub-Signature", Prefix: "sha256=", Algorithm: "sha256", Encoding: "hex"},
			TypeFrom:   envelope.FieldSelector{Header: "X-Event-Key"},
			IDFrom:     envelope.FieldSelector{Header: "X-Request-UUID"},
		}},
		provider.Base{ProviderName: "gitea", ProviderDefaults: provider.Defaults{
			Auth:       provider.AuthHMAC,
			AuthConfig: provider.AuthConfig{Header: "X-Gitea-Signature", Algorithm: "sha256", Encoding: "hex"},
			TypeFrom:   envelope.FieldSelector{Header: "X-Gitea-Event"},
			IDFrom:     envelope.FieldSelector{Header: "X-Gitea-Delivery"},
		}},
		provider.Base{ProviderName: "forgejo", ProviderDefaults: provider.Defaults{
			Auth:       provider.AuthHMAC,
			AuthConfig: provider.AuthConfig{Header: "X-Forgejo-Signature", Algorithm: "sha256", Encoding: "hex"},
			TypeFrom:   envelope.FieldSelector{Header: "X-Forgejo-Event"},
			IDFrom:     envelope.FieldSelector{Header: "X-Forgejo-Delivery"},
		}},
		provider.Base{ProviderName: "terraform-cloud", ProviderDefaults: provider.Defaults{
			Auth:       provider.AuthHMAC,
			AuthConfig: provider.AuthConfig{Header: "X-TFE-Notification-Signature", Algorithm: "sha512", Encoding: "hex"},
		}},
		provider.Base{ProviderName: "google-pubsub", ProviderDefaults: provider.Defaults{
			Auth:         provider.AuthBearerJWT,
			Envelope:     provider.EnvelopePubSubPush,
			AuthConfig:   provider.AuthConfig{Issuer: "https://accounts.google.com"},
			CallerFields: []provider.CallerField{provider.CallerFieldAudience, provider.CallerFieldExpectedEmail, provider.CallerFieldExpectedSubject},
		}},
		provider.Base{ProviderName: "azure-devops", ProviderDefaults: provider.Defaults{
			Auth:       provider.AuthBasic,
			AuthConfig: provider.AuthConfig{Username: "nuon"},
		}},
		provider.Base{ProviderName: "aws-eventbridge", ProviderDefaults: provider.Defaults{
			Auth:       provider.AuthAPIKey,
			AuthConfig: provider.AuthConfig{Header: "X-Nuon-API-Key"},
			TypeFrom:   envelope.FieldSelector{Payload: `$['detail-type']`},
			IDFrom:     envelope.FieldSelector{Payload: "$.id"},
		}},
		provider.Base{ProviderName: "aws-sns", ProviderDefaults: provider.Defaults{
			Auth:         provider.AuthSNSSignature,
			Envelope:     provider.EnvelopeSNS,
			CallerFields: []provider.CallerField{provider.CallerFieldTopicARN},
		}},
		azureEventGrid{Base: provider.Base{ProviderName: "azure-event-grid", ProviderDefaults: provider.Defaults{
			Auth:           provider.AuthAPIKey,
			AuthConfig:     provider.AuthConfig{Header: "X-Nuon-API-Key"},
			NativeProtocol: true,
		}}},
		slackEvents{Base: provider.Base{ProviderName: "slack-events", ProviderDefaults: provider.Defaults{
			Auth:           provider.AuthHMAC,
			AuthConfig:     provider.AuthConfig{Header: "X-Slack-Signature", Prefix: "v0=", Algorithm: "sha256", Encoding: "hex"},
			NativeProtocol: true,
			CallerSecret:   true,
		}}},
		provider.Base{ProviderName: "datadog", ProviderDefaults: provider.Defaults{
			Auth:       provider.AuthAPIKey,
			AuthConfig: provider.AuthConfig{Header: "X-Nuon-API-Key"},
			TypeFrom:   envelope.FieldSelector{Payload: "$.event_type"},
			IDFrom:     envelope.FieldSelector{Payload: "$.event_id"},
		}},
	}
	byName := make(map[string]provider.Provider, len(all))
	for _, p := range all {
		byName[p.Name()] = p
	}
	return byName
}()

// Lookup returns the provider registered under preset. It returns the default
// provider for the empty preset and reports false for unknown presets.
func Lookup(preset string) (provider.Provider, bool) {
	if preset == "" {
		return defaultProvider, true
	}
	p, ok := registry[preset]
	return p, ok
}

// Resolve returns the provider for preset, falling back to the default
// provider when the preset is empty or unknown.
func Resolve(preset string) provider.Provider {
	if p, ok := Lookup(preset); ok {
		return p
	}
	return defaultProvider
}

// Names returns the registered preset names in sorted order.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

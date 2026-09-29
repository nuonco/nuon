package orgs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/services/orgs/subscriptiontui"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

type SubscriptionFlags struct {
	JSON string
	File string
}

type SubscriptionPayload struct {
	Interests any                       `json:"interests"`
	Match     *labels.SubscriptionMatch `json:"match"`
}

func (f SubscriptionFlags) Resolve() (SubscriptionPayload, error) {
	jsonProvided := strings.TrimSpace(f.JSON) != ""
	fileProvided := strings.TrimSpace(f.File) != ""

	if jsonProvided && fileProvided {
		return SubscriptionPayload{}, fmt.Errorf("only one of --subscription-json, --subscription-file may be set")
	}

	if !jsonProvided && !fileProvided {
		return SubscriptionPayload{
			Interests: map[string]any{"all_events": true},
		}, nil
	}

	var raw []byte
	if fileProvided {
		b, err := os.ReadFile(strings.TrimSpace(f.File))
		if err != nil {
			return SubscriptionPayload{}, fmt.Errorf("read --subscription-file: %w", err)
		}
		raw = b
	} else {
		raw = []byte(f.JSON)
	}

	var payload SubscriptionPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return SubscriptionPayload{}, fmt.Errorf("parse subscription JSON: %w", err)
	}

	if payload.Interests == nil {
		payload.Interests = map[string]any{"all_events": true}
	}

	if payload.Match != nil {
		if err := payload.Match.Validate(); err != nil {
			return SubscriptionPayload{}, fmt.Errorf("invalid match: %w", err)
		}
	}

	return payload, nil
}

func resolveSubscription(ctx context.Context, api nuon.Client, interactive bool, f SubscriptionFlags) (SubscriptionPayload, error) {
	jsonProvided := strings.TrimSpace(f.JSON) != ""
	fileProvided := strings.TrimSpace(f.File) != ""
	if interactive && !jsonProvided && !fileProvided {
		interestsCfg, match, err := subscriptiontui.Run(ctx, api)
		if err != nil {
			return SubscriptionPayload{}, err
		}
		return SubscriptionPayload{Interests: interestsCfg, Match: match}, nil
	}
	return f.Resolve()
}

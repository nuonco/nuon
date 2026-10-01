package activities

import (
	"context"
	"encoding/json"
	"fmt"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (a *Activities) loadSourceArchive(ctx context.Context, appID, configID string) (*pkgconfig.SourceArchive, error) {
	var appCfg app.AppConfig
	res := a.db.WithContext(ctx).
		Where(app.AppConfig{AppID: appID}).
		First(&appCfg, "id = ?", configID)
	if res.Error != nil {
		return nil, fmt.Errorf("config not found: %w", res.Error)
	}

	if appCfg.SourceConfig == nil {
		return nil, nil
	}

	raw, err := appCfg.SourceConfig.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load source config: %w", err)
	}
	if raw == "" {
		return nil, nil
	}

	var archive pkgconfig.SourceArchive
	if err := json.Unmarshal([]byte(raw), &archive); err != nil {
		return nil, fmt.Errorf("unable to parse source config: %w", err)
	}
	if len(archive.Members) == 0 && len(archive.Files) > 0 {
		// Archives captured before the member index existed can be rebuilt
		// from file contents.
		if err := archive.ReindexMembers(); err != nil {
			return nil, fmt.Errorf("unable to index source members: %w", err)
		}
	}
	return &archive, nil
}

// computeSourceArchiveDiff returns nil (not an error) when either config has
// no stored source archive, e.g. configs created before capture existed. An
// empty baseConfigID (no baseline run) diffs against an empty archive so every
// file shows as added.
func (a *Activities) computeSourceArchiveDiff(ctx context.Context, appID, headConfigID, baseConfigID string) (*pkgconfig.SourceArchiveDiff, error) {
	head, err := a.loadSourceArchive(ctx, appID, headConfigID)
	if err != nil {
		return nil, err
	}
	if head == nil {
		return nil, nil
	}

	var base *pkgconfig.SourceArchive
	if baseConfigID != "" {
		base, err = a.loadSourceArchive(ctx, appID, baseConfigID)
		if err != nil {
			return nil, err
		}
	}
	if base == nil {
		base = pkgconfig.NewSourceArchive()
	}

	return head.Diff(base), nil
}

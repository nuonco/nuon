package dev

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/errs"
)

func (s *Service) getApp(ctx context.Context, dir string) (string, error) {
	appName, err := parse.AppNameFromDirName(dir)
	if err != nil {
		err = errs.WithUserFacing(err, "error parsing app name from file")
		return "", err
	}

	appID, err := lookup.AppID(ctx, s.api, appName)
	if err != nil {
		err = errs.WithUserFacing(err, "error looking up app id")
		return "", err
	}
	return appID, nil
}

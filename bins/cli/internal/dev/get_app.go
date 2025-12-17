package dev

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/pkg/errs"
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

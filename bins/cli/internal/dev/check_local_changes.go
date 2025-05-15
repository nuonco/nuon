package dev

import (
	"context"

	"github.com/pkg/errors"
)

func (s *Service) checkLocalChanges(ctx context.Context) error {
	changedFiles, err := checkUnpushedChanges()
	if err != nil {
		return err
	}
	if len(changedFiles) > 0 {
		if err := prompt(s.autoApprove, "You have local changes you haven't pushed. Are you sure you want to continue?"); err != nil {
			return errors.New("Confirmed. Stopping now.")
		}
	}
	return nil
}

package dev

import (
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/pkg/errors"
)

func prompt(autoApprove, interactive bool, msg string, vars ...any) error {
	if autoApprove || !interactive {
		return nil
	}

	promptText := fmt.Sprintf(msg, vars...)
	yes, err := bubbles.Confirm(promptText, interactive)
	if err != nil {
		return err
	}

	if !yes {
		return errors.New("Stopping now")
	}
	return nil
}

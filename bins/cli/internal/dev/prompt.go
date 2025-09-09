package dev

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
)

func prompt(autoApprove bool, msg string, vars ...any) error {
	if autoApprove {
		return nil
	}

	promptText := fmt.Sprintf(msg, vars...)
	yes, err := bubbles.Confirm(promptText)
	if err != nil {
		return err
	}
	
	if !yes {
		return errors.New("Stopping now")
	}
	return nil
}

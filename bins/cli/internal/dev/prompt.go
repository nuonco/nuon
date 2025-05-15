package dev

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pterm/pterm"
)

func prompt(autoApprove bool, msg string, vars ...any) error {
	if autoApprove {
		return nil
	}

	pterm.Println()
	yes, _ := pterm.DefaultInteractiveConfirm.Show(fmt.Sprintf(msg, vars...))
	pterm.Println()
	if !yes {
		return errors.New("Stopping now")
	}
	return nil
}

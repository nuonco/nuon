package dev

import (
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) notifyOrphanedComponents(cmps map[string]string) {
	if len(cmps) == 0 {
		return
	}

	msg := "Existing component(s) are no longer defined in the config:\n"

	for name, id := range cmps {
		msg += fmt.Sprintf("Component: Name=%s | ID=%s\n", name, id)
	}

	ui.PrintLn(msg)
}

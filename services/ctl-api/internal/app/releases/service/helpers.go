package service

import "github.com/powertoolsdev/mono/services/ctl-api/internal/app"

func installsToIDSlice(installs []app.Install) []string {
	installIDs := make([]string, len(installs))
	for idx, install := range installs {
		installIDs[idx] = install.ID
	}

	return installIDs
}

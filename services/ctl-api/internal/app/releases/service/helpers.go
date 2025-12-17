package service

import "github.com/nuonco/nuon/services/ctl-api/internal/app"

func installsToIDSlice(installs []app.Install) []string {
	installIDs := make([]string, len(installs))
	for idx, install := range installs {
		installIDs[idx] = install.ID
	}

	return installIDs
}

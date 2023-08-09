package app

import "fmt"

func (c *commands) ensureAppID() error {
	if c.appID == "" {
		return fmt.Errorf("Please set a nuon app ID using: $%s", APP_ID_VAR_NAME)
	}

	return nil
}

func (c *commands) ensureOrgID() error {
	if c.orgID == "" {
		return fmt.Errorf("Please set a nuon org ID using: $%s", ORG_ID_VAR_NAME)
	}

	return nil
}

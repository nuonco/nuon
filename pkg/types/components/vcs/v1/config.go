package vcsv1

func (c *Config) Directory() string {
	switch cfg := c.Cfg.(type) {
	case *Config_ConnectedGithubConfig:
		return cfg.ConnectedGithubConfig.Directory
	case *Config_PublicGitConfig:
		return cfg.PublicGitConfig.Directory
	}

	return ""
}

func (c *Config) GitRef() string {
	switch cfg := c.Cfg.(type) {
	case *Config_ConnectedGithubConfig:
		return cfg.ConnectedGithubConfig.GitRef
	case *Config_PublicGitConfig:
		return cfg.PublicGitConfig.GitRef
	}

	return ""
}

func (c *Config) Repo() string {
	switch cfg := c.Cfg.(type) {
	case *Config_ConnectedGithubConfig:
		return cfg.ConnectedGithubConfig.Repo
	case *Config_PublicGitConfig:
		return cfg.PublicGitConfig.Repo
	}

	return ""
}

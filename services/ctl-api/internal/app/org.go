package app

type Org struct {
	Model

	CreatedByID     string
	Name            string `gorm:"uniqueIndex"`
	Apps            []App  `faker:"-"`
	IsNew           bool   `gorm:"-:all"`
	GithubInstallID string
}

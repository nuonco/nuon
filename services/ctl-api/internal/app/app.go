package app

type App struct {
	Model

	CreatedByID string
	Name        string
	OrgID       string
	Org         Org         `faker:"-"`
	Components  []Component `faker:"-"`
	Installs    []Install   `faker:"-"`
}

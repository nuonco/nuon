package app

type Instance struct {
	Model

	InstallID string
	Install   Install `faker:"-"`

	ComponentID string
	Component   Component `faker:"-"`

	Deploys []*Deploy `faker:"-"`
}

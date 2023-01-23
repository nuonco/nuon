package plan

type Activities struct {
	planCreator planCreator
}

func NewActivities() *Activities {
	return &Activities{
		planCreator: &planCreatorImpl{},
	}
}

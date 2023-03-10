package repository

type Activities struct {
	repositoryCreator
}

func NewActivities() *Activities {
	return &Activities{
		repositoryCreator: &repositoryCreatorImpl{},
	}
}

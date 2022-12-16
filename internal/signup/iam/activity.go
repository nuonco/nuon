package iam

type Activities struct {
	iamPolicyCreator
	iamRoleAssumer
}

func NewActivities() *Activities {
	return &Activities{
		iamPolicyCreator: &iamPolicyCreatorImpl{},
		iamRoleAssumer:   &iamRoleAssumerImpl{},
	}
}

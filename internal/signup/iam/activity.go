package iam

type Activities struct {
	iamPolicyCreator
	iamRoleCreator
	iamRolePolicyAttachmentCreator

	iamRoleAssumer
}

func NewActivities() *Activities {
	return &Activities{
		iamPolicyCreator:               &iamPolicyCreatorImpl{},
		iamRoleCreator:                 &iamRoleCreatorImpl{},
		iamRoleAssumer:                 &iamRoleAssumerImpl{},
		iamRolePolicyAttachmentCreator: &iamRolePolicyAttachmentCreatorImpl{},
	}
}

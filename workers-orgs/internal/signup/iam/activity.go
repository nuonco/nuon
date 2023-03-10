package iam

import "github.com/go-playground/validator/v10"

type Activities struct {
	iamPolicyCreator               iamPolicyCreator
	iamRoleCreator                 iamRoleCreator
	iamRolePolicyAttachmentCreator iamRolePolicyAttachmentCreator

	validator *validator.Validate
}

func NewActivities() *Activities {
	return &Activities{
		iamPolicyCreator:               &iamPolicyCreatorImpl{},
		iamRoleCreator:                 &iamRoleCreatorImpl{},
		iamRolePolicyAttachmentCreator: &iamRolePolicyAttachmentCreatorImpl{},
		validator:                      validator.New(),
	}
}

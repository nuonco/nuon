package orgiam

import "github.com/go-playground/validator/v10"

type Activities struct {
	iamPolicyCreator iamPolicyCreator
	iamPolicyDeleter iamPolicyDeleter

	iamRoleCreator iamRoleCreator
	iamRoleDeleter iamRoleDeleter

	iamRolePolicyAttachmentCreator iamRolePolicyAttachmentCreator
	iamRolePolicyAttachmentDeleter iamRolePolicyAttachmentDeleter

	validator *validator.Validate
}

func NewActivities() *Activities {
	return &Activities{
		iamPolicyCreator:               &iamPolicyCreatorImpl{},
		iamRoleCreator:                 &iamRoleCreatorImpl{},
		iamRolePolicyAttachmentCreator: &iamRolePolicyAttachmentCreatorImpl{},

		iamRolePolicyAttachmentDeleter: &iamRolePolicyAttachmentDeleterImpl{},
		iamPolicyDeleter:               &iamPolicyDeleterImpl{},
		iamRoleDeleter:                 &iamRoleDeleterImpl{},

		validator: validator.New(),
	}
}

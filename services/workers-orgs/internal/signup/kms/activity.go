package kms

import "github.com/go-playground/validator/v10"

type Activities struct {
	kmsKeyCreator       kmsKeyCreator
	kmsKeyAliasCreator  kmsKeyAliasCreator
	kmsKeyPolicyCreator kmsKeyPolicyCreator

	validator *validator.Validate
}

func NewActivities() *Activities {
	return &Activities{
		kmsKeyCreator:       &kmsKeyCreatorImpl{},
		kmsKeyAliasCreator:  &kmsKeyAliasCreatorImpl{},
		kmsKeyPolicyCreator: &kmsKeyPolicyCreatorImpl{},
		validator:           validator.New(),
	}
}

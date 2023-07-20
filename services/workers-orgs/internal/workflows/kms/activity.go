package kms

import "github.com/go-playground/validator/v10"

type Activities struct {
	kmsKeyCreator       kmsKeyCreator
	kmsKeyGetter        kmsKeyGetter
	kmsKeyAliasCreator  kmsKeyAliasCreator
	kmsKeyPolicyCreator kmsKeyPolicyCreator

	validator *validator.Validate
}

func NewActivities() *Activities {
	return &Activities{
		kmsKeyCreator:       &kmsKeyCreatorImpl{},
		kmsKeyGetter:        &kmsKeyGetterImpl{},
		kmsKeyAliasCreator:  &kmsKeyAliasCreatorImpl{},
		kmsKeyPolicyCreator: &kmsKeyPolicyCreatorImpl{},
		validator:           validator.New(),
	}
}

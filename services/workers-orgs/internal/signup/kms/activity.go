package kms

import "github.com/go-playground/validator/v10"

type Activities struct {
	kmsKeyCreator       kmsKeyCreator
	kmsKeyPolicyCreator kmsKeyPolicyCreator

	validator *validator.Validate
}

func NewActivities() *Activities {
	return &Activities{
		kmsKeyCreator:       &kmsKeyCreatorImpl{},
		kmsKeyPolicyCreator: &kmsKeyPolicyCreatorImpl{},
		validator:           validator.New(),
	}
}

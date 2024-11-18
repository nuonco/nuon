package dev

import (
	"log"
	"os"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/api"
)

// This package contains some tooling to help us run runners locally, while mimicing a real environment.
//
// Notably, when run locally we a.) automatically look up a runner from the API and b.) execute it with credentials for
// an AWS IAM Role if applicable. This allows us to re-enact the credential/sts environment and ensures that we can
// easily run `nctl api seed` to automatically pick the most recent runner and process jobs.
type devver struct {
	runnerTyp      string
	runnerID       string
	runnerIDInput  string
	runnerAPIToken string

	apiClient api.Client
	v         *validator.Validate
}

func New(runnerIDInput string) (*devver, error) {
	v := validator.New()

	adminAPIURL := os.Getenv("INTERNAL_API_URL")
	apiClient, err := api.New(v,
		api.WithURL(adminAPIURL),
		api.WithAdminEmail("runner-local@serviceaccount.nuon.co"),
	)
	if err != nil {
		log.Fatal("unable to create admin api url for run-local")
	}

	return &devver{
		runnerIDInput:  runnerIDInput,
		runnerAPIToken: os.Getenv("RUNNER_API_TOKEN"),
		apiClient:      apiClient,
		v:              v,
	}, nil
}

package dev

import (
	"log"
	"os"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/api"
)

type devver struct {
	watchRunnerType string

	runnerType     string
	runnerID       string
	runnerAPIToken string

	apiClient api.Client
	v         *validator.Validate
}

func New(watchTyp string) (*devver, error) {
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
		watchRunnerType: watchTyp,
		runnerAPIToken:  os.Getenv("RUNNER_API_TOKEN"),
		apiClient:       apiClient,
		v:               v,
	}, nil
}

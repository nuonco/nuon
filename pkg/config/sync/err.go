package sync

import (
	"errors"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	stderrerr "github.com/nuonco/nuon/sdks/nuon-go/models"
)

type SyncInternalErr struct {
	Description string
	Err         error
}

func (s SyncInternalErr) Error() string {
	msg := fmt.Sprintf("error syncing - %s", s.Description)
	if s.Err != nil {
		msg = fmt.Sprintf("%s %s", msg, s.Err.Error())
	}

	return msg
}

type SyncErr struct {
	Resource    string
	Description string
}

func (s SyncErr) Error() string {
	return fmt.Sprintf("unable to sync %s - %s", s.Resource, s.Description)
}

type SyncAPIErr struct {
	Resource string
	Err      error
}

type apiErrPayloader interface {
	GetPayload() *stderrerr.StderrErrResponse
}

func (s SyncAPIErr) Error() string {
	err := s.Err
	for err != nil {
		if p, ok := err.(apiErrPayloader); ok {
			if payload := p.GetPayload(); payload != nil && payload.Description != "" {
				return fmt.Sprintf("unable to sync %s - %s", s.Resource, payload.Description)
			}
		}
		err = errors.Unwrap(err)
	}
	return fmt.Sprintf("unable to sync %s - %s", s.Resource, s.Err.Error())
}

func RejectDockerBuildComponentsForFeature(cfg *config.AppConfig) error {
	if cfg == nil {
		return nil
	}

	for _, comp := range cfg.Components {
		if comp == nil || comp.Type != config.DockerBuildComponentType {
			continue
		}

		return SyncErr{
			Resource: "app config",
			Description: fmt.Sprintf(
				"component %q uses docker_build, but docker_build components have been deprecated and are no longer supported. "+
					"Use a container_image component to reference a pre-built image instead.",
				comp.Name,
			),
		}
	}

	return nil
}

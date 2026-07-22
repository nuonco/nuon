package triggers

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type apiClient interface {
	ListTriggerEvents(context.Context, int, string) ([]*models.TriggerEventSummary, error)
	ListTriggerEventsPage(context.Context, int, string, string) (*models.TriggerEventPage, error)
	SearchTriggerEvents(context.Context, models.TriggerEventListQuery) (*models.TriggerEventPage, error)
	GetTriggerEvent(context.Context, string) (*models.TriggerEvent, error)
	GetTriggerEventRaw(context.Context, string) (*models.TriggerEventRaw, error)
	ReplayTriggerEvent(context.Context, string) (*models.TriggerEventReplayResponse, error)
	GetTriggerEventDispatch(context.Context, string) (*models.TriggerEventDispatch, error)
	RetryTriggerEventDispatch(context.Context, string) (*models.TriggerEventDispatchRetryResponse, error)
	ListTriggerEventDispatchesPage(context.Context, int, string, string) (*models.TriggerEventDispatchPage, error)
	CreateTrigger(context.Context, *models.TriggerCreateRequest) (*models.TriggerCredentialResponse, error)
	ListTriggers(context.Context) ([]*models.Trigger, error)
	GetTrigger(context.Context, string) (*models.Trigger, error)
	RotateTriggerSecret(context.Context, string) (*models.TriggerCredentialResponse, error)
	EnableTrigger(context.Context, string) (*models.Trigger, error)
	DisableTrigger(context.Context, string) (*models.Trigger, error)
	RevokeTriggerSecret(context.Context, string, string) (*models.TriggerRevokeResponse, error)
	RevealTriggerSecret(context.Context, string, string) (*models.TriggerSecretRevealResponse, error)
	GetTriggerIngressURL(context.Context, string) (*models.TriggerIngressURLResponse, error)
	RotateTriggerIngressURL(context.Context, string) (*models.TriggerCredentialResponse, error)
	DeleteTrigger(context.Context, string, bool) error
}

type Service struct {
	api apiClient
}

func New(api apiClient) *Service {
	return &Service{api: api}
}

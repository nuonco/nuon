package nuon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) ListTriggerEvents(ctx context.Context, limit int, trigger string) ([]*models.TriggerEventSummary, error) {
	page, err := c.ListTriggerEventsPage(ctx, limit, trigger, "")
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (c *client) ListTriggerEventsPage(ctx context.Context, limit int, trigger, cursor string) (*models.TriggerEventPage, error) {
	return c.SearchTriggerEvents(ctx, models.TriggerEventListQuery{Limit: limit, Trigger: trigger, Cursor: cursor})
}

func (c *client) SearchTriggerEvents(ctx context.Context, filters models.TriggerEventListQuery) (*models.TriggerEventPage, error) {
	if filters.Trigger == "" {
		return nil, fmt.Errorf("trigger is required to list trigger events")
	}
	query := url.Values{"limit": []string{fmt.Sprint(filters.Limit)}}
	for key, value := range map[string]string{"event_type": filters.EventType, "outcome": filters.Outcome, "query": filters.Search, "received_after": filters.ReceivedAfter, "received_before": filters.ReceivedBefore, "cursor": filters.Cursor} {
		if value != "" {
			query.Set(key, value)
		}
	}
	requestURL := fmt.Sprintf("%s/v1/triggers/%s/events?%s", c.APIURL, url.PathEscape(filters.Trigger), query.Encode())
	var page models.TriggerEventPage
	if err := c.triggerEventRequest(ctx, http.MethodGet, requestURL, http.StatusOK, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *client) GetTriggerEvent(ctx context.Context, id string) (*models.TriggerEvent, error) {
	requestURL := fmt.Sprintf("%s/v1/triggers/events/%s", c.APIURL, url.PathEscape(id))
	var event models.TriggerEvent
	if err := c.triggerEventRequest(ctx, http.MethodGet, requestURL, http.StatusOK, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func (c *client) GetTriggerEventRaw(ctx context.Context, id string) (*models.TriggerEventRaw, error) {
	requestURL := fmt.Sprintf("%s/v1/triggers/events/%s/raw", c.APIURL, url.PathEscape(id))
	var event models.TriggerEventRaw
	if err := c.triggerEventRequest(ctx, http.MethodGet, requestURL, http.StatusOK, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func (c *client) ReplayTriggerEvent(ctx context.Context, id string) (*models.TriggerEventReplayResponse, error) {
	requestURL := fmt.Sprintf("%s/v1/triggers/events/%s/replay", c.APIURL, url.PathEscape(id))
	var response models.TriggerEventReplayResponse
	if err := c.triggerEventRequest(ctx, http.MethodPost, requestURL, http.StatusAccepted, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *client) ListTriggerEventDispatches(ctx context.Context, limit int) ([]*models.TriggerEventDispatch, error) {
	page, err := c.ListTriggerEventDispatchesPage(ctx, limit, "", "")
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (c *client) ListTriggerEventDispatchesPage(ctx context.Context, limit int, eventID, cursor string) (*models.TriggerEventDispatchPage, error) {
	query := url.Values{"limit": []string{fmt.Sprint(limit)}}
	if eventID != "" {
		query.Set("event_id", eventID)
	}
	if cursor != "" {
		query.Set("cursor", cursor)
	}
	requestURL := fmt.Sprintf("%s/v1/triggers/dispatches?%s", c.APIURL, query.Encode())
	var page models.TriggerEventDispatchPage
	if err := c.triggerEventRequest(ctx, http.MethodGet, requestURL, http.StatusOK, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *client) GetTriggerEventDispatch(ctx context.Context, id string) (*models.TriggerEventDispatch, error) {
	var result models.TriggerEventDispatch
	err := c.triggerRequest(ctx, http.MethodGet, fmt.Sprintf("%s/v1/triggers/dispatches/%s", c.APIURL, url.PathEscape(id)), nil, http.StatusOK, &result)
	return &result, err
}

func (c *client) RetryTriggerEventDispatch(ctx context.Context, id string) (*models.TriggerEventDispatchRetryResponse, error) {
	var result models.TriggerEventDispatchRetryResponse
	err := c.triggerRequest(ctx, http.MethodPost, fmt.Sprintf("%s/v1/triggers/dispatches/%s/retry", c.APIURL, url.PathEscape(id)), nil, http.StatusAccepted, &result)
	return &result, err
}

func (c *client) CreateTrigger(ctx context.Context, body *models.TriggerCreateRequest) (*models.TriggerCredentialResponse, error) {
	var result models.TriggerCredentialResponse
	err := c.triggerRequest(ctx, http.MethodPost, c.APIURL+"/v1/triggers", body, http.StatusCreated, &result)
	return &result, err
}

func (c *client) ListTriggers(ctx context.Context) ([]*models.Trigger, error) {
	var result []*models.Trigger
	err := c.triggerRequest(ctx, http.MethodGet, c.APIURL+"/v1/triggers", nil, http.StatusOK, &result)
	return result, err
}

func (c *client) GetTrigger(ctx context.Context, id string) (*models.Trigger, error) {
	return c.triggerAction(ctx, id, "", http.MethodGet)
}
func (c *client) EnableTrigger(ctx context.Context, id string) (*models.Trigger, error) {
	return c.triggerAction(ctx, id, "enable", http.MethodPost)
}
func (c *client) DisableTrigger(ctx context.Context, id string) (*models.Trigger, error) {
	return c.triggerAction(ctx, id, "disable", http.MethodPost)
}
func (c *client) triggerAction(ctx context.Context, id, action, method string) (*models.Trigger, error) {
	path := fmt.Sprintf("%s/v1/triggers/%s", c.APIURL, url.PathEscape(id))
	if action != "" {
		path += "/" + action
	}
	var result models.Trigger
	err := c.triggerRequest(ctx, method, path, nil, http.StatusOK, &result)
	return &result, err
}
func (c *client) RotateTriggerSecret(ctx context.Context, id string) (*models.TriggerCredentialResponse, error) {
	var result models.TriggerCredentialResponse
	err := c.triggerRequest(ctx, http.MethodPost, fmt.Sprintf("%s/v1/triggers/%s/rotate-secret", c.APIURL, url.PathEscape(id)), nil, http.StatusCreated, &result)
	return &result, err
}
func (c *client) RevokeTriggerSecret(ctx context.Context, triggerID, secretID string) (*models.TriggerRevokeResponse, error) {
	var result models.TriggerRevokeResponse
	path := fmt.Sprintf("%s/v1/triggers/%s/secrets/%s/revoke", c.APIURL, url.PathEscape(triggerID), url.PathEscape(secretID))
	err := c.triggerRequest(ctx, http.MethodPost, path, nil, http.StatusOK, &result)
	return &result, err
}
func (c *client) RevealTriggerSecret(ctx context.Context, triggerID, secretID string) (*models.TriggerSecretRevealResponse, error) {
	var result models.TriggerSecretRevealResponse
	path := fmt.Sprintf("%s/v1/triggers/%s/secrets/%s/reveal", c.APIURL, url.PathEscape(triggerID), url.PathEscape(secretID))
	err := c.triggerRequest(ctx, http.MethodPatch, path, nil, http.StatusOK, &result)
	return &result, err
}
func (c *client) GetTriggerIngressURL(ctx context.Context, id string) (*models.TriggerIngressURLResponse, error) {
	var result models.TriggerIngressURLResponse
	err := c.triggerRequest(ctx, http.MethodPatch, fmt.Sprintf("%s/v1/triggers/%s/ingress-url", c.APIURL, url.PathEscape(id)), nil, http.StatusOK, &result)
	return &result, err
}
func (c *client) RotateTriggerIngressURL(ctx context.Context, id string) (*models.TriggerCredentialResponse, error) {
	var result models.TriggerCredentialResponse
	err := c.triggerRequest(ctx, http.MethodPost, fmt.Sprintf("%s/v1/triggers/%s/rotate-ingress-url", c.APIURL, url.PathEscape(id)), nil, http.StatusOK, &result)
	return &result, err
}
func (c *client) DeleteTrigger(ctx context.Context, id string, force bool) error {
	path := fmt.Sprintf("%s/v1/triggers/%s?force=%t", c.APIURL, url.PathEscape(id), force)
	return c.triggerRequest(ctx, http.MethodDelete, path, nil, http.StatusNoContent, nil)
}

func (c *client) triggerEventRequest(ctx context.Context, method, requestURL string, successStatus int, dst any) error {
	return c.triggerRequest(ctx, method, requestURL, nil, successStatus, dst)
}

func (c *client) triggerRequest(ctx context.Context, method, requestURL string, body any, successStatus int, dst any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to encode request: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL, reader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := (&http.Client{Transport: c.appTransport}).Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != successStatus {
		body, _ := io.ReadAll(resp.Body)
		return newHTTPAPIError(resp.StatusCode, string(body))
	}
	if successStatus == http.StatusNoContent || dst == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

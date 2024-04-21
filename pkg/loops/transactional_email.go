package loops

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	transactionalEmailSendURI string = "/api/v1/transactional"
)

type transactionalEmailRequest struct {
	Email           string                 `json:"email"`
	TransactionalID string                 `json:"transactionalId"`
	DataVariables   map[string]interface{} `json:"dataVariables"`
}

type transactionalEmailResponse struct {
	Success bool `json:"success"`
}

func (c *client) SendTransactionalEmail(ctx context.Context, email, transactionalEmailID string, vars map[string]interface{}) error {
	byts, err := json.Marshal(transactionalEmailRequest{
		Email:           email,
		TransactionalID: transactionalEmailID,
		DataVariables:   vars,
	})
	if err != nil {
		return fmt.Errorf("unable to create request json: %w", err)
	}

	responseByts, err := c.postRequest(ctx, transactionalEmailSendURI, byts)
	if err != nil {
		return fmt.Errorf("unable to send email: %w", err)
	}

	var resp transactionalEmailResponse
	if err := json.Unmarshal(responseByts, &resp); err != nil {
		return fmt.Errorf("unable to check response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("response was not successful")
	}

	return nil
}

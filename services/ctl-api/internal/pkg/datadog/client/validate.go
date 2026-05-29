package client

import "context"

// ValidateResponse is the shape of DD's /api/v1/validate response.
type ValidateResponse struct {
	Valid bool `json:"valid"`
}

// Validate probes whether an API key is accepted by DD's auth layer.
// Returns (valid=true, nil) only on a 200 with valid:true. Any other
// outcome returns false plus the underlying error so callers can
// distinguish a bad key (APIError, status=403) from a network issue.
//
// Used by the Test button on the connection form so a user can verify
// their key before persisting it.
func (c *Client) Validate(ctx context.Context, baseURL, apiKey string) (bool, error) {
	var resp ValidateResponse
	if err := c.doJSON(ctx, "GET", baseURL, "/api/v1/validate", apiKey, "", nil, &resp); err != nil {
		return false, err
	}
	return resp.Valid, nil
}

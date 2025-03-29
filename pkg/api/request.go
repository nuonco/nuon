package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (c *client) execGetRequest(ctx context.Context, endpoint string) ([]byte, error) {
	return c.execRequest(ctx, http.MethodGet, endpoint, nil)
}

func (c *client) execPostRequest(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
	return c.execRequest(ctx, http.MethodPost, endpoint, data)
}

func (c *client) execPatchRequest(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
	return c.execRequest(ctx, http.MethodPatch, endpoint, data)
}

func (c *client) execRequest(ctx context.Context, method string, endpoint string, data interface{}) ([]byte, error) {
	httpClient := http.Client{
		Timeout: c.Timeout,
	}

	dataByts, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal to json: %w", err)
	}

	url := c.APIURL + endpoint
	req, err := http.NewRequest(method, url, bytes.NewBuffer(dataByts))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}
	req.Header.Add("X-Nuon-Admin-Email", c.AdminEmail)

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
	}
	if !generics.SliceContains(res.StatusCode, []int{http.StatusCreated, http.StatusOK, http.StatusAccepted}) {
		return nil, fmt.Errorf("invalid response status: %d", res.StatusCode)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body: %w", err)
	}

	return body, nil
}

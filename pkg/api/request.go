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
	httpClient := http.Client{
		Timeout: c.Timeout,
	}

	url := c.APIURL + endpoint
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("X-Nuon-Admin-Email", c.AdminEmail)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
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

func (c *client) execPostRequest(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
	httpClient := http.Client{
		Timeout: c.Timeout,
	}

	dataByts, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal to json: %w", err)
	}

	url := c.APIURL + endpoint
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(dataByts))
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

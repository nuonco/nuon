package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
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

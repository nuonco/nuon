package loops

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	loopsAPIBaseURL   string        = "https://app.loops.so"
	defaultAPITimeout time.Duration = time.Second * 2
)

func (c *client) postRequest(ctx context.Context, uri string, byts []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", loopsAPIBaseURL, strings.TrimPrefix(uri, "/"))

	payload := bytes.NewReader(byts)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.APIKey)
	req.Header.Add("Content-Type", "application/json")

	timeoutCtx, cancelFn := context.WithTimeout(ctx, defaultAPITimeout)
	defer cancelFn()
	req = req.WithContext(timeoutCtx)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("response error: %w", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

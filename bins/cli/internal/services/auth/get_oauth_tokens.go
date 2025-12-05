package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TokenResp struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
}

func (a *Service) getOAuthTokens(deviceCode string) (TokenResp, error) {
	tokens := TokenResp{}
	reqURL := AuthDomain + "oauth/token"
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Add("device_code", deviceCode)
	data.Add("client_id", AuthClientID)

	authenticated := false
	for authenticated == false {
		req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(data.Encode()))
		if err != nil {
			return tokens, fmt.Errorf("couldn't create oauth token request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return tokens, fmt.Errorf("oauth token request failed: %w", err)
		}
		defer res.Body.Close()

		err = json.NewDecoder(res.Body).Decode(&tokens)
		if err != nil {
			return tokens, fmt.Errorf("couldn't decode token response: %w", err)
		}

		if res.StatusCode == http.StatusOK {
			authenticated = true
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	return tokens, nil
}

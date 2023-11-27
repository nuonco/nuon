package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/browser"
)

type DeviceCodeRes struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_uri_complete"`
}

func (a *Service) getDeviceCode() (string, error) {
	reqURL := AuthDomain + "oauth/device/code/"
	data := url.Values{}
	data.Set("client_id", AuthClientID)
	data.Add("scope", "openid email profile")
	data.Add("audience", AuthAudience)

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("couldn't create a request for the device code: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("device code request failed: %w", err)
	}
	defer res.Body.Close()

	deviceCodeData := DeviceCodeRes{}
	err = json.NewDecoder(res.Body).Decode(&deviceCodeData)
	if err != nil {
		return "", fmt.Errorf("couldn't decode device code response data: %w", err)
	}

	fmt.Println("Logging in to Nuon")
	fmt.Printf("Attempting to open the SSO authorization page in your default browser.\nIf the browser does not open or you wish to use a different device to\nauthorize this request, open the following URL:\n\n%s\n\nThen enter the code:\n\n%s\n\n", deviceCodeData.VerificationURL, deviceCodeData.UserCode)

	browser.OpenURL(deviceCodeData.VerificationURL)

	return deviceCodeData.DeviceCode, nil
}

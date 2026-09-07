package nuonjwtauthextension

import "go.opentelemetry.io/collector/client"

const principalAttribute = "nuon.telemetry.principal"

type Principal struct {
	OrgID     string
	AppID     string
	InstallID string
	RunnerID  string
}

type AuthData struct {
	principal Principal
}

var _ client.AuthData = (*AuthData)(nil)

func NewAuthData(principal Principal) *AuthData {
	return &AuthData{principal: principal}
}

func (a *AuthData) Principal() Principal {
	return a.principal
}

func (a *AuthData) GetAttribute(name string) any {
	if name == principalAttribute {
		return a.principal
	}
	return nil
}

func (*AuthData) GetAttributeNames() []string {
	return []string{principalAttribute}
}

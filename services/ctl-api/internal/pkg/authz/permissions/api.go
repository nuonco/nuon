package permissions

import "fmt"

const APIPermissionKey = "api"

type API string

const (
	APICtl    API = "ctl-api"
	APIRunner API = "runner"
	APIStack  API = "stack"
)

func ParseAPI(val string) (API, error) {
	switch API(val) {
	case APICtl, APIRunner, APIStack:
		return API(val), nil
	default:
		return "", fmt.Errorf("invalid api %s", val)
	}
}

func APIFromPermissions(perms map[string]*string) (API, bool) {
	raw, ok := perms[APIPermissionKey]
	if !ok || raw == nil || *raw == "" {
		return "", false
	}
	api, err := ParseAPI(*raw)
	if err != nil {
		return "", false
	}
	return api, true
}

package plantypes

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strings"

	assumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

const apiTokenEnvVar = "NUON_API_TOKEN"

var (
	ociRegistryAuthType = reflect.TypeOf(configs.OCIRegistryAuth{})
	awsStaticCredsType  = reflect.TypeOf(awscredentials.StaticCredentials{})
	iamStaticCredsType  = reflect.TypeOf(assumerole.StaticCredentials{})
	gitSourceType       = reflect.TypeOf(GitSource{})
)

// RedactCredentials blanks registry and cloud secrets so the plan can be shown
// to org members. The runner needs the unredacted plan, so only call this on
// copies served outside the runner API.
func (cp *CompositePlan) RedactCredentials() {
	if cp == nil {
		return
	}
	redactValue(reflect.ValueOf(cp).Elem())
}

func redactValue(v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			redactValue(v.Elem())
		}
	case reflect.Map:
		redactStringMap(v)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			redactValue(v.Index(i))
		}
	case reflect.Struct:
		switch v.Type() {
		case ociRegistryAuthType:
			v.FieldByName("Password").SetString("")
			return
		case awsStaticCredsType, iamStaticCredsType:
			v.FieldByName("SecretAccessKey").SetString("")
			v.FieldByName("SessionToken").SetString("")
			return
		case gitSourceType:
			field := v.FieldByName("URL")
			field.SetString(stripURLCredentials(field.String()))
			return
		}
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				redactValue(v.Field(i))
			}
		}
	}
}

// RedactPlanJSON blanks the same secrets as RedactCredentials in a raw plan
// JSON document. The input is returned unchanged when there is nothing to
// redact, so the response stays byte-identical for plans without secrets.
func RedactPlanJSON(raw string) (string, error) {
	if raw == "" {
		return raw, nil
	}

	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return "", err
	}
	if !redactJSON(doc) {
		return raw, nil
	}

	byts, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	return string(byts), nil
}

// The keys mirror the JSON encoding of OCIRegistryRepository.OCIAuth (no json
// tags), awscredentials.Config.Static, assumerole.TwoStepConfig, and GitSource.
func redactJSON(v any) bool {
	changed := false
	switch t := v.(type) {
	case map[string]any:
		changed = blankAPIToken(t)
		for k, child := range t {
			if obj, ok := child.(map[string]any); ok {
				switch k {
				case "OCIAuth":
					changed = blankStrings(obj, "Password") || changed
				case "static", "src_static_credentials":
					changed = blankStrings(obj, "secret_access_key", "session_token") || changed
				case "git_source":
					if raw, ok := obj["url"].(string); ok {
						if stripped := stripURLCredentials(raw); stripped != raw {
							obj["url"] = stripped
							changed = true
						}
					}
				}
			}
			changed = redactJSON(child) || changed
		}
	case []any:
		for _, child := range t {
			changed = redactJSON(child) || changed
		}
	}
	return changed
}

func redactStringMap(v reflect.Value) {
	if v.IsNil() || v.Type().Key().Kind() != reflect.String || v.Type().Elem().Kind() != reflect.String {
		return
	}
	key := reflect.ValueOf(apiTokenEnvVar)
	val := v.MapIndex(key)
	if !val.IsValid() || val.String() == "" {
		return
	}
	v.SetMapIndex(key, reflect.ValueOf(""))
}

func blankAPIToken(obj map[string]any) bool {
	s, ok := obj[apiTokenEnvVar].(string)
	if !ok || s == "" {
		return false
	}
	obj[apiTokenEnvVar] = ""
	return true
}

func stripURLCredentials(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}
	u.User = nil
	return u.String()
}

func blankStrings(obj map[string]any, keys ...string) bool {
	changed := false
	for _, k := range keys {
		if s, ok := obj[k].(string); ok && s != "" {
			obj[k] = ""
			changed = true
		}
	}
	return changed
}

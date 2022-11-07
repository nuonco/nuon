package components

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testUseBlockHappy struct{}

func (ub *testUseBlockHappy) ToJSON() ([]byte, error) {
	return []byte(`useblock`), nil
}

type testUseBlockSad struct{}

func (ub *testUseBlockSad) ToJSON() ([]byte, error) {
	return nil, errors.New("json error")
}

func TestToJSON(t *testing.T) {
	tests := map[string]struct {
		app         *App
		expected    []byte
		errExpected string
	}{
		"happy path": {
			app: &App{
				Name: "app-name",
				Build: &UseBlock{
					Use: &testUseBlockHappy{},
				},
				Deploy: &UseBlock{
					Use: &testUseBlockHappy{},
				},
			},
			expected: []byte(`{ "app-name": { "use": {"useblock": {}}, "use": {"useblock": {}}}}`),
		},
		"sad path build": {
			app: &App{
				Name:   "app-name",
				Build:  &UseBlock{Use: &testUseBlockSad{}},
				Deploy: &UseBlock{Use: &testUseBlockHappy{}},
			},
			errExpected: "json error",
		},
		"sad path deploy": {
			app: &App{
				Name:   "app-name",
				Build:  &UseBlock{Use: &testUseBlockHappy{}},
				Deploy: &UseBlock{Use: &testUseBlockSad{}},
			},
			errExpected: "json error",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := test.app.ToJSON()
			if test.errExpected != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected)
			} else {
				assert.NoError(t, err)
				buffer := new(bytes.Buffer)
				// use json.Compact because newlines and tabs are a pain
				assert.Equal(t, json.Compact(buffer, test.expected), json.Compact(buffer, actual))
			}
		})
	}
}

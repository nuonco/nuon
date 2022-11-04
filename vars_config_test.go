package terraform

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_varsConfigurer_createVarsConfigFile(t *testing.T) {
	errUnableToWrite := fmt.Errorf("unableToWrite")
	vars := map[string]interface{}{
		"var": "value",
	}

	tests := map[string]struct {
		workspaceFn func() workspaceFileWriter
		errExpected error
		assertFn    func(*testing.T, workspaceFileWriter)
	}{
		"happy path": {
			workspaceFn: func() workspaceFileWriter {
				wkspace := &testWorkspaceFileWriter{}
				wkspace.On("writeFile", mock.Anything, mock.Anything).Return(nil)
				return wkspace
			},
			errExpected: nil,
			assertFn: func(t *testing.T, wkspace workspaceFileWriter) {
				obj := wkspace.(*testWorkspaceFileWriter)

				obj.AssertNumberOfCalls(t, "writeFile", 1)
				filename := obj.Calls[0].Arguments[0].(string)
				assert.Equal(t, varsConfigFilename, filename)

				byts := obj.Calls[0].Arguments[1].([]byte)
				expectedByts, err := json.Marshal(vars)
				assert.Nil(t, err)
				assert.Equal(t, string(expectedByts), string(byts))
			},
		},
		"error": {
			workspaceFn: func() workspaceFileWriter {
				wkspace := &testWorkspaceFileWriter{}
				wkspace.On("writeFile", mock.Anything, mock.Anything).Return(errUnableToWrite)
				return wkspace
			},
			errExpected: errUnableToWrite,
			assertFn: func(t *testing.T, wkspace workspaceFileWriter) {
				obj := wkspace.(*testWorkspaceFileWriter)

				obj.AssertNumberOfCalls(t, "writeFile", 1)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			wkspace := test.workspaceFn()

			tr := &tfVarsConfigurer{}
			err := tr.createVarsConfigFile(vars, wkspace)

			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
			} else {
				assert.Nil(t, err)
			}
			test.assertFn(t, wkspace)
		})
	}
}

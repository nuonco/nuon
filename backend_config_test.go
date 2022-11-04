package terraform

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testWorkspaceFileWriter struct {
	mock.Mock
}

func (t *testWorkspaceFileWriter) writeFile(filename string, byts []byte) error {
	args := t.Called(filename, byts)
	return args.Error(0)
}

func Test_s3BackendConfigurer_createBackendConfig(t *testing.T) {
	errUnableToWrite := fmt.Errorf("unableToWrite")

	tests := map[string]struct {
		workspaceFn func() workspaceFileWriter
		cfg         BackendConfig
		errExpected error
		assertFn    func(*testing.T, workspaceFileWriter, BackendConfig)
	}{
		"happy path": {
			workspaceFn: func() workspaceFileWriter {
				wkspace := &testWorkspaceFileWriter{}
				wkspace.On("writeFile", mock.Anything, mock.Anything).Return(nil)
				return wkspace
			},
			cfg: BackendConfig{
				BucketName:   "nuon-installations",
				BucketRegion: "us-west-2",
				BucketKey:    "installations/id/state.tf",
			},
			errExpected: nil,
			assertFn: func(t *testing.T, wkspace workspaceFileWriter, cfg BackendConfig) {
				obj := wkspace.(*testWorkspaceFileWriter)

				obj.AssertNumberOfCalls(t, "writeFile", 1)
				filename := obj.Calls[0].Arguments[0].(string)
				assert.Equal(t, backendConfigFilename, filename)

				byts := obj.Calls[0].Arguments[1].([]byte)
				expectedByts, err := json.Marshal(cfg)
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
			cfg: BackendConfig{
				BucketName:   "nuon-installations",
				BucketRegion: "us-west-2",
				BucketKey:    "installations/id/state.tf",
			},
			errExpected: errUnableToWrite,
			assertFn: func(t *testing.T, wkspace workspaceFileWriter, cfg BackendConfig) {
				obj := wkspace.(*testWorkspaceFileWriter)

				obj.AssertNumberOfCalls(t, "writeFile", 1)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			wkspace := test.workspaceFn()

			tr := &s3BackendConfigurer{}
			err := tr.createBackendConfig(test.cfg, wkspace)

			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
			} else {
				assert.Nil(t, err)
			}
			test.assertFn(t, wkspace, test.cfg)
		})
	}
}

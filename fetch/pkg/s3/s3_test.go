package s3

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	t.Parallel()
	v := validator.New()

	tests := map[string]struct {
		v           *validator.Validate
		opts        []s3Option
		errExpected error
	}{
		"valid": {
			v: v,
			opts: []s3Option{
				WithBucketName("valid"),
				WithBucketKey("valid"),
				WithRoleARN("arn:aws:something"),
				WithRoleSessionName("test-assume-role"),
			},
		},
		"missing bucket name": {
			v: v,
			opts: []s3Option{
				WithBucketKey("valid"),
				WithRoleARN("arn:aws:something"),
				WithRoleSessionName("test-assume-role"),
			},
			errExpected: fmt.Errorf("Error:Field validation for 'BucketName' failed on the 'required' tag"),
		},
		"missing bucket key": {
			v: v,
			opts: []s3Option{
				WithBucketName("valid"),
				WithRoleARN("arn:aws:something"),
				WithRoleSessionName("test-assume-role"),
			},
			errExpected: fmt.Errorf("Error:Field validation for 'Key' failed on the 'required' tag"),
		},
		"missing role arn": {
			v: v,
			opts: []s3Option{
				WithBucketName("valid"),
				WithBucketKey("valid"),
				WithRoleSessionName("test-assume-role"),
			},
			errExpected: fmt.Errorf("Error:Field validation for 'RoleARN' failed on the 'required' tag"),
		},
		"missing role session name": {
			v: v,
			opts: []s3Option{
				WithBucketName("valid"),
				WithBucketKey("valid"),
				WithRoleARN("arn:aws:something"),
			},
			errExpected: fmt.Errorf("Error:Field validation for 'RoleSessionName' failed on the 'required' tag"),
		},

		"missing validator": {
			v:           nil,
			errExpected: fmt.Errorf("validator is nil"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f, err := New(test.v, test.opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, f)
		})
	}
}

type mockFetcher struct{ mock.Mock }

var _ fetcher = (*mockFetcher)(nil)

func (m *mockFetcher) fetch(ctx context.Context, api s3ObjectGetter) (io.ReadCloser, error) {
	args := m.Called(ctx, api)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

type mockS3ObjectGetter struct{ mock.Mock }

func (m *mockS3ObjectGetter) GetObject(
	ctx context.Context,
	in *s3.GetObjectInput,
	opts ...func(*s3.Options),
) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, in, opts)
	err := args.Error(1)
	maybeRet := args.Get(0)
	if maybeRet != nil {
		return maybeRet.(*s3.GetObjectOutput), err
	}
	return nil, err
}

var _ s3ObjectGetter = (*mockS3ObjectGetter)(nil)

var errNotFound = fmt.Errorf("not found")

func TestS3Fetcher_fetch(t *testing.T) {
	t.Parallel()
	s := &s3Fetcher{
		BucketName: "nuon-test-modules",
		Key:        "sandboxes/foobar/v0.0.0.1",
	}

	tests := map[string]struct {
		api         func(*s3Fetcher) *mockS3ObjectGetter
		expected    string
		errExpected error
	}{
		"object not found": {
			api: func(s *s3Fetcher) *mockS3ObjectGetter {
				in := &s3.GetObjectInput{Bucket: &s.BucketName, Key: &s.Key}
				m := &mockS3ObjectGetter{}
				m.
					On("GetObject", mock.Anything, in, ([]func(*s3.Options))(nil)).
					Return(nil, errNotFound)
				return m
			},
			errExpected: errNotFound,
		},

		"successfully returns valid readcloser": {
			api: func(s *s3Fetcher) *mockS3ObjectGetter {
				in := &s3.GetObjectInput{Bucket: &s.BucketName, Key: &s.Key}
				out := &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader("abc"))}
				m := &mockS3ObjectGetter{}
				m.On("GetObject", mock.Anything, in, ([]func(*s3.Options))(nil)).Return(out, nil)
				return m
			},
			expected: "abc",
		},
	}

	for name, test := range tests {
		name := name
		test := test
		ctx := context.Background()
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			test := test
			api := test.api(s)
			resp, err := s.fetch(ctx, api)
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
				return
			}
			assert.NoError(t, err)

			bs, err := io.ReadAll(resp)
			assert.NoError(t, err)

			assert.Equal(t, test.expected, string(bs))
			api.AssertExpectations(t)
		})
	}
}

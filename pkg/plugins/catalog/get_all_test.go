package catalog

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	ecrpublic_types "github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func Test_catalog_getAll(t *testing.T) {
	cat := generics.GetFakeObj[*catalog]()
	cat.DevOverride = false
	pluginTyp := PluginTypeTerraform
	imageTag := generics.GetFakeObj[ecrpublic_types.ImageTagDetail]()
	errGetLatest := fmt.Errorf("error getting latest")

	tests := map[string]struct {
		clientFn    func(*gomock.Controller) ecrpublicClient
		assertFn    func(*testing.T, []*Plugin)
		errExpected error
	}{
		"happy path - no second page": {
			clientFn: func(mockCtl *gomock.Controller) ecrpublicClient {
				mock := NewMockecrpublicClient(mockCtl)
				mock.EXPECT().DescribeImageTags(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *ecrpublic.DescribeImageTagsInput, opts ...ecrpublic.Options) (*ecrpublic.DescribeImageTagsOutput, error) {
						assert.Equal(t, pluginTyp.RepositoryName(), *req.RepositoryName)
						return &ecrpublic.DescribeImageTagsOutput{
							ImageTagDetails: []ecrpublic_types.ImageTagDetail{
								imageTag,
							},
						}, nil
					})
				return mock
			},
			assertFn: func(t *testing.T, plugins []*Plugin) {
				plugin := plugins[0]

				assert.Equal(t, *imageTag.ImageTag, plugin.Tag)
				assert.Equal(t, *imageTag.CreatedAt, plugin.CreatedAt)
				assert.Equal(t, pluginTyp.ImageURL(), plugin.ImageURL)
				assert.Equal(t, pluginTyp.RepositoryName(), plugin.RepositoryName)
			},
		},
		"happy path - multiple pages": {
			clientFn: func(mockCtl *gomock.Controller) ecrpublicClient {
				mock := NewMockecrpublicClient(mockCtl)
				nextToken := generics.GetFakeObj[string]()

				// first call has the actual items
				mock.EXPECT().DescribeImageTags(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *ecrpublic.DescribeImageTagsInput, opts ...ecrpublic.Options) (*ecrpublic.DescribeImageTagsOutput, error) {
						assert.Equal(t, pluginTyp.RepositoryName(), *req.RepositoryName)
						return &ecrpublic.DescribeImageTagsOutput{
							NextToken: generics.ToPtr(nextToken),
							ImageTagDetails: []ecrpublic_types.ImageTagDetail{
								imageTag,
							},
						}, nil
					})

				mock.EXPECT().DescribeImageTags(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *ecrpublic.DescribeImageTagsInput, opts ...ecrpublic.Options) (*ecrpublic.DescribeImageTagsOutput, error) {
						assert.Equal(t, pluginTyp.RepositoryName(), *req.RepositoryName)
						assert.Equal(t, nextToken, *req.NextToken)

						secondImageTag := generics.GetFakeObj[ecrpublic_types.ImageTagDetail]()
						ts := imageTag.CreatedAt.Add(-time.Minute)
						secondImageTag.CreatedAt = &ts
						return &ecrpublic.DescribeImageTagsOutput{
							ImageTagDetails: []ecrpublic_types.ImageTagDetail{
								secondImageTag,
							},
						}, nil
					})
				return mock
			},
			assertFn: func(t *testing.T, plugins []*Plugin) {
				plugin := plugins[0]

				assert.Equal(t, *imageTag.ImageTag, plugin.Tag)
				assert.Equal(t, *imageTag.CreatedAt, plugin.CreatedAt)
				assert.Equal(t, pluginTyp.ImageURL(), plugin.ImageURL)
				assert.Equal(t, pluginTyp.RepositoryName(), plugin.RepositoryName)
			},
		},
		"happy path - multiple pages with out of order": {
			clientFn: func(mockCtl *gomock.Controller) ecrpublicClient {
				mock := NewMockecrpublicClient(mockCtl)
				nextToken := generics.GetFakeObj[string]()

				mock.EXPECT().DescribeImageTags(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *ecrpublic.DescribeImageTagsInput, opts ...ecrpublic.Options) (*ecrpublic.DescribeImageTagsOutput, error) {
						assert.Equal(t, pluginTyp.RepositoryName(), *req.RepositoryName)

						secondImageTag := generics.GetFakeObj[ecrpublic_types.ImageTagDetail]()
						ts := imageTag.CreatedAt.Add(-time.Minute)
						secondImageTag.CreatedAt = &ts
						return &ecrpublic.DescribeImageTagsOutput{
							NextToken: generics.ToPtr(nextToken),
							ImageTagDetails: []ecrpublic_types.ImageTagDetail{
								imageTag,
							},
						}, nil
					})
				mock.EXPECT().DescribeImageTags(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *ecrpublic.DescribeImageTagsInput, opts ...ecrpublic.Options) (*ecrpublic.DescribeImageTagsOutput, error) {
						assert.Equal(t, pluginTyp.RepositoryName(), *req.RepositoryName)
						assert.Equal(t, nextToken, *req.NextToken)

						return &ecrpublic.DescribeImageTagsOutput{
							ImageTagDetails: []ecrpublic_types.ImageTagDetail{
								imageTag,
							},
						}, nil
					})
				return mock
			},
			assertFn: func(t *testing.T, plugins []*Plugin) {
				plugin := plugins[0]

				assert.Equal(t, *imageTag.ImageTag, plugin.Tag)
				assert.Equal(t, *imageTag.CreatedAt, plugin.CreatedAt)
				assert.Equal(t, pluginTyp.ImageURL(), plugin.ImageURL)
				assert.Equal(t, pluginTyp.RepositoryName(), plugin.RepositoryName)
			},
		},
		"error calling describe images": {
			clientFn: func(mockCtl *gomock.Controller) ecrpublicClient {
				mock := NewMockecrpublicClient(mockCtl)
				mock.EXPECT().DescribeImageTags(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errGetLatest)
				return mock
			},
			errExpected: errGetLatest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl, ctx := gomock.WithContext(ctx, t)

			client := test.clientFn(mockCtl)
			plugins, err := cat.getAll(ctx, client, pluginTyp)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, plugins)
		})
	}
}

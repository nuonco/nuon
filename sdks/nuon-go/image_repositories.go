package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type ListImageRepositoriesQuery struct {
	Pagination models.GetPaginatedQuery
	Q          string
	Platform   string
}

func (c *client) CreateImageRepository(ctx context.Context, req *models.ServiceCreateImageRepositoryRequest) (*models.AppImageRepository, error) {
	resp, err := c.genClient.Operations.CreateImageRepository(&operations.CreateImageRepositoryParams{Context: ctx, Req: req}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) ListImageRepositories(ctx context.Context, query *ListImageRepositoriesQuery) ([]*models.AppImageRepository, bool, error) {
	params := &operations.ListImageRepositoriesParams{Context: ctx}
	if query != nil {
		params.Offset, params.Limit = applyPaginationQuery(&query.Pagination)
		if query.Q != "" {
			params.Q = &query.Q
		}
		if query.Platform != "" {
			params.Platform = &query.Platform
		}
	}
	hr := newResponseHeaderReader(&operations.ListImageRepositoriesReader{})
	resp, err := c.genClient.Operations.ListImageRepositories(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}
	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetImageRepository(ctx context.Context, repositoryID string) (*models.AppImageRepository, error) {
	resp, err := c.genClient.Operations.GetImageRepository(&operations.GetImageRepositoryParams{Context: ctx, RepositoryID: repositoryID}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) DeleteImageRepository(ctx context.Context, repositoryID string) error {
	_, err := c.genClient.Operations.DeleteImageRepository(&operations.DeleteImageRepositoryParams{Context: ctx, RepositoryID: repositoryID}, c.getOrgIDAuthInfo())
	return err
}

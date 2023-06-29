package catalog

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	ecrpublic_types "github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
	"github.com/powertoolsdev/mono/pkg/generics"
)

// GetAll returns all versions for a plugin
func (c *catalog) GetAll(ctx context.Context, typ PluginType) ([]*Plugin, error) {
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecrpublic client: %w", err)
	}

	plugins, err := c.getAll(ctx, client, typ)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest: %w", err)
	}

	return plugins, nil
}

func (c *catalog) getAll(ctx context.Context, client ecrpublicClient, typ PluginType) ([]*Plugin, error) {
	images := make([]ecrpublic_types.ImageTagDetail, 0)

	repoName := typ.RepositoryName()
	if c.DevOverride {
		repoName = typ.DevRepositoryName()
	}

	input := &ecrpublic.DescribeImageTagsInput{
		RepositoryName: generics.ToPtr(repoName),
	}
	for {
		resp, err := client.DescribeImageTags(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("unable to describe image tags: %w", err)
		}

		images = append(images, resp.ImageTagDetails...)
		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}
	sort.Slice(images, func(i, j int) bool {
		return images[j].CreatedAt.Before(*images[i].CreatedAt)
	})

	plugins := make([]*Plugin, len(images))

	imgURL := typ.ImageURL()
	if c.DevOverride {
		imgURL = typ.DevImageURL()
	}
	for idx, img := range images {
		plugins[idx] = &Plugin{
			Tag:            *img.ImageTag,
			ImageURL:       imgURL,
			RepositoryName: repoName,
			CreatedAt:      *img.CreatedAt,
		}
	}

	return plugins, nil
}

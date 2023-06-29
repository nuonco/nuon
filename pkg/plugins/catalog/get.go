package catalog

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	ecrpublic_types "github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
	"github.com/powertoolsdev/mono/pkg/generics"
)

// GetLatest returns the latest plugin, based on a type
func (c *catalog) GetLatest(ctx context.Context, typ PluginType) (*Plugin, error) {
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecrpublic client: %w", err)
	}

	plugin, err := c.getLatest(ctx, client, typ)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest: %w", err)
	}

	return plugin, nil
}
func (c *catalog) getLatest(ctx context.Context, client ecrpublicClient, typ PluginType) (*Plugin, error) {
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

	if len(images) < 1 {
		return nil, fmt.Errorf("no images found %s", repoName)
	}

	img := images[0]
	imgURL := typ.ImageURL()
	if c.DevOverride {
		imgURL = typ.DevImageURL()
	}

	return &Plugin{
		Tag:            *img.ImageTag,
		ImageURL:       imgURL,
		RepositoryName: repoName,
		CreatedAt:      *img.CreatedAt,
	}, nil
}

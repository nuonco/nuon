package outputs

import (
	"context"
)

type imageOutputs struct {
	Tag string `json:"tag"`
}

// ImageOutputs are used anywhere an image is created, pushed or pulled.
//
// TODO(jm): use ORAS to grab image digests and other information off of the container image.
func ImageOutputs(ctx context.Context, tag string) (map[string]interface{}, error) {
	obj := imageOutputs{
		Tag: tag,
	}

	return ToMapstructure(obj)
}

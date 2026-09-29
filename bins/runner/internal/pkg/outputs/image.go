package outputs

import (
	"context"
)

type imageOutputs struct {
	Tag string `json:"tag"`
}

func ImageOutputs(ctx context.Context, tag string) (map[string]interface{}, error) {
	obj := imageOutputs{
		Tag: tag,
	}

	return ToMapstructure(obj)
}

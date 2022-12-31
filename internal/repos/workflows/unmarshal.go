package workflows

import (
	"fmt"

	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
	"google.golang.org/protobuf/proto"
)

func unmarshalResponse(byts []byte) (*sharedv1.Response, error) {
	resp := sharedv1.Response{}
	if err := proto.Unmarshal(byts, &resp); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	return &resp, nil
}

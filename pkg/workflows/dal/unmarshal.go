package dal

import (
	"fmt"

	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	"google.golang.org/protobuf/proto"
)

func unmarshalResponse(byts []byte) (*sharedv1.Response, error) {
	resp := sharedv1.Response{}
	if err := proto.Unmarshal(byts, &resp); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	return &resp, nil
}

func unmarshalRequest(byts []byte) (*sharedv1.Request, error) {
	req := sharedv1.Request{}
	if err := proto.Unmarshal(byts, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	return &req, nil
}

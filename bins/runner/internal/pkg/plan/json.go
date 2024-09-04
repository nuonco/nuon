package plan

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

func FromJSON(byt []byte, obj interface{}) error {
	if err := proto.Unmarshal(byt, &obj); err != nil {
		return fmt.Errorf("unable to parse json: %w", err)
	}

	return nil
}

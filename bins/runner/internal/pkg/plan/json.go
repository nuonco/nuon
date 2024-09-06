package plan

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func FromJSON(byt []byte, obj protoreflect.ProtoMessage) error {
	if err := proto.Unmarshal(byt, obj); err != nil {
		return fmt.Errorf("unable to parse json: %w", err)
	}

	return nil
}

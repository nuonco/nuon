package plan

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func FromJSON(byt []byte, obj protoreflect.ProtoMessage) error {
	if err := protojson.Unmarshal(byt, obj); err != nil {
		return fmt.Errorf("unable to parse json: %w", err)
	}

	return nil
}

package protos

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func ToJSON(msg protoreflect.ProtoMessage) ([]byte, error) {
	byts, err := protojson.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("unable to convert msg to json: %w", err)
	}

	return byts, nil
}

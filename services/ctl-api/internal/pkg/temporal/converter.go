package temporal

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

const (
	defaultTemporalJSONTag string = "temporal"
)

// temporalJSONConverter converts to/from JSON.
type temporalJSONConverter struct {
	json jsoniter.API
}

func (*temporalJSONConverter) newPayload(data []byte, c converter.PayloadConverter) *commonpb.Payload {
	return &commonpb.Payload{
		Metadata: map[string][]byte{
			"encoding": []byte(c.Encoding()),
		},
		Data: data,
	}
}

// ToPayload converts single value to payload.
func (c *temporalJSONConverter) ToPayload(value interface{}) (*commonpb.Payload, error) {
	data, err := c.json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", converter.ErrUnableToEncode, err)
	}
	return c.newPayload(data, c), nil
}

// FromPayload converts single value from payload.
func (c *temporalJSONConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
	err := c.json.Unmarshal(payload.GetData(), valuePtr)
	if err != nil {
		return fmt.Errorf("%w: %v", converter.ErrUnableToDecode, err)
	}
	return nil
}

// ToString converts payload object into human readable string.
func (c *temporalJSONConverter) ToString(payload *commonpb.Payload) string {
	return string(payload.GetData())
}

// Encoding returns MetadataEncodingJSON.
func (c *temporalJSONConverter) Encoding() string {
	return fmt.Sprintf("json/%s", defaultTemporalJSONTag)
}

var _ converter.PayloadConverter = (*temporalJSONConverter)(nil)

func newJSONConverter() converter.DataConverter {
	jc := &temporalJSONConverter{
		json: jsoniter.Config{
			EscapeHTML:             true,
			SortMapKeys:            true,
			ValidateJsonRawMessage: true,
			TagKey:                 defaultTemporalJSONTag,
		}.Froze(),
	}

	// note: the order of this is inmportant, to ensure we also support protobufs as well.
	return converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		converter.NewProtoJSONPayloadConverter(),
		converter.NewProtoPayloadConverter(),
		jc,
		converter.NewJSONPayloadConverter(),
	)
}

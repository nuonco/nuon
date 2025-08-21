package dataconverter

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

const (
	defaultTemporalJSONTag string = "temporaljson"
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

func (c *temporalJSONConverter) ToPayloads(values ...interface{}) (*commonpb.Payloads, error) {
	payloads := make([]*commonpb.Payload, len(values))
	for i, value := range values {
		payload, err := c.ToPayload(value)
		if err != nil {
			return nil, err
		}
		payloads[i] = payload
	}
	return &commonpb.Payloads{Payloads: payloads}, nil
}

// FromPayload converts single value from payload.
func (c *temporalJSONConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
	err := c.json.Unmarshal(payload.GetData(), valuePtr)
	if err != nil {
		return fmt.Errorf("%w: %v", converter.ErrUnableToDecode, err)
	}
	return nil
}

func (c *temporalJSONConverter) FromPayloads(payloads *commonpb.Payloads, valuePtrs ...interface{}) error {
	for i := range payloads.Payloads {
		err := c.FromPayload(payloads.Payloads[i], valuePtrs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// ToString converts payload object into human readable string.
func (c *temporalJSONConverter) ToString(payload *commonpb.Payload) string {
	return string(payload.GetData())
}

func (c *temporalJSONConverter) ToStrings(input *commonpb.Payloads) []string {
	strings := make([]string, len(input.Payloads))
	for i, payload := range input.Payloads {
		strings[i] = c.ToString(payload)
	}
	return strings
}

// Encoding returns MetadataEncodingJSON.
func (c *temporalJSONConverter) Encoding() string {
	return fmt.Sprintf("json/%s", defaultTemporalJSONTag)
}

var _ converter.PayloadConverter = (*temporalJSONConverter)(nil)

func NewJSONConverter() *temporalJSONConverter {
	return &temporalJSONConverter{
		json: jsoniter.Config{
			EscapeHTML:             true,
			SortMapKeys:            true,
			ValidateJsonRawMessage: true,
			TagKey:                 defaultTemporalJSONTag,
		}.Froze(),
	}
}

package signaldb

import (
	"encoding/base64"
	"encoding/json"
	"reflect"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal"
)

const (
	MetadataEncodingKey  = "encoding"
	MetadataEncodingType = "nuon/signal"
)

type PayloadConverter struct{}

func NewPayloadConverter() *PayloadConverter {
	return &PayloadConverter{}
}

var _ converter.PayloadConverter = (*PayloadConverter)(nil)

func newPayload(data []byte, c converter.PayloadConverter) *commonpb.Payload {
	return &commonpb.Payload{
		Metadata: map[string][]byte{
			MetadataEncodingKey: []byte(c.Encoding()),
		},
		Data: data,
	}
}

func (c *PayloadConverter) ToPayload(value interface{}) (*commonpb.Payload, error) {
	sig, ok := value.(signal.Signal)
	if !ok {
		return nil, nil
	}

	obj := signalJSON{
		Type: sig.Type(),
		Data: sig,
	}

	byts, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert signal into wire")
	}

	return newPayload(byts, c), nil
}

func (c *PayloadConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
	var out anyJSON
	if err := json.Unmarshal(payload.Data, &out); err != nil {
		return errors.Wrap(err, "unable to convert payload to object")
	}

	obj, err := catalog.NewFromType(out.Type)
	if err != nil {
		return err
	}

	if obj == nil {
		return errors.New("catalog type was nil")
	}

	if err := json.Unmarshal(out.Data, &obj); err != nil {
		return err
	}

	rv := reflect.ValueOf(valuePtr)
	if rv.Kind() != reflect.Ptr {
		return errors.New("valuePtr must be a pointer")
	}
	if rv.IsNil() {
		return errors.New("valuePtr cannot be nil")
	}

	// Dereference the pointer and get the underlying value
	elem := rv.Elem()

	// Check if the element is settable
	if !elem.CanSet() {
		return errors.New("cannot set value of valuePtr")
	}

	// Set the value
	elem.Set(reflect.ValueOf(obj))
	return nil
}

func (c *PayloadConverter) ToString(payload *commonpb.Payload) string {
	var byteSlice []byte
	err := c.FromPayload(payload, &byteSlice)
	if err != nil {
		return err.Error()
	}
	return base64.RawStdEncoding.EncodeToString(byteSlice)
}

func (c *PayloadConverter) Encoding() string {
	return MetadataEncodingType
}

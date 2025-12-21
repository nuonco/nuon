package signaldb

import (
	"encoding/base64"
	"encoding/json"
	"reflect"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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
	// Check if it's a direct signal
	if sig, ok := value.(signal.Signal); ok {
		return c.encodeSignal(sig)
	}

	// Check if it's a struct with a Signal field
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Struct {
		signalInterfaceType := reflect.TypeOf((*signal.Signal)(nil)).Elem()

		// Look for a Signal field
		for i := 0; i < rv.NumField(); i++ {
			field := rv.Field(i)
			if field.Type() == signalInterfaceType {
				// Found a Signal field - encode the struct with the signal
				sig, ok := field.Interface().(signal.Signal)
				if !ok || sig == nil {
					return nil, errors.New("Signal field is nil or invalid")
				}
				return c.encodeStructWithSignal(value, sig)
			}
		}
	}

	// Not something we handle
	return nil, nil
}

// encodeSignal encodes a bare signal
func (c *PayloadConverter) encodeSignal(sig signal.Signal) (*commonpb.Payload, error) {
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

// encodeStructWithSignal encodes a struct that contains a Signal field
func (c *PayloadConverter) encodeStructWithSignal(structValue interface{}, sig signal.Signal) (*commonpb.Payload, error) {
	// Encode the signal with type information
	signalPayload, err := c.encodeSignal(sig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode signal")
	}

	// Create a copy of the struct with the Signal field zeroed out for JSON serialization
	rv := reflect.ValueOf(structValue)
	if rv.Kind() != reflect.Struct {
		return nil, errors.New("structValue must be a struct")
	}

	// Create a map to hold non-Signal fields
	structFields := make(map[string]interface{})
	signalInterfaceType := reflect.TypeOf((*signal.Signal)(nil)).Elem()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rv.Type().Field(i)

		// Skip the Signal field
		if field.Type() == signalInterfaceType {
			continue
		}

		// Add other fields
		structFields[fieldType.Name] = field.Interface()
	}

	// Create composite payload
	composite := map[string]interface{}{
		"signal_data":   string(signalPayload.Data),
		"signal_meta":   signalPayload.Metadata,
		"struct_fields": structFields,
	}

	byts, err := json.Marshal(composite)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal composite payload")
	}

	// Use the same encoding - we'll detect composite structure in FromPayload
	return &commonpb.Payload{
		Metadata: map[string][]byte{
			MetadataEncodingKey: []byte(MetadataEncodingType),
		},
		Data: byts,
	}, nil
}

func (c *PayloadConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
	// Try to unmarshal and check if it's a composite payload
	var composite map[string]interface{}
	if err := json.Unmarshal(payload.Data, &composite); err == nil {
		// Check if it has the composite structure markers
		if _, hasSignalData := composite["signal_data"]; hasSignalData {
			if _, hasStructFields := composite["struct_fields"]; hasStructFields {
				// This is a composite payload
				return c.decodeStructWithSignal(payload, valuePtr)
			}
		}
	}

	// Standard signal decoding
	var out anyJSON
	if err := json.Unmarshal(payload.Data, &out); err != nil {
		return errors.Wrap(err, "unable to convert payload to object")
	}

	obj, err := catalog.NewFromType(out.Type)
	if err != nil {
		return errors.Wrap(err, "unable to get type from catalog")
	}

	if obj == nil {
		return errors.New("catalog type was nil")
	}

	if err := json.Unmarshal(out.Data, obj); err != nil {
		return errors.Wrap(err, "unable to unmarshal signal into underlying type")
	}

	// Verify obj is not nil after unmarshaling (check both interface and underlying value)
	if obj == nil {
		return errors.New("unmarshaled object is nil (interface is nil)")
	}

	objValue := reflect.ValueOf(obj)
	if !objValue.IsValid() {
		return errors.New("unmarshaled object has invalid reflect value")
	}

	// For pointer types, check if the pointer itself is nil
	if objValue.Kind() == reflect.Ptr && objValue.IsNil() {
		return errors.New("unmarshaled object is nil (underlying value is nil)")
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

	// Check if valuePtr is a direct signal.Signal interface
	signalInterfaceType := reflect.TypeOf((*signal.Signal)(nil)).Elem()
	if elem.Type() == signalInterfaceType {
		// Direct signal deserialization
		elem.Set(reflect.ValueOf(obj))
		return nil
	}

	// Check if valuePtr is a struct with a Signal field (like CreateQueueSignalRequest)
	if elem.Kind() == reflect.Struct {
		for i := 0; i < elem.NumField(); i++ {
			field := elem.Field(i)
			if field.Type() == signalInterfaceType {
				if field.CanSet() {
					field.Set(reflect.ValueOf(obj))
					return nil
				}
				return errors.New("Signal field found but cannot be set")
			}
		}
		return errors.New("no Signal field found in struct")
	}

	// Fallback: unknown type
	return errors.Errorf("unsupported valuePtr type: %T", valuePtr)
}

// decodeStructWithSignal decodes a composite payload back into a struct with Signal field
func (c *PayloadConverter) decodeStructWithSignal(payload *commonpb.Payload, valuePtr interface{}) error {
	// Unmarshal the composite structure
	var composite map[string]interface{}
	if err := json.Unmarshal(payload.Data, &composite); err != nil {
		return errors.Wrap(err, "unable to unmarshal composite payload")
	}

	// Extract signal data
	signalDataStr, ok := composite["signal_data"].(string)
	if !ok {
		return errors.New("missing or invalid signal_data in composite payload")
	}

	// Reconstruct the signal
	var signalOut anyJSON
	if err := json.Unmarshal([]byte(signalDataStr), &signalOut); err != nil {
		return errors.Wrap(err, "unable to unmarshal signal data")
	}

	obj, err := catalog.NewFromType(signalOut.Type)
	if err != nil {
		return errors.Wrap(err, "unable to get type from catalog")
	}

	if obj == nil {
		return errors.New("catalog type was nil")
	}

	if err := json.Unmarshal(signalOut.Data, obj); err != nil {
		return errors.Wrap(err, "unable to unmarshal signal into underlying type")
	}

	// Verify obj is not nil after unmarshaling
	if obj == nil {
		return errors.New("unmarshaled object is nil (interface is nil)")
	}

	objValue := reflect.ValueOf(obj)
	if !objValue.IsValid() {
		return errors.New("unmarshaled object has invalid reflect value")
	}

	// For pointer types, check if the pointer itself is nil
	if objValue.Kind() == reflect.Ptr && objValue.IsNil() {
		return errors.New("unmarshaled object is nil (underlying value is nil)")
	}

	// Now handle the struct
	rv := reflect.ValueOf(valuePtr)
	if rv.Kind() != reflect.Ptr {
		return errors.New("valuePtr must be a pointer")
	}
	if rv.IsNil() {
		return errors.New("valuePtr cannot be nil")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("valuePtr must be a pointer to a struct for composite payloads")
	}

	// Populate non-Signal fields from struct_fields
	structFieldsData, ok := composite["struct_fields"].(map[string]interface{})
	if ok {
		// Set each field individually
		for i := 0; i < elem.NumField(); i++ {
			field := elem.Field(i)
			fieldType := elem.Type().Field(i)

			// Skip Signal fields
			signalInterfaceType := reflect.TypeOf((*signal.Signal)(nil)).Elem()
			if field.Type() == signalInterfaceType {
				continue
			}

			// Set other fields from the map
			if fieldValue, exists := structFieldsData[fieldType.Name]; exists && field.CanSet() {
				// Convert the interface value to the correct type
				fieldValueJSON, err := json.Marshal(fieldValue)
				if err != nil {
					return errors.Wrapf(err, "unable to marshal field %s", fieldType.Name)
				}

				fieldPtr := reflect.New(field.Type())
				if err := json.Unmarshal(fieldValueJSON, fieldPtr.Interface()); err != nil {
					return errors.Wrapf(err, "unable to unmarshal field %s", fieldType.Name)
				}

				field.Set(fieldPtr.Elem())
			}
		}
	}

	// Now set the Signal field
	signalInterfaceType := reflect.TypeOf((*signal.Signal)(nil)).Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if field.Type() == signalInterfaceType {
			if field.CanSet() {
				field.Set(reflect.ValueOf(obj))
				return nil
			}
			return errors.New("Signal field found but cannot be set")
		}
	}

	return errors.New("no Signal field found in target struct")
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

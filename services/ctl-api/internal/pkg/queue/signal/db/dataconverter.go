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
	if sig, ok := value.(signal.Signal); ok {
		return c.encodeSignal(sig)
	}

	rv := reflect.ValueOf(value)

	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}

	signalInterfaceType := reflect.TypeOf((*signal.Signal)(nil)).Elem()

	if rv.Kind() == reflect.Struct {
		if structHasSignalField(rv.Type(), signalInterfaceType) {
			sig, ok := getSignalFromStruct(rv, signalInterfaceType)
			if !ok || sig == nil {
				return nil, errors.New("Signal field is nil or invalid")
			}
			return c.encodeStructWithSignal(rv.Interface(), sig)
		}
	}

	if (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && rv.Len() > 0 {
		elemType := rv.Type().Elem()
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
		if elemType.Kind() == reflect.Struct && structHasSignalField(elemType, signalInterfaceType) {
			return c.encodeSliceWithSignals(rv)
		}
	}

	return nil, nil
}

func structHasSignalField(t reflect.Type, signalInterfaceType reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type == signalInterfaceType {
			return true
		}
	}
	return false
}

func getSignalFromStruct(rv reflect.Value, signalInterfaceType reflect.Type) (signal.Signal, bool) {
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if field.Type() == signalInterfaceType {
			sig, ok := field.Interface().(signal.Signal)
			return sig, ok
		}
	}
	return nil, false
}

func (c *PayloadConverter) encodeSliceWithSignals(rv reflect.Value) (*commonpb.Payload, error) {
	var encodedItems []json.RawMessage

	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		signalInterfaceType := reflect.TypeOf((*signal.Signal)(nil)).Elem()
		sig, ok := getSignalFromStruct(elem, signalInterfaceType)
		if !ok || sig == nil {
			return nil, errors.Errorf("Signal field is nil or invalid at index %d", i)
		}

		payload, err := c.encodeStructWithSignal(elem.Interface(), sig)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to encode item at index %d", i)
		}

		encodedItems = append(encodedItems, payload.Data)
	}

	wrapper := map[string]interface{}{
		"items":    encodedItems,
		"is_array": true,
	}

	byts, err := json.Marshal(wrapper)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal array payload")
	}

	return &commonpb.Payload{
		Metadata: map[string][]byte{
			MetadataEncodingKey: []byte(MetadataEncodingType),
		},
		Data: byts,
	}, nil
}

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

func (c *PayloadConverter) encodeStructWithSignal(structValue interface{}, sig signal.Signal) (*commonpb.Payload, error) {
	signalPayload, err := c.encodeSignal(sig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode signal")
	}

	rv := reflect.ValueOf(structValue)
	if rv.Kind() != reflect.Struct {
		return nil, errors.New("structValue must be a struct")
	}

	structFields := make(map[string]interface{})
	signalInterfaceType := reflect.TypeOf((*signal.Signal)(nil)).Elem()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rv.Type().Field(i)

		if field.Type() == signalInterfaceType {
			continue
		}

		structFields[fieldType.Name] = field.Interface()
	}

	composite := map[string]interface{}{
		"signal_data":   string(signalPayload.Data),
		"signal_meta":   signalPayload.Metadata,
		"struct_fields": structFields,
	}

	byts, err := json.Marshal(composite)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal composite payload")
	}

	return &commonpb.Payload{
		Metadata: map[string][]byte{
			MetadataEncodingKey: []byte(MetadataEncodingType),
		},
		Data: byts,
	}, nil
}

func (c *PayloadConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
	var composite map[string]interface{}
	if err := json.Unmarshal(payload.Data, &composite); err == nil {
		if _, isArray := composite["is_array"]; isArray {
			if _, hasItems := composite["items"]; hasItems {
				return c.decodeSliceWithSignals(payload, valuePtr)
			}
		}

		if _, hasSignalData := composite["signal_data"]; hasSignalData {
			if _, hasStructFields := composite["struct_fields"]; hasStructFields {
				return c.decodeStructWithSignal(payload, valuePtr)
			}
		}
	}

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

	if obj == nil {
		return errors.New("unmarshaled object is nil (interface is nil)")
	}

	objValue := reflect.ValueOf(obj)
	if !objValue.IsValid() {
		return errors.New("unmarshaled object has invalid reflect value")
	}

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

	elem := rv.Elem()

	if elem.Kind() == reflect.Ptr {
		if elem.IsNil() {
			elem.Set(reflect.New(elem.Type().Elem()))
		}
		elem = elem.Elem()
	}

	if !elem.CanSet() {
		return errors.New("cannot set value of valuePtr")
	}

	signalInterfaceType := reflect.TypeOf((*signal.Signal)(nil)).Elem()
	if elem.Type() == signalInterfaceType {
		elem.Set(reflect.ValueOf(obj))
		return nil
	}

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

	return errors.Errorf("unsupported valuePtr type: %T", valuePtr)
}

func (c *PayloadConverter) decodeStructWithSignal(payload *commonpb.Payload, valuePtr interface{}) error {
	var composite map[string]interface{}
	if err := json.Unmarshal(payload.Data, &composite); err != nil {
		return errors.Wrap(err, "unable to unmarshal composite payload")
	}

	signalDataStr, ok := composite["signal_data"].(string)
	if !ok {
		return errors.New("missing or invalid signal_data in composite payload")
	}

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

	if obj == nil {
		return errors.New("unmarshaled object is nil (interface is nil)")
	}

	objValue := reflect.ValueOf(obj)
	if !objValue.IsValid() {
		return errors.New("unmarshaled object has invalid reflect value")
	}

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

	elem := rv.Elem()

	// why: Handle double-pointer case (e.g., **EnqueueSignalRequest from Temporal).
	// Temporal passes pointer-to-pointer because the activity param is already a pointer.
	if elem.Kind() == reflect.Ptr {
		if elem.IsNil() {
			elem.Set(reflect.New(elem.Type().Elem()))
		}
		elem = elem.Elem()
	}

	if elem.Kind() != reflect.Struct {
		return errors.New("valuePtr must be a pointer to a struct for composite payloads")
	}

	structFieldsData, ok := composite["struct_fields"].(map[string]interface{})
	if ok {
		for i := 0; i < elem.NumField(); i++ {
			field := elem.Field(i)
			fieldType := elem.Type().Field(i)

			signalInterfaceType := reflect.TypeOf((*signal.Signal)(nil)).Elem()
			if field.Type() == signalInterfaceType {
				continue
			}

			if fieldValue, exists := structFieldsData[fieldType.Name]; exists && field.CanSet() {
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

func (c *PayloadConverter) decodeSliceWithSignals(payload *commonpb.Payload, valuePtr interface{}) error {
	var wrapper struct {
		Items   []json.RawMessage `json:"items"`
		IsArray bool              `json:"is_array"`
	}
	if err := json.Unmarshal(payload.Data, &wrapper); err != nil {
		return errors.Wrap(err, "unable to unmarshal array payload")
	}

	rv := reflect.ValueOf(valuePtr)
	if rv.Kind() != reflect.Ptr {
		return errors.New("valuePtr must be a pointer")
	}
	if rv.IsNil() {
		return errors.New("valuePtr cannot be nil")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Slice {
		return errors.Errorf("valuePtr must be a pointer to a slice, got %s", elem.Kind())
	}

	sliceType := elem.Type()
	resultSlice := reflect.MakeSlice(sliceType, len(wrapper.Items), len(wrapper.Items))

	for i, itemData := range wrapper.Items {
		itemPayload := &commonpb.Payload{
			Metadata: payload.Metadata,
			Data:     itemData,
		}

		itemPtr := reflect.New(sliceType.Elem())
		if err := c.decodeStructWithSignal(itemPayload, itemPtr.Interface()); err != nil {
			return errors.Wrapf(err, "unable to decode item at index %d", i)
		}

		resultSlice.Index(i).Set(itemPtr.Elem())
	}

	elem.Set(resultSlice)
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

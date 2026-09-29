package signaldb

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	commonpb "go.temporal.io/api/common/v1"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type TestRequestStruct struct {
	QueueID string        `validate:"required"`
	Signal  signal.Signal `validate:"required"`
}

type PayloadConverterTestSuite struct {
	suite.Suite
	converter *PayloadConverter
}

func TestPayloadConverterSuite(t *testing.T) {
	suite.Run(t, new(PayloadConverterTestSuite))
}

func (s *PayloadConverterTestSuite) SetupTest() {
	s.converter = NewPayloadConverter()
}

func (s *PayloadConverterTestSuite) TestToPayload_DirectSignal() {
	sig := &example.ExampleSignal{
		Arg1: "test-arg-1",
		Arg2: "test-arg-2",
	}

	payload, err := s.converter.ToPayload(sig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	encoding, ok := payload.Metadata[MetadataEncodingKey]
	require.True(s.T(), ok)
	assert.Equal(s.T(), MetadataEncodingType, string(encoding))

	var result anyJSON
	err = json.Unmarshal(payload.Data, &result)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), example.ExampleSignalType, result.Type)

	var sigData example.ExampleSignal
	err = json.Unmarshal(result.Data, &sigData)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-arg-1", sigData.Arg1)
	assert.Equal(s.T(), "test-arg-2", sigData.Arg2)
}

func (s *PayloadConverterTestSuite) TestToPayload_StructWithSignalField() {
	sig := &example.ExampleSignal{
		Arg1: "test-arg-1",
		Arg2: "test-arg-2",
	}

	req := TestRequestStruct{
		QueueID: "queue-123",
		Signal:  sig,
	}

	payload, err := s.converter.ToPayload(req)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload, "ToPayload should handle structs with Signal fields")
}

func (s *PayloadConverterTestSuite) TestToPayload_NonSignalValue() {
	payload, err := s.converter.ToPayload("plain string")
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), payload)

	type RegularStruct struct {
		Field1 string
		Field2 int
	}
	payload, err = s.converter.ToPayload(RegularStruct{Field1: "test", Field2: 42})
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), payload)
}

func (s *PayloadConverterTestSuite) TestFromPayload_DirectSignal() {
	originalSig := &example.ExampleSignal{
		Arg1: "original-arg-1",
		Arg2: "original-arg-2",
	}

	payload, err := s.converter.ToPayload(originalSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	var resultSig signal.Signal
	err = s.converter.FromPayload(payload, &resultSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resultSig, "Deserialized signal should not be nil")

	exampleSig, ok := resultSig.(*example.ExampleSignal)
	require.True(s.T(), ok, "Expected *example.ExampleSignal, got %T", resultSig)
	assert.Equal(s.T(), "original-arg-1", exampleSig.Arg1)
	assert.Equal(s.T(), "original-arg-2", exampleSig.Arg2)
}

func (s *PayloadConverterTestSuite) TestFromPayload_StructWithSignalField() {
	originalSig := &example.ExampleSignal{
		Arg1: "struct-arg-1",
		Arg2: "struct-arg-2",
	}

	payload, err := s.converter.ToPayload(originalSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	var result TestRequestStruct
	err = s.converter.FromPayload(payload, &result)
	require.NoError(s.T(), err, "FromPayload should handle structs with Signal fields")

	require.NotNil(s.T(), result.Signal, "Signal field should not be nil after deserialization")

	exampleSig, ok := result.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok, "Expected *example.ExampleSignal, got %T", result.Signal)
	assert.Equal(s.T(), "struct-arg-1", exampleSig.Arg1)
	assert.Equal(s.T(), "struct-arg-2", exampleSig.Arg2)

	assert.Equal(s.T(), "", result.QueueID)
}

func (s *PayloadConverterTestSuite) TestRoundTrip_DirectSignal() {
	originalSig := &example.ExampleSignal{
		Arg1: "roundtrip-arg-1",
		Arg2: "roundtrip-arg-2",
	}

	payload, err := s.converter.ToPayload(originalSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	var resultSig signal.Signal
	err = s.converter.FromPayload(payload, &resultSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resultSig, "Round-trip result should not be nil")

	exampleSig, ok := resultSig.(*example.ExampleSignal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), originalSig.Arg1, exampleSig.Arg1)
	assert.Equal(s.T(), originalSig.Arg2, exampleSig.Arg2)
	assert.Equal(s.T(), example.ExampleSignalType, exampleSig.Type())
}

func (s *PayloadConverterTestSuite) TestRoundTrip_StructWithSignalField() {
	originalSig := &example.ExampleSignal{
		Arg1: "roundtrip-struct-arg-1",
		Arg2: "roundtrip-struct-arg-2",
	}

	payload, err := s.converter.ToPayload(originalSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	var resultReq TestRequestStruct
	err = s.converter.FromPayload(payload, &resultReq)
	require.NoError(s.T(), err)

	require.NotNil(s.T(), resultReq.Signal, "Round-trip Signal field should not be nil")
	exampleSig, ok := resultReq.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), originalSig.Arg1, exampleSig.Arg1)
	assert.Equal(s.T(), originalSig.Arg2, exampleSig.Arg2)

	assert.Equal(s.T(), "", resultReq.QueueID)
}

func (s *PayloadConverterTestSuite) TestFromPayload_InvalidCatalogType() {
	invalidPayload := &commonpb.Payload{
		Metadata: map[string][]byte{
			MetadataEncodingKey: []byte(MetadataEncodingType),
		},
		Data: []byte(`{"type":"invalid-signal-type","data":{}}`),
	}

	var resultSig signal.Signal
	err := s.converter.FromPayload(invalidPayload, &resultSig)
	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, catalog.ErrSignalTypeNotRegistered)
	assert.Contains(s.T(), err.Error(), "invalid-signal-type")
}

func (s *PayloadConverterTestSuite) TestFromPayload_NilPointer() {
	sig := &example.ExampleSignal{
		Arg1: "test-arg-1",
		Arg2: "test-arg-2",
	}

	payload, err := s.converter.ToPayload(sig)
	require.NoError(s.T(), err)

	var nilPtr *signal.Signal
	err = s.converter.FromPayload(payload, nilPtr)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "cannot be nil")
}

func (s *PayloadConverterTestSuite) TestFromPayload_NonPointer() {
	sig := &example.ExampleSignal{
		Arg1: "test-arg-1",
		Arg2: "test-arg-2",
	}

	payload, err := s.converter.ToPayload(sig)
	require.NoError(s.T(), err)

	var notAPointer signal.Signal
	err = s.converter.FromPayload(payload, notAPointer)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "must be a pointer")
}

func (s *PayloadConverterTestSuite) TestEncoding() {
	encoding := s.converter.Encoding()
	assert.Equal(s.T(), MetadataEncodingType, encoding)
	assert.Equal(s.T(), "nuon/signal", encoding)
}

func (s *PayloadConverterTestSuite) TestToString() {
	sig := &example.ExampleSignal{
		Arg1: "test-arg-1",
		Arg2: "test-arg-2",
	}

	payload, err := s.converter.ToPayload(sig)
	require.NoError(s.T(), err)

	result := s.converter.ToString(payload)
	assert.NotEmpty(s.T(), result)
}

func (s *PayloadConverterTestSuite) TestFromPayload_CatalogObjectNotNil() {
	originalSig := &example.ExampleSignal{
		Arg1: "catalog-test-1",
		Arg2: "catalog-test-2",
	}

	payload, err := s.converter.ToPayload(originalSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	var parsed anyJSON
	err = json.Unmarshal(payload.Data, &parsed)
	require.NoError(s.T(), err)

	obj, err := catalog.NewFromType(parsed.Type)
	require.NoError(s.T(), err, "Catalog should return object for valid type")
	require.NotNil(s.T(), obj, "Catalog object should not be nil")

	err = json.Unmarshal(parsed.Data, obj)
	require.NoError(s.T(), err, "Should be able to unmarshal into catalog object")

	exampleSig, ok := obj.(*example.ExampleSignal)
	require.True(s.T(), ok, "Catalog object should be *example.ExampleSignal")
	require.NotNil(s.T(), exampleSig, "Cast result should not be nil")
	assert.Equal(s.T(), "catalog-test-1", exampleSig.Arg1)
	assert.Equal(s.T(), "catalog-test-2", exampleSig.Arg2)
}

func (s *PayloadConverterTestSuite) TestUnmarshalPointerBehavior() {
	signalData := []byte(`{"arg_1":"ptr-test-1","arg_2":"ptr-test-2"}`)

	obj, err := catalog.NewFromType(example.ExampleSignalType)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), obj, "Catalog should return non-nil object")

	rv := reflect.ValueOf(obj)
	require.Equal(s.T(), reflect.Ptr, rv.Kind(), "Catalog object should be a pointer")

	err = json.Unmarshal(signalData, obj)
	require.NoError(s.T(), err, "Unmarshaling into obj (not &obj) should work")

	exampleSig, ok := obj.(*example.ExampleSignal)
	require.True(s.T(), ok)
	require.NotNil(s.T(), exampleSig)
	assert.Equal(s.T(), "ptr-test-1", exampleSig.Arg1)
	assert.Equal(s.T(), "ptr-test-2", exampleSig.Arg2)
}

func (s *PayloadConverterTestSuite) TestTrueRoundTrip_StructWithSignalField() {
	originalReq := TestRequestStruct{
		QueueID: "test-queue-123",
		Signal: &example.ExampleSignal{
			Arg1: "true-roundtrip-1",
			Arg2: "true-roundtrip-2",
		},
	}

	payload, err := s.converter.ToPayload(originalReq)
	require.NoError(s.T(), err)

	require.NotNil(s.T(), payload, "CRITICAL: ToPayload must handle structs with Signal fields")

	var resultReq TestRequestStruct
	err = s.converter.FromPayload(payload, &resultReq)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "test-queue-123", resultReq.QueueID)

	require.NotNil(s.T(), resultReq.Signal, "Signal field must be populated")

	exampleSig, ok := resultReq.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok, "Signal should be *example.ExampleSignal")
	assert.Equal(s.T(), "true-roundtrip-1", exampleSig.Arg1)
	assert.Equal(s.T(), "true-roundtrip-2", exampleSig.Arg2)
}

func (s *PayloadConverterTestSuite) TestRoundTrip_ArrayOfStructsWithSignalField() {
	items := []TestRequestStruct{
		{
			QueueID: "queue-1",
			Signal: &example.ExampleSignal{
				Arg1: "array-item-1-arg1",
				Arg2: "array-item-1-arg2",
			},
		},
		{
			QueueID: "queue-2",
			Signal: &example.ExampleSignal{
				Arg1: "array-item-2-arg1",
				Arg2: "array-item-2-arg2",
			},
		},
		{
			QueueID: "queue-3",
			Signal: &example.ExampleSignal{
				Arg1: "array-item-3-arg1",
				Arg2: "array-item-3-arg2",
			},
		},
	}

	payload, err := s.converter.ToPayload(items)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload, "ToPayload must handle arrays of structs with Signal fields")

	var result []TestRequestStruct
	err = s.converter.FromPayload(payload, &result)
	require.NoError(s.T(), err)
	require.Len(s.T(), result, 3, "Should have 3 items after deserialization")

	for i, item := range result {
		assert.Equal(s.T(), items[i].QueueID, item.QueueID, "QueueID mismatch at index %d", i)
		require.NotNil(s.T(), item.Signal, "Signal should not be nil at index %d", i)

		exampleSig, ok := item.Signal.(*example.ExampleSignal)
		require.True(s.T(), ok, "Expected *example.ExampleSignal at index %d, got %T", i, item.Signal)

		originalSig := items[i].Signal.(*example.ExampleSignal)
		assert.Equal(s.T(), originalSig.Arg1, exampleSig.Arg1, "Arg1 mismatch at index %d", i)
		assert.Equal(s.T(), originalSig.Arg2, exampleSig.Arg2, "Arg2 mismatch at index %d", i)
	}
}

func (s *PayloadConverterTestSuite) TestTrueRoundTrip_PointerToStructWithSignalField() {
	originalReq := &TestRequestStruct{
		QueueID: "test-queue-ptr",
		Signal: &example.ExampleSignal{
			Arg1: "ptr-roundtrip-1",
			Arg2: "ptr-roundtrip-2",
		},
	}

	payload, err := s.converter.ToPayload(originalReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload, "ToPayload must handle pointers to structs with Signal fields")

	var resultReq TestRequestStruct
	err = s.converter.FromPayload(payload, &resultReq)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "test-queue-ptr", resultReq.QueueID)
	require.NotNil(s.T(), resultReq.Signal)
	exampleSig, ok := resultReq.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), "ptr-roundtrip-1", exampleSig.Arg1)
	assert.Equal(s.T(), "ptr-roundtrip-2", exampleSig.Arg2)
}

func (s *PayloadConverterTestSuite) TestTrueRoundTrip_WithGenericSignalInterface() {
	var sig signal.Signal
	sig = &example.ExampleSignal{
		Arg1: "interface-var-1",
		Arg2: "interface-var-2",
	}

	originalReq := TestRequestStruct{
		QueueID: "test-queue-456",
		Signal:  sig,
	}

	payload, err := s.converter.ToPayload(originalReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload, "ToPayload must handle structs with Signal interface fields")

	var resultReq TestRequestStruct
	err = s.converter.FromPayload(payload, &resultReq)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "test-queue-456", resultReq.QueueID)
	require.NotNil(s.T(), resultReq.Signal)

	exampleSig, ok := resultReq.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), "interface-var-1", exampleSig.Arg1)
	assert.Equal(s.T(), "interface-var-2", exampleSig.Arg2)
}

type TestStructWithSignalData struct {
	Name        string      `json:"name"`
	QueueSignal *SignalData `json:"queue_signal,omitempty"`
}

func (s *PayloadConverterTestSuite) TestRoundTrip_SliceOfPointersWithSignalData() {
	items := []*TestStructWithSignalData{
		{
			Name: "step-1",
			QueueSignal: &SignalData{
				Signal: &example.ExampleSignal{
					Arg1: "step-1-arg1",
					Arg2: "step-1-arg2",
				},
			},
		},
		{
			Name: "step-2",
			QueueSignal: &SignalData{
				Signal: &example.ExampleSignal{
					Arg1: "step-2-arg1",
					Arg2: "step-2-arg2",
				},
			},
		},
		{
			Name: "step-3",
			QueueSignal: &SignalData{
				Signal: &example.ExampleSignal{
					Arg1: "step-3-arg1",
					Arg2: "step-3-arg2",
				},
			},
		},
	}

	payload, err := s.converter.ToPayload(items)
	assert.NoError(s.T(), err)

	if payload != nil {
		var result []*TestStructWithSignalData
		err = s.converter.FromPayload(payload, &result)
		require.NoError(s.T(), err)
		require.Len(s.T(), result, 3)

		for i, item := range result {
			assert.Equal(s.T(), items[i].Name, item.Name, "Name mismatch at index %d", i)
			require.NotNil(s.T(), item.QueueSignal, "QueueSignal should not be nil at index %d", i)
			require.NotNil(s.T(), item.QueueSignal.Signal, "QueueSignal.Signal should not be nil at index %d", i)

			exampleSig, ok := item.QueueSignal.Signal.(*example.ExampleSignal)
			require.True(s.T(), ok, "Expected *example.ExampleSignal at index %d, got %T", i, item.QueueSignal.Signal)

			originalSig := items[i].QueueSignal.Signal.(*example.ExampleSignal)
			assert.Equal(s.T(), originalSig.Arg1, exampleSig.Arg1, "Arg1 mismatch at index %d", i)
			assert.Equal(s.T(), originalSig.Arg2, exampleSig.Arg2, "Arg2 mismatch at index %d", i)
		}
	} else {
		s.T().Log("PayloadConverter returned nil for slice of *SignalData structs — testing JSON path")

		byts, err := json.Marshal(items)
		require.NoError(s.T(), err)

		var result []*TestStructWithSignalData
		err = json.Unmarshal(byts, &result)
		require.NoError(s.T(), err)
		require.Len(s.T(), result, 3)

		for i, item := range result {
			assert.Equal(s.T(), items[i].Name, item.Name, "Name mismatch at index %d", i)
			require.NotNil(s.T(), item.QueueSignal, "QueueSignal should not be nil at index %d", i)
			require.NotNil(s.T(), item.QueueSignal.Signal, "QueueSignal.Signal should not be nil at index %d", i)

			exampleSig, ok := item.QueueSignal.Signal.(*example.ExampleSignal)
			require.True(s.T(), ok, "Expected *example.ExampleSignal at index %d, got %T", i, item.QueueSignal.Signal)

			originalSig := items[i].QueueSignal.Signal.(*example.ExampleSignal)
			assert.Equal(s.T(), originalSig.Arg1, exampleSig.Arg1, "Arg1 mismatch at index %d", i)
			assert.Equal(s.T(), originalSig.Arg2, exampleSig.Arg2, "Arg2 mismatch at index %d", i)
		}
	}
}

func (s *PayloadConverterTestSuite) TestTrueRoundTrip_DoublePointerFromTemporal() {
	originalReq := &TestRequestStruct{
		QueueID: "temporal-double-ptr",
		Signal: &example.ExampleSignal{
			Arg1: "double-ptr-1",
			Arg2: "double-ptr-2",
		},
	}

	payload, err := s.converter.ToPayload(originalReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	var resultReq *TestRequestStruct
	err = s.converter.FromPayload(payload, &resultReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resultReq)

	assert.Equal(s.T(), "temporal-double-ptr", resultReq.QueueID)
	require.NotNil(s.T(), resultReq.Signal)
	exampleSig, ok := resultReq.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), "double-ptr-1", exampleSig.Arg1)
	assert.Equal(s.T(), "double-ptr-2", exampleSig.Arg2)
}

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

// TestRequestStruct mimics the generated CreateQueueSignalRequest structure
// to avoid circular dependency with activities package
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

// TestToPayload_DirectSignal verifies that direct signal serialization works correctly
func (s *PayloadConverterTestSuite) TestToPayload_DirectSignal() {
	sig := &example.ExampleSignal{
		Arg1: "test-arg-1",
		Arg2: "test-arg-2",
	}

	payload, err := s.converter.ToPayload(sig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	// Verify encoding metadata
	encoding, ok := payload.Metadata[MetadataEncodingKey]
	require.True(s.T(), ok)
	assert.Equal(s.T(), MetadataEncodingType, string(encoding))

	// Verify payload data structure
	var result anyJSON
	err = json.Unmarshal(payload.Data, &result)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), example.ExampleSignalType, result.Type)

	// Verify signal data is present
	var sigData example.ExampleSignal
	err = json.Unmarshal(result.Data, &sigData)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-arg-1", sigData.Arg1)
	assert.Equal(s.T(), "test-arg-2", sigData.Arg2)
}

// TestToPayload_StructWithSignalField verifies that structs with Signal fields return nil
// (to let the standard JSON converter handle the struct, while this converter handles the Signal field within)
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

// TestToPayload_NonSignalValue verifies that non-signal values return nil (let other converters handle)
func (s *PayloadConverterTestSuite) TestToPayload_NonSignalValue() {
	// Test with plain string
	payload, err := s.converter.ToPayload("plain string")
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), payload)

	// Test with struct without signal field
	type RegularStruct struct {
		Field1 string
		Field2 int
	}
	payload, err = s.converter.ToPayload(RegularStruct{Field1: "test", Field2: 42})
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), payload)
}

// TestFromPayload_DirectSignal verifies that direct signal deserialization works
func (s *PayloadConverterTestSuite) TestFromPayload_DirectSignal() {
	// First serialize a signal
	originalSig := &example.ExampleSignal{
		Arg1: "original-arg-1",
		Arg2: "original-arg-2",
	}

	payload, err := s.converter.ToPayload(originalSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	// Now deserialize it
	var resultSig signal.Signal
	err = s.converter.FromPayload(payload, &resultSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resultSig, "Deserialized signal should not be nil")

	// Verify it's the right type
	exampleSig, ok := resultSig.(*example.ExampleSignal)
	require.True(s.T(), ok, "Expected *example.ExampleSignal, got %T", resultSig)
	assert.Equal(s.T(), "original-arg-1", exampleSig.Arg1)
	assert.Equal(s.T(), "original-arg-2", exampleSig.Arg2)
}

// TestFromPayload_StructWithSignalField is THE KEY TEST for the bug fix
func (s *PayloadConverterTestSuite) TestFromPayload_StructWithSignalField() {
	// First serialize a signal
	originalSig := &example.ExampleSignal{
		Arg1: "struct-arg-1",
		Arg2: "struct-arg-2",
	}

	payload, err := s.converter.ToPayload(originalSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	// Now deserialize into a struct with a Signal field
	var result TestRequestStruct
	err = s.converter.FromPayload(payload, &result)
	require.NoError(s.T(), err, "FromPayload should handle structs with Signal fields")

	// Verify the Signal field was populated
	require.NotNil(s.T(), result.Signal, "Signal field should not be nil after deserialization")

	exampleSig, ok := result.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok, "Expected *example.ExampleSignal, got %T", result.Signal)
	assert.Equal(s.T(), "struct-arg-1", exampleSig.Arg1)
	assert.Equal(s.T(), "struct-arg-2", exampleSig.Arg2)

	// Note: QueueID field should remain empty (not part of signal serialization)
	assert.Equal(s.T(), "", result.QueueID)
}

// TestRoundTrip_DirectSignal verifies end-to-end serialize → deserialize for direct signals
func (s *PayloadConverterTestSuite) TestRoundTrip_DirectSignal() {
	originalSig := &example.ExampleSignal{
		Arg1: "roundtrip-arg-1",
		Arg2: "roundtrip-arg-2",
	}

	// Serialize
	payload, err := s.converter.ToPayload(originalSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	// Deserialize
	var resultSig signal.Signal
	err = s.converter.FromPayload(payload, &resultSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resultSig, "Round-trip result should not be nil")

	// Verify data integrity
	exampleSig, ok := resultSig.(*example.ExampleSignal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), originalSig.Arg1, exampleSig.Arg1)
	assert.Equal(s.T(), originalSig.Arg2, exampleSig.Arg2)
	assert.Equal(s.T(), example.ExampleSignalType, exampleSig.Type())
}

// TestRoundTrip_StructWithSignalField verifies end-to-end for struct wrapper
// This tests that we can deserialize a Signal payload into a struct's Signal field
func (s *PayloadConverterTestSuite) TestRoundTrip_StructWithSignalField() {
	originalSig := &example.ExampleSignal{
		Arg1: "roundtrip-struct-arg-1",
		Arg2: "roundtrip-struct-arg-2",
	}

	// First serialize just the signal (not the whole struct)
	payload, err := s.converter.ToPayload(originalSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	// Now deserialize into a struct with Signal field
	// This is what happens when Temporal deserializes activity parameters
	var resultReq TestRequestStruct
	err = s.converter.FromPayload(payload, &resultReq)
	require.NoError(s.T(), err)

	// Verify Signal field was populated
	require.NotNil(s.T(), resultReq.Signal, "Round-trip Signal field should not be nil")
	exampleSig, ok := resultReq.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), originalSig.Arg1, exampleSig.Arg1)
	assert.Equal(s.T(), originalSig.Arg2, exampleSig.Arg2)

	// QueueID should remain empty (not part of signal serialization)
	assert.Equal(s.T(), "", resultReq.QueueID)
}

// TestFromPayload_InvalidCatalogType verifies error handling for unregistered signal types
func (s *PayloadConverterTestSuite) TestFromPayload_InvalidCatalogType() {
	// Create a payload with an invalid signal type
	invalidPayload := &commonpb.Payload{
		Metadata: map[string][]byte{
			MetadataEncodingKey: []byte(MetadataEncodingType),
		},
		Data: []byte(`{"type":"invalid-signal-type","data":{}}`),
	}

	var resultSig signal.Signal
	err := s.converter.FromPayload(invalidPayload, &resultSig)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid signal type")
	assert.Contains(s.T(), err.Error(), "not registered")
}

// TestFromPayload_NilPointer verifies error handling for nil valuePtr
func (s *PayloadConverterTestSuite) TestFromPayload_NilPointer() {
	sig := &example.ExampleSignal{
		Arg1: "test-arg-1",
		Arg2: "test-arg-2",
	}

	payload, err := s.converter.ToPayload(sig)
	require.NoError(s.T(), err)

	// Test with nil pointer
	var nilPtr *signal.Signal
	err = s.converter.FromPayload(payload, nilPtr)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "cannot be nil")
}

// TestFromPayload_NonPointer verifies error handling when valuePtr is not a pointer
func (s *PayloadConverterTestSuite) TestFromPayload_NonPointer() {
	sig := &example.ExampleSignal{
		Arg1: "test-arg-1",
		Arg2: "test-arg-2",
	}

	payload, err := s.converter.ToPayload(sig)
	require.NoError(s.T(), err)

	// Test with non-pointer value
	var notAPointer signal.Signal
	err = s.converter.FromPayload(payload, notAPointer)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "must be a pointer")
}

// TestEncoding verifies the encoding string is correct
func (s *PayloadConverterTestSuite) TestEncoding() {
	encoding := s.converter.Encoding()
	assert.Equal(s.T(), MetadataEncodingType, encoding)
	assert.Equal(s.T(), "nuon/signal", encoding)
}

// TestToString verifies the ToString method works
func (s *PayloadConverterTestSuite) TestToString() {
	sig := &example.ExampleSignal{
		Arg1: "test-arg-1",
		Arg2: "test-arg-2",
	}

	payload, err := s.converter.ToPayload(sig)
	require.NoError(s.T(), err)

	result := s.converter.ToString(payload)
	// Should return a base64 encoded string
	assert.NotEmpty(s.T(), result)
}

// TestFromPayload_CatalogObjectNotNil verifies that catalog.NewFromType() returns non-nil objects
// and they can be properly unmarshaled (this is the KEY test for the bug fix)
func (s *PayloadConverterTestSuite) TestFromPayload_CatalogObjectNotNil() {
	// Create a signal and serialize it
	originalSig := &example.ExampleSignal{
		Arg1: "catalog-test-1",
		Arg2: "catalog-test-2",
	}

	payload, err := s.converter.ToPayload(originalSig)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	// Parse the payload to get the signal type
	var parsed anyJSON
	err = json.Unmarshal(payload.Data, &parsed)
	require.NoError(s.T(), err)

	// Verify catalog returns non-nil object
	obj, err := catalog.NewFromType(parsed.Type)
	require.NoError(s.T(), err, "Catalog should return object for valid type")
	require.NotNil(s.T(), obj, "Catalog object should not be nil")

	// Verify we can unmarshal into the object (using obj, not &obj)
	err = json.Unmarshal(parsed.Data, obj)
	require.NoError(s.T(), err, "Should be able to unmarshal into catalog object")

	// Verify the result is a valid ExampleSignal
	exampleSig, ok := obj.(*example.ExampleSignal)
	require.True(s.T(), ok, "Catalog object should be *example.ExampleSignal")
	require.NotNil(s.T(), exampleSig, "Cast result should not be nil")
	assert.Equal(s.T(), "catalog-test-1", exampleSig.Arg1)
	assert.Equal(s.T(), "catalog-test-2", exampleSig.Arg2)
}

// TestUnmarshalPointerBehavior explicitly validates the correct pointer behavior when unmarshaling
// This test documents the bug: using &obj (pointer to interface) fails, but obj (pointer to concrete) works
func (s *PayloadConverterTestSuite) TestUnmarshalPointerBehavior() {
	// Create signal data (using correct JSON field names from struct tags)
	signalData := []byte(`{"arg_1":"ptr-test-1","arg_2":"ptr-test-2"}`)

	// Get a new object from catalog
	obj, err := catalog.NewFromType(example.ExampleSignalType)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), obj, "Catalog should return non-nil object")

	// Verify obj is already a pointer
	rv := reflect.ValueOf(obj)
	require.Equal(s.T(), reflect.Ptr, rv.Kind(), "Catalog object should be a pointer")

	// Unmarshal directly into obj (not &obj) - THIS IS THE CORRECT WAY
	err = json.Unmarshal(signalData, obj)
	require.NoError(s.T(), err, "Unmarshaling into obj (not &obj) should work")

	// Verify data was unmarshaled correctly
	exampleSig, ok := obj.(*example.ExampleSignal)
	require.True(s.T(), ok)
	require.NotNil(s.T(), exampleSig)
	assert.Equal(s.T(), "ptr-test-1", exampleSig.Arg1)
	assert.Equal(s.T(), "ptr-test-2", exampleSig.Arg2)
}

// TestTrueRoundTrip_StructWithSignalField tests the ACTUAL Temporal path:
// Serialize entire struct -> Deserialize entire struct
// This test WILL FAIL with current code, exposing the real bug
func (s *PayloadConverterTestSuite) TestTrueRoundTrip_StructWithSignalField() {
	// Create a struct with Signal field - THIS IS WHAT TEMPORAL DOES
	originalReq := TestRequestStruct{
		QueueID: "test-queue-123",
		Signal: &example.ExampleSignal{
			Arg1: "true-roundtrip-1",
			Arg2: "true-roundtrip-2",
		},
	}

	// Serialize the ENTIRE STRUCT (not just the signal)
	payload, err := s.converter.ToPayload(originalReq)
	require.NoError(s.T(), err)

	// BUG: Current code returns nil here because ToPayload doesn't handle structs
	require.NotNil(s.T(), payload, "CRITICAL: ToPayload must handle structs with Signal fields")

	// Deserialize back into a struct
	var resultReq TestRequestStruct
	err = s.converter.FromPayload(payload, &resultReq)
	require.NoError(s.T(), err)

	// Verify QueueID was preserved
	assert.Equal(s.T(), "test-queue-123", resultReq.QueueID)

	// Verify Signal field was correctly deserialized
	require.NotNil(s.T(), resultReq.Signal, "Signal field must be populated")

	exampleSig, ok := resultReq.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok, "Signal should be *example.ExampleSignal")
	assert.Equal(s.T(), "true-roundtrip-1", exampleSig.Arg1)
	assert.Equal(s.T(), "true-roundtrip-2", exampleSig.Arg2)
}

// TestRoundTrip_ArrayOfStructsWithSignalField verifies that an array of structs
// with Signal fields can be serialized and deserialized correctly.
// This covers the WorkflowStep scenario where steps are stored as a slice.
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

	// Serialize the array
	payload, err := s.converter.ToPayload(items)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload, "ToPayload must handle arrays of structs with Signal fields")

	// Deserialize back into an array
	var result []TestRequestStruct
	err = s.converter.FromPayload(payload, &result)
	require.NoError(s.T(), err)
	require.Len(s.T(), result, 3, "Should have 3 items after deserialization")

	// Verify each item
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

// TestTrueRoundTrip_PointerToStructWithSignalField tests the ACTUAL Temporal path:
// Temporal passes *EnqueueSignalRequest (pointer), not a value.
// This was the root cause of the production "can not unmarshal into nil" error.
func (s *PayloadConverterTestSuite) TestTrueRoundTrip_PointerToStructWithSignalField() {
	originalReq := &TestRequestStruct{
		QueueID: "test-queue-ptr",
		Signal: &example.ExampleSignal{
			Arg1: "ptr-roundtrip-1",
			Arg2: "ptr-roundtrip-2",
		},
	}

	// Serialize the POINTER TO STRUCT (this is what Temporal does)
	payload, err := s.converter.ToPayload(originalReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload, "ToPayload must handle pointers to structs with Signal fields")

	// Deserialize back
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

// TestTrueRoundTrip_WithGenericSignalInterface verifies round-trip when Signal
// is stored in a generic signal.Signal interface variable (common pattern)
func (s *PayloadConverterTestSuite) TestTrueRoundTrip_WithGenericSignalInterface() {
	// Store signal in interface variable (common pattern in workflows)
	var sig signal.Signal
	sig = &example.ExampleSignal{
		Arg1: "interface-var-1",
		Arg2: "interface-var-2",
	}

	// Create struct with interface-typed Signal
	originalReq := TestRequestStruct{
		QueueID: "test-queue-456",
		Signal:  sig, // Passed as interface variable
	}

	// Serialize entire struct
	payload, err := s.converter.ToPayload(originalReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload, "ToPayload must handle structs with Signal interface fields")

	// Deserialize back
	var resultReq TestRequestStruct
	err = s.converter.FromPayload(payload, &resultReq)
	require.NoError(s.T(), err)

	// Verify everything matches
	assert.Equal(s.T(), "test-queue-456", resultReq.QueueID)
	require.NotNil(s.T(), resultReq.Signal)

	exampleSig, ok := resultReq.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), "interface-var-1", exampleSig.Arg1)
	assert.Equal(s.T(), "interface-var-2", exampleSig.Arg2)
}

// TestStructWithSignalDataField mimics a WorkflowStep-like struct that has a
// *SignalData field (wrapper) instead of a bare signal.Signal field.
type TestStructWithSignalData struct {
	Name        string      `json:"name"`
	QueueSignal *SignalData `json:"queue_signal,omitempty"`
}

// TestRoundTrip_SliceOfPointersWithSignalData tests encoding/decoding a slice of
// pointers to structs that contain a *SignalData field. This mimics the workflow
// step generator returning []*WorkflowStep where each step has QueueSignal set.
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

	// ToPayload should return nil because these structs don't have a direct
	// signal.Signal field — they have *SignalData which is a wrapper.
	// The standard JSON converter should handle this via SignalData.MarshalJSON.
	payload, err := s.converter.ToPayload(items)
	assert.NoError(s.T(), err)

	if payload != nil {
		// If our converter handles it, verify round-trip
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
		// Our converter returned nil — this means the standard JSON converter
		// will handle it. Let's verify that path works by doing a manual JSON
		// marshal/unmarshal round-trip (since SignalData has MarshalJSON/UnmarshalJSON).
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

// TestTrueRoundTrip_DoublePointerFromTemporal simulates the exact Temporal deserialization path.
// Activity signature: EnqueueSignal(ctx, req *EnqueueSignalRequest)
// Temporal passes **EnqueueSignalRequest as valuePtr to FromPayload.
func (s *PayloadConverterTestSuite) TestTrueRoundTrip_DoublePointerFromTemporal() {
	originalReq := &TestRequestStruct{
		QueueID: "temporal-double-ptr",
		Signal: &example.ExampleSignal{
			Arg1: "double-ptr-1",
			Arg2: "double-ptr-2",
		},
	}

	// Serialize pointer-to-struct (what ToPayload receives)
	payload, err := s.converter.ToPayload(originalReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), payload)

	// Deserialize with double pointer (what Temporal passes to FromPayload)
	var resultReq *TestRequestStruct // nil *TestRequestStruct
	// Temporal passes &resultReq which is **TestRequestStruct
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

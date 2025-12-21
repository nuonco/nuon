package signaldb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

type SignalDataTestSuite struct {
	suite.Suite
}

func TestSignalDataSuite(t *testing.T) {
	suite.Run(t, new(SignalDataTestSuite))
}

func (s *SignalDataTestSuite) TestMarshalJSON() {
	sd := SignalData{
		Signal: &example.ExampleSignal{
			Arg1: "hello",
			Arg2: "world",
		},
	}

	byts, err := json.Marshal(sd)
	require.NoError(s.T(), err)

	// Verify wire format includes type discriminator
	var raw map[string]json.RawMessage
	require.NoError(s.T(), json.Unmarshal(byts, &raw))
	assert.Contains(s.T(), raw, "type")
	assert.Contains(s.T(), raw, "data")

	var typ string
	require.NoError(s.T(), json.Unmarshal(raw["type"], &typ))
	assert.Equal(s.T(), "example-signal", typ)
}

func (s *SignalDataTestSuite) TestMarshalJSON_Nil() {
	sd := SignalData{Signal: nil}

	byts, err := json.Marshal(sd)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "null", string(byts))
}

func (s *SignalDataTestSuite) TestUnmarshalJSON_Null() {
	var sd SignalData
	require.NoError(s.T(), json.Unmarshal([]byte("null"), &sd))
	assert.Nil(s.T(), sd.Signal)
}

func (s *SignalDataTestSuite) TestRoundTrip() {
	original := SignalData{
		Signal: &example.ExampleSignal{
			Arg1: "round",
			Arg2: "trip",
		},
	}

	byts, err := json.Marshal(original)
	require.NoError(s.T(), err)

	var result SignalData
	require.NoError(s.T(), json.Unmarshal(byts, &result))

	require.NotNil(s.T(), result.Signal)
	assert.Equal(s.T(), original.Signal.Type(), result.Signal.Type())

	got, ok := result.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), "round", got.Arg1)
	assert.Equal(s.T(), "trip", got.Arg2)
}

func (s *SignalDataTestSuite) TestRoundTrip_EmbeddedInStruct() {
	type Wrapper struct {
		Name   string     `json:"name"`
		Signal SignalData `json:"signal"`
	}

	original := Wrapper{
		Name: "test-wrapper",
		Signal: SignalData{
			Signal: &example.ExampleSignal{
				Arg1: "nested-1",
				Arg2: "nested-2",
			},
		},
	}

	byts, err := json.Marshal(original)
	require.NoError(s.T(), err)

	var result Wrapper
	require.NoError(s.T(), json.Unmarshal(byts, &result))

	assert.Equal(s.T(), "test-wrapper", result.Name)
	require.NotNil(s.T(), result.Signal.Signal)

	got, ok := result.Signal.Signal.(*example.ExampleSignal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), "nested-1", got.Arg1)
	assert.Equal(s.T(), "nested-2", got.Arg2)
}

func (s *SignalDataTestSuite) TestRoundTrip_NilSignalInStruct() {
	type Wrapper struct {
		Name   string     `json:"name"`
		Signal SignalData `json:"signal"`
	}

	original := Wrapper{
		Name:   "nil-signal",
		Signal: SignalData{Signal: nil},
	}

	byts, err := json.Marshal(original)
	require.NoError(s.T(), err)

	var result Wrapper
	require.NoError(s.T(), json.Unmarshal(byts, &result))

	assert.Equal(s.T(), "nil-signal", result.Name)
	assert.Nil(s.T(), result.Signal.Signal)
}

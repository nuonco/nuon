package example

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type ExampleSignalTestSuite struct {
	suite.Suite
}

func TestExampleSignalSuite(t *testing.T) {
	suite.Run(t, new(ExampleSignalTestSuite))
}

func (s *ExampleSignalTestSuite) TestType() {
	sig := &ExampleSignal{}
	assert.Equal(s.T(), ExampleSignalType, sig.Type())
	assert.Equal(s.T(), signal.SignalType("example-signal"), sig.Type())
}

func (s *ExampleSignalTestSuite) TestImplementsSignalInterface() {
	var _ signal.Signal = (*ExampleSignal)(nil)
}

func (s *ExampleSignalTestSuite) TestRegisteredInCatalog() {
	sig, err := catalog.NewFromType(ExampleSignalType)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), sig)

	_, ok := sig.(*ExampleSignal)
	assert.True(s.T(), ok)
}

package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// testSignal is a local signal type to avoid import cycles with the example package
type testSignal struct {
	signal.Hooks
	Value string
}

func (s *testSignal) Type() signal.SignalType             { return "test-catalog-signal" }
func (s *testSignal) Validate(ctx workflow.Context) error { return nil }
func (s *testSignal) Execute(ctx workflow.Context) error  { return nil }

const testSignalType signal.SignalType = "test-catalog-signal"

type CatalogTestSuite struct {
	suite.Suite
}

func TestCatalogSuite(t *testing.T) {
	suite.Run(t, new(CatalogTestSuite))
}

func (s *CatalogTestSuite) SetupTest() {
	// Register our test signal for each test
	Register(testSignalType, func() signal.Signal {
		return &testSignal{}
	})
}

func (s *CatalogTestSuite) TearDownTest() {
	delete(SignalCatalog, testSignalType)
}

func (s *CatalogTestSuite) TestNewFromType_RegisteredType() {
	sig, err := NewFromType(testSignalType)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), sig)
	assert.Equal(s.T(), testSignalType, sig.Type())
}

func (s *CatalogTestSuite) TestNewFromType_UnregisteredType() {
	sig, err := NewFromType(signal.SignalType("nonexistent-signal"))
	require.Error(s.T(), err)
	assert.Nil(s.T(), sig)
	assert.Contains(s.T(), err.Error(), "not registered")
}

func (s *CatalogTestSuite) TestNewFromType_ReturnsNewInstanceEachTime() {
	sig1, err := NewFromType(testSignalType)
	require.NoError(s.T(), err)

	sig2, err := NewFromType(testSignalType)
	require.NoError(s.T(), err)

	// Different instances
	assert.NotSame(s.T(), sig1, sig2)
}

func (s *CatalogTestSuite) TestRegister_WithDefaultValues() {
	const customType signal.SignalType = "test-catalog-custom"

	Register(customType, func() signal.Signal {
		return &testSignal{Value: "default-value"}
	})
	defer delete(SignalCatalog, customType)

	sig, err := NewFromType(customType)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), sig)

	ts, ok := sig.(*testSignal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), "default-value", ts.Value)
}

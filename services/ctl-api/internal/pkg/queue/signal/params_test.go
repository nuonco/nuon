package signal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// testSignalWithParams implements both Signal and SignalWithParams
type testSignalWithParams struct {
	params *Params
}

func (s *testSignalWithParams) Type() SignalType                { return "test-with-params" }
func (s *testSignalWithParams) Validate(workflow.Context) error { return nil }
func (s *testSignalWithParams) Execute(workflow.Context) error  { return nil }
func (s *testSignalWithParams) WithParams(p *Params)            { s.params = p }

// testSignalWithoutParams implements only Signal (no SignalWithParams)
type testSignalWithoutParams struct{}

func (s *testSignalWithoutParams) Type() SignalType                { return "test-without-params" }
func (s *testSignalWithoutParams) Validate(workflow.Context) error { return nil }
func (s *testSignalWithoutParams) Execute(workflow.Context) error  { return nil }

type ParamsTestSuite struct {
	suite.Suite
}

func TestParamsSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (s *ParamsTestSuite) TestApplyParams_WithSignalWithParams() {
	sig := &testSignalWithParams{}
	cfg := &internal.Config{}
	params := &Params{Cfg: cfg}

	ApplyParams(sig, params)

	assert.NotNil(s.T(), sig.params)
	assert.Same(s.T(), cfg, sig.params.Cfg)
}

func (s *ParamsTestSuite) TestApplyParams_WithoutSignalWithParams_NoPanic() {
	sig := &testSignalWithoutParams{}
	params := &Params{Cfg: &internal.Config{}}

	// Should not panic
	assert.NotPanics(s.T(), func() {
		ApplyParams(sig, params)
	})
}

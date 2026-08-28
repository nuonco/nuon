package customermanaged

import shared "github.com/nuonco/nuon/pkg/customer_managed"

type Envelope = shared.Envelope
type OutputBinding = shared.OutputBinding
type ActionTemplate = shared.ActionTemplate
type DriftTemplate = shared.DriftTemplate
type RunbookTemplate = shared.RunbookTemplate
type RunbookStep = shared.RunbookStep
type ComponentSpec = shared.ComponentSpec
type InputSpec = shared.InputSpec
type Step = shared.Step

const (
	RunbookStepKindAction     = shared.RunbookStepKindAction
	RunbookStepKindDrift      = shared.RunbookStepKindDrift
	RunbookStepKindHealthGate = shared.RunbookStepKindHealthGate
)

func Load(path string) (*Envelope, error) {
	return shared.Load(path)
}

func Parse(data []byte) (*Envelope, error) {
	return shared.Parse(data)
}

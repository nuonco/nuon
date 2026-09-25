package app

type StepChangeState string

const (
	StepChangeStateUnknown     StepChangeState = ""
	StepChangeStateOK          StepChangeState = "ok"
	StepChangeStateUnsupported StepChangeState = "unsupported"
	StepChangeStateError       StepChangeState = "error"
)

type StepChangeCounts struct {
	Create  int `json:"create"`
	Update  int `json:"update"`
	Delete  int `json:"delete"`
	Replace int `json:"replace"`
	Noop    int `json:"noop"`
}

func (c StepChangeCounts) HasChanges() bool {
	return c.Create > 0 || c.Update > 0 || c.Delete > 0 || c.Replace > 0
}

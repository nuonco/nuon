package state

func NewInputsState() *InputsState {
	return &InputsState{
		Inputs: make(map[string]string, 0),
	}
}

type InputsState struct {
	Populated bool              `json:"populated"`
	Inputs    map[string]string `json:"inputs"`
}

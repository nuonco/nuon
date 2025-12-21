package state

func NewOrgState() *OrgState {
	return &OrgState{}
}

type OrgState struct {
	Status    string `json:"status"`
	Populated bool   `json:"populated"`
	ID        string `json:"id"`
	Name      string `json:"name"`
}

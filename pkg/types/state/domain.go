package state

func NewDomainState() *DomainState {
	return &DomainState{}
}

type DomainState struct {
	Populated bool `json:"populated"`

	PublicDomain   string `json:"public_domain"`
	InternalDomain string `json:"internal_domain"`
}

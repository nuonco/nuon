package api

type APIContextType string

const (
	APIContextTypePublic   APIContextType = "public"
	APIContextTypeRunner   APIContextType = "runner"
	APIContextTypeInternal APIContextType = "internal"
)

func (a APIContextType) String() string {
	return string(a)
}

package core

// Kind identifies the operation a stack run represents. Mirrors the ctl-api
// `app.InstallStackVersionRunKind` enum.
type Kind string

const (
	KindProvision   Kind = "provision"
	KindReprovision Kind = "reprovision"
	KindDeprovision Kind = "deprovision"
)

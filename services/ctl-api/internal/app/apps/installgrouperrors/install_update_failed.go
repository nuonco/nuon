package installgrouperrors

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const InstallUpdateFailedType compositeerrors.Type = "install_group.install_update_failed"

type InstallUpdateFailedError struct {
	InstallID   string `json:"install_id"`
	InstallName string `json:"install_name,omitempty"`
	WorkflowID  string `json:"workflow_id,omitempty"`
	Detail      string `json:"detail,omitempty"`
}

var _ compositeerrors.CompositeError = (*InstallUpdateFailedError)(nil)

func (e *InstallUpdateFailedError) Error() string {
	name := e.InstallName
	if name == "" {
		name = e.InstallID
	}
	if name == "" {
		return "An install failed during deploy"
	}
	return fmt.Sprintf("%s failed during deploy", name)
}

func (e *InstallUpdateFailedError) Type() compositeerrors.Type {
	return InstallUpdateFailedType
}

func (e *InstallUpdateFailedError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityError
}

func (e *InstallUpdateFailedError) Sections() []compositeerrors.Section {
	if e.Detail == "" {
		return nil
	}
	return []compositeerrors.Section{
		compositeerrors.TextSection("What happened", e.Detail),
	}
}

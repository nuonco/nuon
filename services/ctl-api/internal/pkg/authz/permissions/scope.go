package permissions

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	KindRunner         ResourceKind = "runner"
	KindLogStream      ResourceKind = "log_stream"
	KindExecutionState ResourceKind = "execution_state"
	KindEvent          ResourceKind = "event"
	KindNotification   ResourceKind = "notification"
	KindOperation      ResourceKind = "operation"
)

type Scope struct {
	Kind  ResourceKind `json:"kind"`
	OrgID string       `json:"org_id"`
	ID    string       `json:"id"`
}

type ResolvedTarget struct {
	Scope
	AppID     string
	InstallID string
}

func validScopeID(value string) bool {
	return value != "" && !strings.ContainsAny(value, "*:/") && strings.IndexFunc(value, unicode.IsSpace) == -1
}

func (s Scope) ValidateShape() error {
	if !validScopeID(s.OrgID) || !validScopeID(s.ID) {
		return fmt.Errorf("scope requires explicit org and resource IDs")
	}
	switch s.Kind {
	case KindOrg:
		if s.ID != s.OrgID {
			return fmt.Errorf("org scope ID must match its org")
		}
	case KindApp, KindInstall, KindStack, KindTelemetry, KindPhoneHome, KindRunnerJob,
		KindRunner, KindLogStream, KindExecutionState, KindEvent, KindNotification, KindOperation:
	default:
		return fmt.Errorf("invalid scope kind %q", s.Kind)
	}
	return nil
}

func (t ResolvedTarget) ValidateShape() error {
	if err := t.Scope.ValidateShape(); err != nil {
		return err
	}
	if t.AppID != "" && !validScopeID(t.AppID) {
		return fmt.Errorf("invalid target app ID")
	}
	if t.InstallID != "" && (!validScopeID(t.InstallID) || t.AppID == "") {
		return fmt.Errorf("install target requires explicit install and app IDs")
	}
	switch t.Kind {
	case KindOrg:
		if t.AppID != "" || t.InstallID != "" {
			return fmt.Errorf("org target cannot have app or install parents")
		}
	case KindApp:
		if t.AppID != t.ID || t.InstallID != "" {
			return fmt.Errorf("app target must identify itself without an install parent")
		}
	case KindInstall:
		if t.InstallID != t.ID || t.AppID == "" {
			return fmt.Errorf("install target must identify itself and its app parent")
		}
	}
	return nil
}

func (s Scope) covers(target ResolvedTarget) bool {
	if s.OrgID != target.OrgID {
		return false
	}
	switch s.Kind {
	case KindOrg:
		return true
	case KindApp:
		return s.ID == target.AppID
	case KindInstall:
		return s.ID == target.InstallID
	default:
		return s.Kind == target.Kind && s.ID == target.ID
	}
}

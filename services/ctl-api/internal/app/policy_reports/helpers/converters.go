package helpers

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// PolicyViolationDisplay is a display-friendly representation of a policy violation.
type PolicyViolationDisplay struct {
	PolicyID      string
	PolicyName    string
	Message       string
	Severity      string
	InputIndex    int
	InputIdentity string
}

// PolicyResultDisplay is a display-friendly representation of a policy result.
type PolicyResultDisplay struct {
	PolicyID   string
	PolicyName string
	Status     string
	DenyCount  int
	WarnCount  int
	PassCount  int
	InputCount int
}

// PolicyInputDisplay is a display-friendly representation of a policy input reference.
type PolicyInputDisplay struct {
	ID   string
	Name string
	Type string
}

// ToViolationDisplay converts an app.PolicyViolation to a display representation.
func ToViolationDisplay(v app.PolicyViolation) PolicyViolationDisplay {
	return PolicyViolationDisplay{
		PolicyID:      v.PolicyID,
		PolicyName:    v.PolicyName,
		Message:       v.Message,
		Severity:      v.Severity,
		InputIndex:    v.InputIndex,
		InputIdentity: v.InputIdentity,
	}
}

// ToViolationDisplays converts a slice of app.PolicyViolation to display representations.
func ToViolationDisplays(violations []app.PolicyViolation) []PolicyViolationDisplay {
	result := make([]PolicyViolationDisplay, len(violations))
	for i, v := range violations {
		result[i] = ToViolationDisplay(v)
	}
	return result
}

// ToAppViolation converts a PolicyViolationDisplay back to an app.PolicyViolation.
func ToAppViolation(v PolicyViolationDisplay) app.PolicyViolation {
	return app.PolicyViolation{
		PolicyID:      v.PolicyID,
		PolicyName:    v.PolicyName,
		InputIndex:    v.InputIndex,
		InputIdentity: v.InputIdentity,
		Message:       v.Message,
		Severity:      v.Severity,
	}
}

// ToAppViolations converts a slice of PolicyViolationDisplay to app.PolicyViolation.
func ToAppViolations(violations []PolicyViolationDisplay) []app.PolicyViolation {
	result := make([]app.PolicyViolation, len(violations))
	for i, v := range violations {
		result[i] = ToAppViolation(v)
	}
	return result
}

// ToResultDisplay converts an app.PolicyResult to a display representation.
func ToResultDisplay(r app.PolicyResult) PolicyResultDisplay {
	return PolicyResultDisplay{
		PolicyID:   r.PolicyID,
		PolicyName: r.PolicyName,
		Status:     r.Status,
		DenyCount:  r.DenyCount,
		WarnCount:  r.WarnCount,
		PassCount:  r.PassCount,
		InputCount: r.InputCount,
	}
}

// ToResultDisplays converts a slice of app.PolicyResult to display representations.
func ToResultDisplays(results []app.PolicyResult) []PolicyResultDisplay {
	result := make([]PolicyResultDisplay, len(results))
	for i, r := range results {
		result[i] = ToResultDisplay(r)
	}
	return result
}

// ToAppResult converts a PolicyResultDisplay back to an app.PolicyResult.
func ToAppResult(r PolicyResultDisplay) app.PolicyResult {
	return app.PolicyResult{
		PolicyID:   r.PolicyID,
		PolicyName: r.PolicyName,
		Status:     r.Status,
		DenyCount:  r.DenyCount,
		WarnCount:  r.WarnCount,
		PassCount:  r.PassCount,
		InputCount: r.InputCount,
	}
}

// ToAppResults converts a slice of PolicyResultDisplay to app.PolicyResult.
func ToAppResults(results []PolicyResultDisplay) []app.PolicyResult {
	result := make([]app.PolicyResult, len(results))
	for i, r := range results {
		result[i] = ToAppResult(r)
	}
	return result
}

// ToInputDisplay converts an app.PolicyInputRef to a display representation.
func ToInputDisplay(inp app.PolicyInputRef) PolicyInputDisplay {
	return PolicyInputDisplay{
		ID:   inp.ID,
		Type: inp.Type,
		Name: inp.Name,
	}
}

// ToInputDisplays converts a slice of app.PolicyInputRef to display representations.
func ToInputDisplays(inputs []app.PolicyInputRef) []PolicyInputDisplay {
	result := make([]PolicyInputDisplay, len(inputs))
	for i, inp := range inputs {
		result[i] = ToInputDisplay(inp)
	}
	return result
}

// ToAppInputRef converts a PolicyInputDisplay back to an app.PolicyInputRef.
func ToAppInputRef(inp PolicyInputDisplay) app.PolicyInputRef {
	return app.PolicyInputRef{
		ID:   inp.ID,
		Type: inp.Type,
		Name: inp.Name,
	}
}

// ToAppInputRefs converts a slice of PolicyInputDisplay to app.PolicyInputRef.
func ToAppInputRefs(inputs []PolicyInputDisplay) []app.PolicyInputRef {
	result := make([]app.PolicyInputRef, len(inputs))
	for i, inp := range inputs {
		result[i] = ToAppInputRef(inp)
	}
	return result
}

// PolicyResultInternal is the internal representation used during policy evaluation.
type PolicyResultInternal struct {
	PolicyID   string
	Status     string
	DenyCount  int
	WarnCount  int
	PassCount  int
	InputCount int
}

// ToAppResultFromInternal converts a PolicyResultInternal to an app.PolicyResult.
func ToAppResultFromInternal(r PolicyResultInternal) app.PolicyResult {
	return app.PolicyResult{
		PolicyID:   r.PolicyID,
		Status:     r.Status,
		DenyCount:  r.DenyCount,
		WarnCount:  r.WarnCount,
		PassCount:  r.PassCount,
		InputCount: r.InputCount,
	}
}

// ToAppResultsFromInternal converts a slice of PolicyResultInternal to app.PolicyResult.
func ToAppResultsFromInternal(results []PolicyResultInternal) []app.PolicyResult {
	result := make([]app.PolicyResult, len(results))
	for i, r := range results {
		result[i] = ToAppResultFromInternal(r)
	}
	return result
}

// PolicyInputRefInternal is the internal representation of a policy input reference.
type PolicyInputRefInternal struct {
	ID   string
	Type string
	Name string
}

// ToAppInputRefFromInternal converts a PolicyInputRefInternal to an app.PolicyInputRef.
func ToAppInputRefFromInternal(inp PolicyInputRefInternal) app.PolicyInputRef {
	return app.PolicyInputRef{
		ID:   inp.ID,
		Type: inp.Type,
		Name: inp.Name,
	}
}

// ToAppInputRefsFromInternal converts a slice of PolicyInputRefInternal to app.PolicyInputRef.
func ToAppInputRefsFromInternal(inputs []PolicyInputRefInternal) []app.PolicyInputRef {
	result := make([]app.PolicyInputRef, len(inputs))
	for i, inp := range inputs {
		result[i] = ToAppInputRefFromInternal(inp)
	}
	return result
}

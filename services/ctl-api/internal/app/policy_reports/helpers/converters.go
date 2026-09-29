package helpers

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type PolicyViolationDisplay struct {
	PolicyID      string
	PolicyName    string
	Message       string
	Severity      string
	InputIndex    int
	InputIdentity string
}

type PolicyResultDisplay struct {
	PolicyID   string
	PolicyName string
	Status     string
	DenyCount  int
	WarnCount  int
	PassCount  int
	InputCount int
}

type PolicyInputDisplay struct {
	ID   string
	Name string
	Type string
}

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

func ToViolationDisplays(violations []app.PolicyViolation) []PolicyViolationDisplay {
	result := make([]PolicyViolationDisplay, len(violations))
	for i, v := range violations {
		result[i] = ToViolationDisplay(v)
	}
	return result
}

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

func ToAppViolations(violations []PolicyViolationDisplay) []app.PolicyViolation {
	result := make([]app.PolicyViolation, len(violations))
	for i, v := range violations {
		result[i] = ToAppViolation(v)
	}
	return result
}

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

func ToResultDisplays(results []app.PolicyResult) []PolicyResultDisplay {
	result := make([]PolicyResultDisplay, len(results))
	for i, r := range results {
		result[i] = ToResultDisplay(r)
	}
	return result
}

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

func ToAppResults(results []PolicyResultDisplay) []app.PolicyResult {
	result := make([]app.PolicyResult, len(results))
	for i, r := range results {
		result[i] = ToAppResult(r)
	}
	return result
}

func ToInputDisplay(inp app.PolicyInputRef) PolicyInputDisplay {
	return PolicyInputDisplay{
		ID:   inp.ID,
		Type: inp.Type,
		Name: inp.Name,
	}
}

func ToInputDisplays(inputs []app.PolicyInputRef) []PolicyInputDisplay {
	result := make([]PolicyInputDisplay, len(inputs))
	for i, inp := range inputs {
		result[i] = ToInputDisplay(inp)
	}
	return result
}

func ToAppInputRef(inp PolicyInputDisplay) app.PolicyInputRef {
	return app.PolicyInputRef{
		ID:   inp.ID,
		Type: inp.Type,
		Name: inp.Name,
	}
}

func ToAppInputRefs(inputs []PolicyInputDisplay) []app.PolicyInputRef {
	result := make([]app.PolicyInputRef, len(inputs))
	for i, inp := range inputs {
		result[i] = ToAppInputRef(inp)
	}
	return result
}

type PolicyResultInternal struct {
	PolicyID   string
	Status     string
	DenyCount  int
	WarnCount  int
	PassCount  int
	InputCount int
}

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

func ToAppResultsFromInternal(results []PolicyResultInternal) []app.PolicyResult {
	result := make([]app.PolicyResult, len(results))
	for i, r := range results {
		result[i] = ToAppResultFromInternal(r)
	}
	return result
}

type PolicyInputRefInternal struct {
	ID   string
	Type string
	Name string
}

func ToAppInputRefFromInternal(inp PolicyInputRefInternal) app.PolicyInputRef {
	return app.PolicyInputRef{
		ID:   inp.ID,
		Type: inp.Type,
		Name: inp.Name,
	}
}

func ToAppInputRefsFromInternal(inputs []PolicyInputRefInternal) []app.PolicyInputRef {
	result := make([]app.PolicyInputRef, len(inputs))
	for i, inp := range inputs {
		result[i] = ToAppInputRefFromInternal(inp)
	}
	return result
}

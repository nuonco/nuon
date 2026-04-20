package templates

// Template represents a sandbox mode template that can be used to populate
// runner job logs, plans, or outputs in sandbox mode.
type Template struct {
	Key         string   `json:"key"`
	Description string   `json:"description"`
	Category    string   `json:"category"`  // "logs", "plans", "outputs"
	JobTypes    []string `json:"job_types"` // which runner job types this template applies to
	Contents    string   `json:"contents"`  // the actual template content
	IsNoop      bool     `json:"is_noop"`   // whether this is a noop template
}

// AllTemplates returns all available templates across all categories.
func AllTemplates() []Template {
	var all []Template
	all = append(all, LogTemplates()...)
	all = append(all, PlanTemplates()...)
	all = append(all, PlanDisplayTemplates()...)
	all = append(all, StateTemplates()...)
	all = append(all, OutputTemplates()...)
	return all
}

// LogTemplates returns all log templates.
func LogTemplates() []Template {
	return logTemplates()
}

// PlanTemplates returns all plan templates (machine-readable JSON).
func PlanTemplates() []Template {
	return planTemplates()
}

// PlanDisplayTemplates returns all human-readable plan display templates.
func PlanDisplayTemplates() []Template {
	return planDisplayTemplates()
}

// StateTemplates returns all terraform state JSON templates.
func StateTemplates() []Template {
	return stateTemplates()
}

// OutputTemplates returns all output templates.
func OutputTemplates() []Template {
	return outputTemplates()
}

// TemplatesForJobType returns all templates that apply to the given runner job type.
func TemplatesForJobType(jobType string) []Template {
	var result []Template
	for _, t := range AllTemplates() {
		for _, jt := range t.JobTypes {
			if jt == jobType {
				result = append(result, t)
				break
			}
		}
	}
	return result
}

// NoopTemplates returns all templates marked as noop.
func NoopTemplates() []Template {
	var result []Template
	for _, t := range AllTemplates() {
		if t.IsNoop {
			result = append(result, t)
		}
	}
	return result
}

// FlowTemplate is a pre-built collection that configures multiple runner jobs
// at once to simulate an entire deployment flow.
type FlowTemplate struct {
	Key         string       `json:"key"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	IsNoop      bool         `json:"is_noop"`
	Configs     []FlowConfig `json:"configs"`
}

// FlowConfig defines one runner job config within a flow template.
type FlowConfig struct {
	JobType             string `json:"job_type"`
	LogTemplate         string `json:"log_template,omitempty"`          // key into LogTemplates()
	PlanTemplate        string `json:"plan_template,omitempty"`         // key into PlanTemplates()
	PlanDisplayTemplate string `json:"plan_display_template,omitempty"` // key into PlanDisplayTemplates()
	StateTemplate       string `json:"state_template,omitempty"`        // key into StateTemplates()
	OutputTemplate      string `json:"output_template,omitempty"`       // key into OutputTemplates()
	DurationMs          int64  `json:"duration_ms"`
	Enabled             bool   `json:"enabled"`
}

// FlowTemplates returns all pre-built flow templates.
func FlowTemplates() []FlowTemplate {
	return flowTemplates()
}

// FindFlowTemplate returns the flow template with the given key, or nil.
func FindFlowTemplate(key string) *FlowTemplate {
	for _, ft := range flowTemplates() {
		if ft.Key == key {
			return &ft
		}
	}
	return nil
}

// FindTemplate returns the individual template with the given key, or nil.
func FindTemplate(key string) *Template {
	for _, t := range AllTemplates() {
		if t.Key == key {
			return &t
		}
	}
	return nil
}

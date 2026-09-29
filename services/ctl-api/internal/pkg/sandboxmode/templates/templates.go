package templates

type Template struct {
	Key         string   `json:"key"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	JobTypes    []string `json:"job_types"`
	Contents    string   `json:"contents"`
	IsNoop      bool     `json:"is_noop"`
}

func AllTemplates() []Template {
	var all []Template
	all = append(all, LogTemplates()...)
	all = append(all, PlanTemplates()...)
	all = append(all, PlanDisplayTemplates()...)
	all = append(all, StateTemplates()...)
	all = append(all, OutputTemplates()...)
	return all
}

func LogTemplates() []Template {
	return logTemplates()
}

func PlanTemplates() []Template {
	return planTemplates()
}

func PlanDisplayTemplates() []Template {
	return planDisplayTemplates()
}

func StateTemplates() []Template {
	return stateTemplates()
}

func OutputTemplates() []Template {
	return outputTemplates()
}

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

func NoopTemplates() []Template {
	var result []Template
	for _, t := range AllTemplates() {
		if t.IsNoop {
			result = append(result, t)
		}
	}
	return result
}

type FlowTemplate struct {
	Key         string       `json:"key"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	IsNoop      bool         `json:"is_noop"`
	Configs     []FlowConfig `json:"configs"`
}

type FlowConfig struct {
	JobType             string `json:"job_type"`
	LogTemplate         string `json:"log_template,omitempty"`
	PlanTemplate        string `json:"plan_template,omitempty"`
	PlanDisplayTemplate string `json:"plan_display_template,omitempty"`
	StateTemplate       string `json:"state_template,omitempty"`
	OutputTemplate      string `json:"output_template,omitempty"`
	DurationMs          int64  `json:"duration_ms"`
	Enabled             bool   `json:"enabled"`
}

func FlowTemplates() []FlowTemplate {
	return flowTemplates()
}

func FindFlowTemplate(key string) *FlowTemplate {
	for _, ft := range flowTemplates() {
		if ft.Key == key {
			return &ft
		}
	}
	return nil
}

func FindTemplate(key string) *Template {
	for _, t := range AllTemplates() {
		if t.Key == key {
			return &t
		}
	}
	return nil
}

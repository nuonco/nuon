package customermanaged

import (
	"sort"
	"time"
)

const (
	CapturedInputsKey = "metadata/current-inputs.json"
	CapturedRolesKey  = "metadata/roles.json"
)

type CapturedInputs struct {
	ObservedAt time.Time       `json:"observed_at"`
	Inputs     []CapturedInput `json:"inputs"`
}

type CapturedInput struct {
	Name           string  `json:"name"`
	Type           string  `json:"type,omitempty"`
	Description    string  `json:"description,omitempty"`
	Required       bool    `json:"required,omitempty"`
	Secret         bool    `json:"secret,omitempty"`
	Bindable       bool    `json:"bindable,omitempty"`
	Value          *string `json:"value,omitempty"`
	Default        *string `json:"default,omitempty"`
	ValueStatus    string  `json:"value_status"`
	ValueAvailable bool    `json:"value_available,omitempty"`
}

func CaptureInputs(specs []InputSpec, provided map[string]string, observedAt time.Time) CapturedInputs {
	result := CapturedInputs{ObservedAt: observedAt, Inputs: make([]CapturedInput, 0, len(specs))}
	resolved := ResolveInputValues(specs, provided)
	for _, spec := range specs {
		input := CapturedInput{
			Name: spec.Name, Type: spec.Type, Description: spec.Description,
			Required: spec.Required, Secret: spec.Secret, Bindable: spec.Bindable,
			ValueStatus: "unavailable",
		}
		if spec.Secret {
			input.ValueStatus = "redacted"
			result.Inputs = append(result.Inputs, input)
			continue
		}
		input.Default = stringPointer(spec.Default)
		if !spec.Bindable {
			input.ValueStatus = "embedded-in-bundle"
			result.Inputs = append(result.Inputs, input)
			continue
		}
		if value, ok := resolved[spec.Name]; ok {
			input.Value = stringPointer(value)
			input.ValueAvailable = true
			if _, supplied := provided[spec.Name]; supplied {
				input.ValueStatus = "provided"
			} else {
				input.ValueStatus = "default"
			}
		}
		result.Inputs = append(result.Inputs, input)
	}
	return result
}

func stringPointer(value string) *string { return &value }

type CapturedRoles struct {
	ObservedAt time.Time      `json:"observed_at"`
	Roles      []CapturedRole `json:"roles"`
}

type CapturedRole struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	CloudID     string `json:"cloud_id"`
	Provisioned bool   `json:"provisioned"`
}

type roleOutput struct {
	Name string
	Type string
}

var scalarRoleOutputs = map[string]roleOutput{
	"provision_iam_role_arn":         {Name: "Provision", Type: "provision"},
	"deprovision_iam_role_arn":       {Name: "Deprovision", Type: "deprovision"},
	"maintenance_iam_role_arn":       {Name: "Maintenance", Type: "maintenance"},
	"runner_iam_role_arn":            {Name: "Runner", Type: "runner"},
	"provision_identity_client_id":   {Name: "Provision", Type: "provision"},
	"deprovision_identity_client_id": {Name: "Deprovision", Type: "deprovision"},
	"maintenance_identity_client_id": {Name: "Maintenance", Type: "maintenance"},
	"runner_identity_client_id":      {Name: "Runner", Type: "runner"},
	"provision_sa_email":             {Name: "Provision", Type: "provision"},
	"deprovision_sa_email":           {Name: "Deprovision", Type: "deprovision"},
	"maintenance_sa_email":           {Name: "Maintenance", Type: "maintenance"},
	"runner_service_account_email":   {Name: "Runner", Type: "runner"},
}

var groupedRoleOutputs = map[string]string{
	"break_glass_role_arns":           "break-glass",
	"custom_role_arns":                "custom",
	"break_glass_identity_client_ids": "break-glass",
	"custom_identity_client_ids":      "custom",
	"break_glass_sa_emails":           "break-glass",
	"custom_sa_emails":                "custom",
}

func CaptureRoles(outputs map[string]any, observedAt time.Time) CapturedRoles {
	roles := map[string]CapturedRole{}
	collectRoles(outputs, roles)
	result := CapturedRoles{ObservedAt: observedAt, Roles: make([]CapturedRole, 0, len(roles))}
	for _, role := range roles {
		result.Roles = append(result.Roles, role)
	}
	sort.Slice(result.Roles, func(i, j int) bool {
		if result.Roles[i].Type == result.Roles[j].Type {
			return result.Roles[i].Name < result.Roles[j].Name
		}
		return result.Roles[i].Type < result.Roles[j].Type
	})
	return result
}

func collectRoles(node map[string]any, roles map[string]CapturedRole) {
	for key, value := range node {
		if descriptor, ok := scalarRoleOutputs[key]; ok {
			if cloudID, ok := value.(string); ok && cloudID != "" {
				roles[descriptor.Type+"\x00"+descriptor.Name] = CapturedRole{
					Name: descriptor.Name, Type: descriptor.Type, CloudID: cloudID, Provisioned: true,
				}
			}
			continue
		}
		if roleType, ok := groupedRoleOutputs[key]; ok {
			if values, ok := value.(map[string]any); ok {
				for name, rawID := range values {
					if cloudID, ok := rawID.(string); ok && cloudID != "" {
						roles[roleType+"\x00"+name] = CapturedRole{
							Name: name, Type: roleType, CloudID: cloudID, Provisioned: true,
						}
					}
				}
			}
			continue
		}
		if child, ok := value.(map[string]any); ok {
			collectRoles(child, roles)
		}
	}
}

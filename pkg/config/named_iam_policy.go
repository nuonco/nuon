package config

import (
	"github.com/invopop/jsonschema"
)

// NamedIAMPolicy is a customer-managed IAM policy defined once under
// permissions and attached to any number of roles. It is not a Kyverno/OPA
// AppPolicy (those live in the root policies/ directory).
type NamedIAMPolicy struct {
	// Name is both the config identifier and the AWS IAM managed policy name.
	// Roles attach this policy by repeating the same name, matching inline
	// [[policies]] name = "..." attachments. Supports templating.
	Name        string `json:"Name" mapstructure:"name" toml:"name" jsonschema:"required" features:"template"`
	Description string `json:"Description" mapstructure:"description,omitempty" toml:"description,omitempty" features:"template"`
	Contents    string `json:"Contents" mapstructure:"contents" toml:"contents" jsonschema:"required" features:"template,get"`

	SourceFile string `mapstructure:"-" toml:"-" json:"-" jsonschema:"-"`
}

func (p *NamedIAMPolicy) SetSourceFile(path string) {
	p.SourceFile = path
}

func (p *NamedIAMPolicy) GetSourceFile() string {
	return p.SourceFile
}

func (p NamedIAMPolicy) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("AWS IAM managed policy name").Required().
		Long("Name used for the customer-managed IAM policy in the install account, and the identifier roles use in [[named_policies]]. Supports Nuon templating").
		Example("{{.nuon.install.id}}-alb-create").
		Example("shared-logs-policy").
		Field("description").Short("description of the policy").
		Long("Human-readable description of the named IAM policy").
		Example("Shared CloudWatch Logs access for runner roles").
		Field("contents").Short("IAM policy document").Required().
		Long("JSON policy document for the customer-managed IAM policy. Created in the install stack even when no role is enabled. Supports Nuon templating and external file sources: HTTP(S) URLs, git repositories, file paths, and relative paths (./logs.json)").
		Example("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":\"logs:*\",\"Resource\":\"*\"}]}")
}

// NamedPolicyRef attaches a NamedIAMPolicy to a role by name.
type NamedPolicyRef struct {
	Name string `json:"Name" mapstructure:"name" toml:"name" jsonschema:"required"`
}

func (p NamedPolicyRef) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("named IAM policy name").Required().
		Long("Must match the name of a named IAM policy defined under permissions. Same string as the policy's name field, including any templates").
		Example("{{.nuon.install.id}}-alb-create")
}

func NamedPolicyRefNames(refs []NamedPolicyRef) []string {
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		out = append(out, ref.Name)
	}
	return out
}

func NamedPolicyRefs(names []string) []NamedPolicyRef {
	out := make([]NamedPolicyRef, 0, len(names))
	for _, name := range names {
		out = append(out, NamedPolicyRef{Name: name})
	}
	return out
}

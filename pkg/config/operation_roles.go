package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/nuonco/nuon/pkg/principal"
)

type OperationRuleConfigType string

const (
	OperationRuleConfigTypeMatrix OperationRuleConfigType = "matrix"
)

type PrincipalType = principal.Type

const (
	PrincipalTypeComponent PrincipalType = principal.TypeComponent
	PrincipalTypeSandbox   PrincipalType = principal.TypeSandbox
	PrincipalTypeAction    PrincipalType = principal.TypeAction
)

type OperationType string

const (
	// for sandbox
	OperationProvision   OperationType = "provision"
	OperationDeprovision OperationType = "deprovision"
	OperationReprovision OperationType = "reprovision"
	// for components
	OperationDeploy   OperationType = "deploy"
	OperationTeardown OperationType = "teardown"
	// for actions
	OperationTrigger OperationType = "trigger"
)

type ValidOperations []OperationType

func (v ValidOperations) String() string {
	operationStrings := make([]string, len(v))
	for i, op := range v {
		operationStrings[i] = string(op)
	}
	return strings.Join(operationStrings, ",")
}

var validOperations ValidOperations = []OperationType{
	OperationProvision,
	OperationDeprovision,
	OperationReprovision,
	OperationDeploy,
	OperationTeardown,
	OperationTrigger,
}

// OperationRolesConfig defines role assignments for operations at the app level
type OperationRolesConfig struct {
	Type       OperationRuleConfigType `mapstructure:"type" toml:"type" jsonschema:"required"` // Should be "matrix"
	RuleMatrix []*OperationRoleRule    `mapstructure:"rules,omitempty" toml:"rules,omitempty"`
}

func (c OperationRolesConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("configuration type").
		Long("Type of operation roles configuration. Must be 'matrix' for rule-based role assignment").
		Field("rules").Short("operation role rules").
		Long("Array of rules that map principals (components, sandboxes, actions) and operations to IAM roles")
}

func (c *OperationRolesConfig) Parse() error {
	return nil
}

func (c *OperationRolesConfig) Validate() error {
	if c == nil {
		return nil
	}

	if c.Type != "matrix" {
		return errors.New("operation roles supports only matrix type config")
	}

	// validate if a particular rule is duplicated for a principal and a operation
	seen := make(map[string]bool)
	for i, rule := range c.RuleMatrix {
		key := fmt.Sprintf("%s:%s", rule.Principal, rule.Operation)
		if seen[key] {
			return fmt.Errorf("duplicate rule at index %d: principal %q with operation %q is already defined", i, rule.Principal, rule.Operation)
		}
		seen[key] = true
	}

	for i, rule := range c.RuleMatrix {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("rule at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidateWithConfig validates operation role rules against actual components and actions, to make sure refs are correct
func (c *OperationRolesConfig) ValidateWithConfig(
	components []*Component,
	actions []*ActionConfig,
	permissions *PermissionsConfig,
	breakGlass *BreakGlass,
) error {
	if c == nil {
		return nil // Optional config
	}

	if err := c.Validate(); err != nil {
		return err
	}

	componentNames := make(map[string]bool)
	for _, comp := range components {
		componentNames[comp.Name] = true
	}

	actionNames := make(map[string]bool)
	for _, action := range actions {
		actionNames[action.Name] = true
	}

	availableRoles := make(map[string]bool)
	if permissions != nil {
		for _, role := range permissions.Roles {
			if role.Name != "" {
				roleName := strings.ReplaceAll(role.Name, " ", "")
				availableRoles[roleName] = true
			}
		}
	}
	if breakGlass != nil {
		for _, role := range breakGlass.Roles {
			if role.Name != "" {
				roleName := strings.ReplaceAll(role.Name, " ", "")
				availableRoles[roleName] = true
			}
		}
	}

	for i, rule := range c.RuleMatrix {
		principalType, principalName, err := rule.ParsePrincipal()
		if err != nil {
			return fmt.Errorf("rule at index %d: failed to parse principal: %w", i, err)
		}

		// skip checking refs for wildcard rules
		if principalName != "*" {
			switch PrincipalType(principalType) {
			case PrincipalTypeComponent:
				if principalName == "" {
					return fmt.Errorf("rule at index %d: component principal has empty name", i)
				}
				if !componentNames[principalName] {
					return fmt.Errorf("rule at index %d: component %q does not exist in app config", i, principalName)
				}

			case PrincipalTypeAction:
				if principalName == "" {
					return fmt.Errorf("rule at index %d: action principal has empty name", i)
				}
				if !actionNames[principalName] {
					return fmt.Errorf("rule at index %d: action %q does not exist in app config", i, principalName)
				}

			case PrincipalTypeSandbox:
				continue

			default:
				return fmt.Errorf("rule at index %d: unknown principal type %q", i, principalType)
			}
		}

		roleName := strings.ReplaceAll(rule.RoleName, " ", "")
		if !availableRoles[roleName] {
			return fmt.Errorf("rule at index %d: role %q does not exist in permissions or break_glass config", i, rule.RoleName)
		}
	}

	return nil
}

// OperationRoleRule maps a principal (component/sandbox/action) + operation to a role name
type OperationRoleRule struct {
	// Format: "nuon::component:name", "nuon::sandbox", "nuon::action:name"
	Principal string `mapstructure:"principal" toml:"principal" jsonschema:"required"`
	// "provision", "deprovision", "update", "reprovision", "trigger"
	Operation OperationType `mapstructure:"operation" toml:"operation" jsonschema:"required"`
	RoleName  string        `mapstructure:"role" toml:"role" jsonschema:"required"`
}

func (r OperationRoleRule) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("principal").Short("principal identifier").
		Long("Identifier for the entity: 'nuon::component:name', 'nuon::sandbox', or 'nuon::action:name'. Supports wildcards like 'nuon::component:*'").
		Examples("nuon::component:database", "nuon::sandbox", "nuon::action:healthcheck").
		Field("operation").Short("operation type").
		Long("Type of operation: provision, deprovision, update, reprovision, or trigger").
		Examples("deploy", "provision", "deprovision").
		Field("role").Short("IAM role name").
		Long("Name of the IAM role to use for this operation (not ARN). Role must exist in install stack outputs").
		Example("{{.nuon.install.id}}-maintenance")
}

func (r *OperationRoleRule) Validate() error {
	if err := r.ValidatePrincipal(); err != nil {
		return err
	}

	if !slices.Contains(validOperations, r.Operation) {
		return fmt.Errorf("operation must be one of: %s", validOperations.String())
	}

	if strings.TrimSpace(r.RoleName) == "" {
		return errors.New("role name cannot be empty")
	}

	return nil
}

func (r *OperationRoleRule) ValidatePrincipal() error {
	if r.Principal == "" {
		return errors.New("principal cannot be empty")
	}

	if !strings.HasPrefix(r.Principal, "nuon::") {
		return fmt.Errorf("principal must start with 'nuon::' (got: %s)", r.Principal)
	}

	principalType, principalName, err := r.ParsePrincipal()
	if err != nil {
		return err
	}

	validTypes := []PrincipalType{PrincipalTypeComponent, PrincipalTypeSandbox, PrincipalTypeAction}
	if !slices.Contains(validTypes, PrincipalType(principalType)) {
		validTypeStrings := make([]string, len(validTypes))
		for i, t := range validTypes {
			validTypeStrings[i] = string(t)
		}
		return fmt.Errorf("principal type must be one of: %s", strings.Join(validTypeStrings, ", "))
	}

	if PrincipalType(principalType) == PrincipalTypeSandbox && principalName != "" {
		return errors.New("sandbox principal should not have a name (use 'nuon::sandbox')")
	}

	if (PrincipalType(principalType) == PrincipalTypeComponent ||
		PrincipalType(principalType) == PrincipalTypeAction) &&
		principalName == "" {
		return fmt.Errorf("%s principal must have a name (e.g., 'nuon::%s:name' or 'nuon::%s:*')",
			principalType, principalType, principalType)
	}

	return nil
}

// ParsePrincipal extracts the principal type and name from the principal string
// Examples:
//   - "nuon::component:database" -> ("component", "database", nil)
//   - "nuon::sandbox" -> ("sandbox", "", nil)
//   - "nuon::action:*" -> ("action", "*", nil)
func (r *OperationRoleRule) ParsePrincipal() (string, string, error) {
	p, err := principal.ParsePrincipal(r.Principal)
	if err != nil {
		return "", "", fmt.Errorf("unable to parse principal: %s", err)
	}

	return string(p.Type), p.Name, nil
}

// EntityOperationRole is used within entities like component, sandbox to specify roles for specific operation
// todo(sk): see if this needs to be used for actions  as well or we can use just role, using this makes easier for
// future modifications, otoh using just role makes it a bit brittle
type EntityOperationRole struct {
	Operation OperationType `mapstructure:"operation" toml:"operation" jsonschema:"required"`
	RoleName  string        `mapstructure:"role" toml:"role" jsonschema:"required"`
}

func (e EntityOperationRole) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("operation").Short("operation type").
		Long("Type of operation: provision, deprovision, update, reprovision, or trigger").
		Examples("provision", "deploy", "deprovision").
		Field("role").Short("IAM role name").
		Long("Name of the IAM role to use for this operation (not ARN). Role must exist in install stack outputs").
		Examples("{{.nuon.install.id}}-maintenance", "{{.nuon.install.id}}-provision")
}

func (e *EntityOperationRole) Validate() error {
	if !slices.Contains(validOperations, e.Operation) {
		return fmt.Errorf("operation must be one of: %s", validOperations.String())
	}

	// Validate role name is not empty
	if strings.TrimSpace(e.RoleName) == "" {
		return errors.New("role name cannot be empty")
	}

	return nil
}

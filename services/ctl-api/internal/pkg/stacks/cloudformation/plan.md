# Plan: Auto-populate nested template parameters from app inputs

## Problem

When multiple additional nested CloudFormation stacks declare the same parameter (e.g., `Namespaces`), the current conflict detection in `getAdditionalNestedStacks` errors out because each stack tries to hoist that parameter to the parent stack independently. There is no way to share a single customer-provided value across multiple nested templates.

## Solution

Before hoisting a nested template parameter to the parent stack, check if a matching app input already exists. If it does, wire the nested template parameter directly to the input's parent-level CloudFormation parameter (`!Ref InstallXxx`) instead of hoisting a duplicate. This generalizes the existing reserved-parameter pattern to app inputs.

## How parameter naming works today

- App input `namespaces` → `computeCloudFormationStackParameterName("namespaces")` → `"InstallNamespaces"`
- This `InstallNamespaces` parameter is already added to the parent CF template by `getInstallInputGroupParameters` in `app_inputs.go`
- Nested template declares `Namespaces` as a `Parameter` with `Type: String`

The matching rule: `strcase.ToCamel(input.Name) == nestedParamName` (e.g., input `namespaces` → `Namespaces`).

## Files to modify

### 1. `nested_template_additional.go`

#### a. Add an input-matching step in `buildAdditionalNestedStack`

After extracting parameters and before the conflict check, build a lookup map from camel-cased input names to their CF parameter names, then match against nested template parameters.

Insert a new block between the first-class output wiring (lines 139-144) and the return (line 146):

```go
// Wire app inputs to matching nested template parameters.
// If a nested template parameter name matches an app input name (via ToCamel),
// reference the input's parent-level CF parameter instead of hoisting.
inputMatched := matchInputsToNestedParams(inp, defaultParameters)
for paramName, inputRef := range inputMatched {
    parameters[paramName] = inputRef
    delete(defaultParameters, paramName)
}
```

#### b. Add the `matchInputsToNestedParams` function

```go
func matchInputsToNestedParams(inp *stacks.TemplateInput, nestedParams map[string]cloudformation.Parameter) map[string]string {
    if inp.AppCfg.InputConfig == nil {
        return nil
    }

    // Build lookup: ToCamel(input.Name) → input.CloudFormationStackParamName
    inputLookup := map[string]string{}
    for _, input := range inp.AppCfg.InputConfig.AppInputs {
        if input.Source != app.AppInputSourceCustomer {
            continue
        }
        inputLookup[strcase.ToCamel(input.Name)] = input.CloudFormationStackParamName
    }

    matched := map[string]string{}
    for paramName := range nestedParams {
        if cfnParamName, ok := inputLookup[paramName]; ok {
            matched[paramName] = cloudformation.Ref(cfnParamName)
        }
    }
    return matched
}
```

#### c. Relax the conflict check

The conflict check on lines 92-97 must skip parameters that were matched to inputs (since they are no longer in `defaultParameters` after the `delete` call above). No change is needed here because the `delete(defaultParameters, paramName)` in step (a) already removes matched params before the conflict loop runs.

Verify: the conflict loop iterates `defaultParameters` returned from `buildAdditionalNestedStack`, which now excludes input-matched params. Two stacks declaring `Namespaces` will both match the same input and neither will hoist, so no conflict.

### 2. `nested_template_additional_test.go`

Add the following tests:

#### a. `TestGetAdditionalNestedStacks_InputMatchedParam`

One nested stack declares `Namespaces` parameter. An app input `namespaces` exists. Verify:
- `Namespaces` is NOT in `result.params` (not hoisted)
- The nested stack's `Parameters["Namespaces"]` equals `cloudformation.Ref("InstallNamespaces")`
- The param group for this stack does NOT include `Namespaces`

#### b. `TestGetAdditionalNestedStacks_InputMatchedAcrossMultipleStacks`

Two nested stacks both declare `Namespaces`. An app input `namespaces` exists. Verify:
- No conflict error
- Both stacks' `Parameters["Namespaces"]` equal `cloudformation.Ref("InstallNamespaces")`
- `Namespaces` is NOT in `result.params`

#### c. `TestGetAdditionalNestedStacks_InputMatchDoesNotAffectUnmatchedParams`

A nested stack declares `Namespaces` (matched by input) and `CustomParam` (not matched). Verify:
- `Namespaces` is wired via input ref
- `CustomParam` is still hoisted normally in `result.params`

#### d. `TestGetAdditionalNestedStacks_NoInputConfig`

`inp.AppCfg.InputConfig` is nil. A nested stack declares `Namespaces`. Verify:
- No panic
- `Namespaces` is hoisted normally (no input matching)

#### Test helper changes

Update `newTestInput` to accept an optional `*app.AppInputConfig` or add a variant that includes inputs. The input config needs:
```go
InputConfig: &app.AppInputConfig{
    AppInputs: []app.AppInput{
        {
            Name:                        "namespaces",
            Source:                       app.AppInputSourceCustomer,
            CloudFormationStackParamName: "InstallNamespaces",
        },
    },
},
```

Note: `CloudFormationStackParamName` is normally computed in `AfterQuery` via gorm hook. In tests, set it directly since there's no DB.

### 3. No changes needed to these files

- `nested_template_vpc.go` — `extractNestedStackParameters` stays as-is. Input matching is applied after extraction, in `buildAdditionalNestedStack`.
- `app_inputs.go` — Input parameters are already added to the parent template. No changes.
- `aws_eks_template.go` — Inputs are already wired to parent params on lines 82-85. The `Ref` from nested stacks will resolve to those params. No changes.

## Execution order in the composed parent template

```
Parent CF Template
├── Parameters:
│   ├── InstallNamespaces (from app_inputs.go, source: customer)
│   ├── InstallClusterName (from app_inputs.go)
│   └── CustomParam (hoisted from nested template, no input match)
│
├── Resources:
│   ├── VPC (nested stack)
│   ├── RunnerAutoScalingGroup (nested stack)
│   ├── EksAccessEntries (additional nested stack)
│   │   └── Parameters:
│   │       ├── ClusterName: "literal-value" (reserved)
│   │       ├── NuonInstallID: "literal-value" (reserved)
│   │       └── Namespaces: !Ref InstallNamespaces (input-matched)
│   ├── RunnerSgEksAccess (additional nested stack)
│   │   └── Parameters:
│   │       ├── ClusterName: "literal-value" (reserved)
│   │       └── NuonInstallID: "literal-value" (reserved)
│   └── PhoneHomeProps, RunnerPhoneHome, etc.
```

## Priority order for parameter resolution

When `buildAdditionalNestedStack` processes a nested template parameter, the following precedence applies (first match wins):

1. **Reserved params** — `ClusterName`, `NuonInstallID`, `NuonAppID`, `NuonOrgID` → literal values injected by nuon
2. **First-class outputs** — parameter name matches an output from VPC or RunnerASG templates → `!GetAtt` wiring
3. **Input-matched params** — parameter name matches `ToCamel(input.Name)` for a customer-sourced input → `!Ref InstallXxx`
4. **Hoisted params** — everything else → hoisted to parent stack as a customer-facing CF parameter

This is already the natural order in `buildAdditionalNestedStack`: reserved params are handled first (lines 123-137), then first-class outputs (lines 139-144), then the new input-matching step, with anything remaining in `defaultParameters` being hoisted.

## Import to add

In `nested_template_additional.go`, add:
```go
"github.com/nuonco/nuon/services/ctl-api/internal/app"
```

The `strcase` import is already present.

## Verify

```sh
cd services/ctl-api/internal/pkg/stacks/cloudformation
go test -v -run TestGetAdditionalNestedStacks
```

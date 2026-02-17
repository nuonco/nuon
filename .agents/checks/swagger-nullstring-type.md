---
name: Swagger NullString Type Tag
description: Ensures generics.NullString fields have swaggertype:"string" to prevent SDK deserialization failures
severity-default: critical
tools: [Grep, Read]
globs: ["services/ctl-api/internal/app/*.go"]
---

When reviewing GORM model structs in `services/ctl-api/internal/app/`, check that every field using `generics.NullString` has either `swaggertype:"string"` or `swaggerignore:"true"` in its struct tag.

## Why This Is Critical

`generics.NullString` has a custom `MarshalJSON` that serializes as a **plain JSON string** (e.g., `"some-id"`). However, go-swagger sees the underlying struct fields (`String`, `Valid`) and generates a `GenericsNullString` object type in SDK models. This causes a fatal deserialization error on the runner side:

```
json: cannot unmarshal string into Go struct field ... of type models.GenericsNullString
```

Adding `swaggertype:"string"` tells swag to emit `"type": "string"` in the OpenAPI spec, so the generated SDK models use `string` instead of `GenericsNullString`.

## Check

For every field of type `generics.NullString`:

- ✅ Has `swaggertype:"string"` — correctly treated as string in SDK
- ✅ Has `swaggerignore:"true"` — excluded from swagger spec entirely
- ❌ Has neither — **will break SDK deserialization at runtime**

## Example

```go
// ❌ BAD — SDK will generate GenericsNullString object, causing unmarshal errors
InstallActionWorkflowID generics.NullString `json:"install_action_workflow_id,omitzero"`

// ✅ GOOD — SDK will generate string type
InstallActionWorkflowID generics.NullString `json:"install_action_workflow_id,omitzero" swaggertype:"string"`

// ✅ ALSO GOOD — field excluded from swagger entirely
OrgID generics.NullString `json:"org_id,omitzero" swaggerignore:"true"`
```

## How to Find Violations

```bash
grep -n 'generics\.NullString' services/ctl-api/internal/app/*.go | grep -v 'swaggertype' | grep -v 'swaggerignore'
```

Any results from the above command are violations that must be fixed before merging.

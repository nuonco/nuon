---
name: write-rego-policy
description: Write or update an OPA/Rego policy for a Nuon app's `policies/` directory (guardrails on terraform, helm, pulumi, kubernetes, or sandbox components). Use when asked to add, update, or explain a policy that blocks or warns on a component's deploy plan or manifest.
domain: apps
---

# Writing Rego policies

A Nuon policy is two files, same directory, same base name:

- `<name>.toml` — metadata: what it applies to and how it's evaluated
- `<name>.rego` — the actual check, evaluated by OPA

```toml
type       = "terraform_module"
engine     = "opa"
contents   = "./<name>.rego"
components = ["*"]
```

This skill covers `engine = "opa"` only — that's the only engine to use for Rego policies.

## Choosing `type`

`type` determines what `input` your policy sees. Pick the one matching the component(s) you're guarding:

| `type` | Component | `input` shape |
|---|---|---|
| `terraform_module` | Terraform module components | Terraform JSON plan: `input.plan.resource_changes[]` |
| `sandbox` | Sandbox infra (terraform-backed) | Terraform JSON plan: `input.plan.resource_changes[]` |
| `sandbox` | Sandbox infra (pulumi-backed) | Pulumi preview: `input.plan.change_summary`, resource diffs |
| `pulumi` | Pulumi components | Pulumi preview: `input.plan.change_summary`, resource diffs |
| `helm_chart` | Helm chart components | Rendered manifest / Kubernetes object |
| `kubernetes_manifest` | Raw k8s manifest components | Kubernetes AdmissionReview: `input.review.object`, `input.review.kind` |
| `kubernetes_cluster` | Cluster-level resources (namespaces, CRDs) | Kubernetes AdmissionReview: `input.review.object`, `input.review.kind` |
| `container_image` | Container image components | Image reference/metadata |
| `docker_build` | Docker build components | Build metadata |

When unsure which `input` shape a type produces, look for an existing policy of the same `type` in the app's `policies/` directory and match its input access pattern.

## Package and rule conventions

Every policy file:

```rego
package nuon

import future.keywords.contains
import future.keywords.if
import future.keywords.in
```

Two rule kinds, both accumulate `msg` strings:

- `deny contains msg if { ... }` — hard fail, blocks the operation
- `warn contains msg if { ... }` — advisory, surfaced but doesn't block

**Start new checks as `warn`, with a `TODO: promote to deny` comment**, unless the user explicitly asks for a hard block. This is the pattern every existing policy in this codebase follows — it lets a new check surface violations across existing installs before it starts blocking them.

Every message should name the offending resource (`rc.address`, `input.review.object.metadata.name`, etc.) via `sprintf`, so the violation is actionable from the dashboard/report without re-reading the plan.

## Worked examples

**Terraform — parsing an embedded JSON policy document, flagging overly broad IAM grants:**

```rego
package nuon

import future.keywords.contains
import future.keywords.if
import future.keywords.in

iam_policy_resources := {"aws_iam_policy", "aws_iam_role_policy", "aws_iam_user_policy"}
admin_action_patterns := {"*", "*:*", "iam:*", "s3:*", "ec2:*"}

parse_policy(s) := json.unmarshal(s)

statements(doc) := s if {
	is_array(doc.Statement)
	s := doc.Statement
} else := [doc.Statement] if {
	is_object(doc.Statement)
}

to_set(x) := {x} if { is_string(x) }
to_set(x) := {v | some v in x} if { is_array(x) }

warn contains msg if {
	some rc in input.plan.resource_changes
	rc.type in iam_policy_resources
	rc.change.actions[_] in ["create", "update"]
	some stmt in statements(parse_policy(rc.change.after.policy))
	stmt.Effect == "Allow"
	some action in to_set(stmt.Action)
	action in admin_action_patterns
	msg := sprintf("IAM policy '%s' grants wide action '%s'. Scope it down.", [rc.address, action])
}
```

Resource changes live at `input.plan.resource_changes[]`, each with `.type`, `.address`, and `.change.actions` / `.change.after`. Reference: `iam-least-privilege.rego`, `block-destructive-changes.rego`, `network-no-public-ingress.rego` in an app's `policies/` directory for more terraform-shape checks (security groups, ingress rules, destroy detection).

**Helm — dispatching on Kubernetes kind to reach the pod spec, then checking `securityContext`:**

```rego
package nuon

import future.keywords.contains
import future.keywords.if
import future.keywords.in

pod_spec_resources := {"Pod", "Deployment", "StatefulSet", "DaemonSet", "Job", "CronJob", "ReplicaSet"}

get_pod_spec(obj) := obj.spec if { input.review.kind.kind == "Pod" }
get_pod_spec(obj) := obj.spec.template.spec if { input.review.kind.kind in {"Deployment", "StatefulSet", "DaemonSet", "ReplicaSet", "Job"} }
get_pod_spec(obj) := obj.spec.jobTemplate.spec.template.spec if { input.review.kind.kind == "CronJob" }

deny contains msg if {
	input.review.kind.kind in pod_spec_resources
	pod_spec := get_pod_spec(input.review.object)
	some container in pod_spec.containers
	container.securityContext.privileged == true
	msg := sprintf("%s '%s' container '%s' runs privileged. Not allowed.", [input.review.kind.kind, input.review.object.metadata.name, container.name])
}
```

Reference: `pod-security.rego` (this example, plus non-root and host-namespace checks), `no-cluster-roles.rego`, `no-loadbalancer-services.rego`.

## Testing

Write a sibling `<name>_test.rego` (same `package nuon`) with mock `input` fixtures and `test_*` rules asserting `count(deny) > 0` / `count(warn) == 0` etc. — see `iam-least-privilege_test.rego`, `sg-protocol-restriction_test.rego` for the pattern. Run `opa test policies/` before treating the policy as done.

## Wiring `components`

`components = ["*"]` applies the policy to every component of the matching `type`. To scope it, list specific component names instead: `components = ["rds_cluster", "eks_cluster"]`.

## `contents`: keep it a local file

`contents` technically accepts more than a relative path — it supports Nuon templating and external sources (an `https://` URL, a git repo reference). Don't reach for those for a new policy: every policy in practice is `contents = "./<name>.rego"`, a plain relative path next to the `.toml`. Only use templating/external sources if the user explicitly asks for one.

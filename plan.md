# Overview

This RFC proposes a unified, declarative policy configuration model for Nuon applications, along with an execution strategy for evaluating OPA-based policies across Kubernetes manifests, Terraform modules, and Helm charts. The goal is to make policy authoring, distribution, evaluation, and auditability consistent and predictable across all types of components, while keeping the system flexible and scalable.

The core idea:

**A single policies.toml defines all policies (1 → N mapping), and workflows—not runners—own plan evaluation and policy enforcement.**

# Problem Statement

Describe the problem that needs to be solved.

### **Policy Definition**

Policies need to be tied to specific components. This makes global, cross-component rules easy to enforce.

Example problem:

- A global “no public ingress” rule must be applied across every Kubernetes component and be forwards compatible to new components that can be added.

### **Directory-Based Grouping**

Nuon heavily relies on directory structure for semantic grouping (e.g., components, actions)

Policies should follow a similar pattern.

## **Goals**

1. **Simple, centralized policy definition** via a single `policies.toml`.
2. **Support for inline rego based policies**.
3. **Declarative mapping of policies to components/types** (1 policy → N components).
4. **Enforce policies on real plan outputs**, including:
    - Terraform plan JSON
    - Kube manifests after dry-run (server defaults applied)
    - Helm chart templates (full template output)
5. **Policy evaluation happens inside workflows**, not runners.
6. **Store policy reports in DB** for auditability and approvals.

# Solution

```jsx
[[policy]]
type = "kubernetes_manifest"
engine = "opa"
contents = "./no-update-cluster-policy.rego"
components = ["*"]

[[policy]]
type = "terraform_module"
engine = "opa"
contents = "./rds-encrypted-policy.rego"
components = ["rds_cluster"]

[[policy]]
type = "terraform_module"
engine = "opa"
contents = "./tags-missing.rego"
sandbox = true
```

Key characteristics:

- A policy maps to **multiple components** instead of embedding policies inside each component.
- `contents` can point to a file or be in-lined.
- Policies can be restricted to sandbox environments (`sandbox_only`).

## Policies are **resource constraints**, not **process steps**.

Consider a vendor who wants to prevent insecure RDS configurations, such as:

- RDS instances missing encryption,
- Publicly exposed RDS instances,
- RDS instances missing mandatory tags,
- RDS instances being modified after initial creation.

These constraints are properties of the **RDS resource itself**, not properties of a workflow like “sandbox provision” or “sandbox update.”

The same RDS misconfiguration is unsafe regardless of the workflow step that produced it. A vendor’s intent is not:

> “Enforce this rule only when the user is running the sandbox_provision workflow.”
> 

The vendor’s actual intent is:

> “This rule must apply to any operation that creates, updates, or deletes RDS resources.”
> 

### Why we shouldn’t have process level policies.

If policy binding is done using workflow-based types (e.g., `policy_sandbox_provision`, `policy_sandbox_update`, `policy_deploy`, etc.), then the vendor must manually attach the same rule to **every workflow path** that might affect the resource. This creates an immediate correctness hazard:

- Missing one workflow path means the policy silently does not run.
- Adding a new workflow path requires updating every existing policy binding.
- Refactoring workflows breaks policy coverage even though the resource invariants remain unchanged.

This design forces policy authors to reason about the *internal state machine of the platform* rather than the *security posture of the resource*. As workflows evolve over time, policy bindings must be updated continuously, even though the security requirement (“RDS must be encrypted”) has not changed. This is brittle & error-prone.

In contrast, binding policies to **resource roles**—such as `terraform_module`, `kubernetes_manifest`, or `helm_chart`—aligns the enforcement model with how vendors actually think about risk:

- “For all Terraform modules that produce RDS resources, the following invariants must hold.”
- “For all Kubernetes manifests, Ingress objects must never be public.”
- “For all Helm charts, PVC names must not be mutated.”

These constraints are properties of the resource domain, not of the execution flow.

With a role-based model:

- A policy attached to an RDS module applies regardless of whether the system is provisioning, updating, reconciling, or destroying.
- The policy remains correct even as resources are added, removed, or modified.
- Vendors do not need to understand or track the internal workflow architecture of the platform.
- Policy coverage is complete by construction, because policies follow the resource, not the process.

**Policies attach to resource types and resource semantics, not to lifecycle events.**

## Implementation

For correct evaluation of policies, runner job output needs to contain:

| Type | Required Input | Why |
| --- | --- | --- |
| Kubernetes Manifest | **Dry-run output** (defaults applied) | Policies expect real server-side manifests, not raw YAML. |
| Helm Chart | **helm template output** | Policies expect full K8s objects. |
| Terraform Module | **Terraform plan JSON** | Required for resource-level decisions. |

We’ve the necessary data for terraform. However, we should have also have the Runner job output include details for Kubernetes Manifest and Helm. For MVP, we could consider doing it on a non-dry run output for Kubernetes Manifest. 

### **Opa Integration**

OPA allows for policies to be evaluated in this manner using  "[github.com/open-policy-agent/opa/rego](http://github.com/open-policy-agent/opa/rego)" Go module-

```
r := rego.New(
    rego.Query("data.example.allow"), // your decision
    rego.Module("policy.rego", policy),
    rego.Input(input),
)

rs, err := r.Eval(ctx)
if err != nil {
    return nil, err
}
if len(rs) == 0 {
    return nil, fmt.Errorf("undefined decision")
}

return rs[0].Expressions[0].Value, nil
}
```

**Enforcement Architecture**

### **Where Enforcement Runs**

**Enforcement runs in workflows / ctl-api**, not in the runner.

Reasons:

- Workflows already process plans (Terraform plans, manifests, diffs).
- DB is the correct long-term storage for:
    - Full helm template output
    - K8s dry-run manifest
    - Terraform plan JSON
- Avoid round-trips to runner jobs.
- Avoid complications of:
    - Relying on runner availability
    - Log shipping
    - Build pipelines
- Temporal workflows already handle:
    - No-op approvals
    
    ### Workflow Integration
    
    New Child Workflow: `CheckPolicies` 
    
    Input:
    
    - plan object (terraform or helm or kube)
    - policy set filtered by:
        - type
        - component match
        - sandbox_only
    
    Steps:
    
    1. Load policies from DB.
    2. Evaluate each policy against its plan.
    3. Store the violations as a workflow property
    
    ## Go-getter embedding in component config
    
    - Embed plain Rego text into DB.
    - Simple to implement.
    - No build pipeline required.
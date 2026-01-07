# Disallow Reading Secrets

A Kubewarden policy that prevents the creation of `Role` and `ClusterRole` resources that grant permissions to read `Secrets`.

## Description

This policy inspects `Role` and `ClusterRole` resources and violates if any rule allows `get`, `list`, or `watch` verbs on `secrets` resources.

## How to use

You can run this policy using `kwctl`:

```bash
kwctl run -e gatekeeper annotated-policy.wasm -r test_data/role-reading-secret.json
```

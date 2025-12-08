# Disallow Creating Roles and ClusterRoles

This policy prevents the creation of `Role` and `ClusterRole` resources in the cluster.

## Usage

This policy is intended to be used with [Kubewarden](https://kubewarden.io).

## Testing

To run the tests:

```bash
make test
make e2e-tests
```

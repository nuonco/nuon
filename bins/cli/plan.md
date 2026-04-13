# CLI Config Loading Improvements

We have a few issues and inconsistencies with our config loading as related to the loading of permissions:

- permissions/

We intend to support two paths:

Scenario 1

```txt
permissions.toml
permissions/*.json <- individual policy files referenced from permissions.toml: contents = "./permissions/policy.json"
```

Scenario 2

```txt
permissions/
 - maintenance.toml <- the actual role
 - *.json           <- the policy json files referenced from within maintenance.toml: contents = "./policy.json"
```

`permissions.toml` and `permissions/*.toml` are mutually exclusive.

if a user tries to use `permissions.toml` and `permissions/*.toml` we should raise an error.

When searching for permissions, we _want_ to be able to reference files using their relative path if they are referenced
from within a `permissions/*.toml` file. If `permissions/*.json` files are being referenced from `permissions.toml` at
the root of the app config:

1. if a polcy file referenced does not have the `permissions/` directory in its path:

- look for it in the root
- if not present, look for it in `permissions/`
  - if found in permissions, raise a warning telling the user to use the full path.

2. if the file does have `permissions/` in its path load it as desired (happy path).

The end state is one in which users are able to use permissions.toml files but are encouraged to use the permissions/
directory to store the policy documents themselves.

## Task

We need to ensure we support both scenarios. Read the `nuon apps sync` (`SyncDir`) code and `pkg/config`. Ensure we have
explicit handling for both cases with an early exit in case `permissions.toml` and `permissions/*.toml` are both in use.

## Development Notes

- There is a running `nctl`/`nuonctl` process that's watching and building `nuon-dev` which we can use for testing.
- Use `NUON_DEBUG=true` for additional logging

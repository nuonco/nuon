Returns a list of available IAM roles that can be used for a specific operation on an install.

The endpoint filters roles based on the operation type:
- **provision/reprovision**: Custom roles, break glass roles, provision IAM role
- **deprovision/teardown**: Custom roles, break glass roles, deprovision IAM role
- **deploy**: Custom roles, break glass roles, maintenance IAM role
- **trigger** (actions): Custom roles, break glass roles, provision + maintenance IAM roles

Roles are sourced from the install's stack outputs.

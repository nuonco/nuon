# sandboxes

Sandboxes are terraform templates that we use to provision installations.

## Local

You can provision a terraform workspace for each sandbox, just like you normally would, by setting values etc. However, in most cases running `earthly +lint` and then creating a real installation in stage is the best approach to validate your changes.

To execute `earthly` targets:

```bash
$ earthly +lint --SANDBOX=aws-eks
```

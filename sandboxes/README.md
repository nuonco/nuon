# sandboxes

Sandboxes are terraform templates that we use to provision installations. They define the environment in the customer's cloud that a vendor's software will be deployed to.

Currently we only use the aws-eks sandbox, and have the empty sandbox for testing. Note that whatever outputs you add to the aws-eks sandbox, will also need to be added to the empty sandbox even though there's nothing in it. We just need them for testing purposes.

## Local

You can provision a terraform workspace for each sandbox, just like you normally would, by setting values etc. However, in most cases running `earthly +lint` and then creating a real installation in stage is the best approach to validate your changes.

To execute `earthly` targets:

```bash
$ earthly +lint --SANDBOX=aws-eks
```


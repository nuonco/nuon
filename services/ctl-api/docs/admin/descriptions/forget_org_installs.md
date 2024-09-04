Forget all installs for an org.

This should only be used in _dire_ cases where an org has been lost, and we need to deprovision it but do not know / care about the end installs. The primary use case for this is to be able to run aws-nuke on the canary account when things go wrong, but still allow us to cleanup the apps/orgs after.

Recover a Helm release that was left part-way through an operation.

Helm records a `pending-install`, `pending-upgrade` or `pending-rollback` status before it
starts changing the cluster and clears it once the operation finishes. A release left in one
of those statuses is a rollout whose runner went away — a crash, a cancelled workflow, or a
job that timed out. Helm then refuses every further operation on that release, and retrying
the deploy cannot clear it.

This endpoint starts a workflow that returns the release to a usable state:

- when an earlier revision finished a rollout, the release is rolled back to it
- when no revision ever rolled out, the stuck release is removed

It deploys nothing and changes no desired state. Deploy the component afterwards to roll out
the version you want.

The recovery refuses to act on a release that is not pending, so it is safe to run when you
are unsure and it is a no-op on a second run.

Returns `409` when a job is already running for the component (recovering while Helm is
genuinely mid-operation can corrupt the release) or when the component has never been
deployed on this install. Returns `400` when the component is not a Helm chart.

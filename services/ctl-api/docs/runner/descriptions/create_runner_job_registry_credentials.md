Mint a short-lived credential for the app's repository in Nuon's management registry, for use by one runner job.

Only the job's own token (claimed via `ClaimRunnerJobToken`) may call this, and only while the job is available or in progress. The credential is limited to the repository the job's plan names: pull only, or pull and push when the plan writes to that repository. The credential is returned once and is never stored in the plan.

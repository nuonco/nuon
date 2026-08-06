# Airgap follow-ups (post M1–M3)

Status: M1–M3 shipped in PR #2097 (2026-08-10). This tracks what was
deferred or discovered and not yet done.

## Deferred by the M1–M3 plan

- [ ] Secrets/inputs resolution: bundles assume all inputs are resolvable
      at export time; no customer-side secret injection yet.
- [ ] Runner heartbeats (liveness signal into the state bucket).
- [ ] CloudFormation WaitCondition + instance-replacement resume demo.

## Hardening (from the pre-PR review pass)

- [ ] `nuon-bundle push` drops the underlying error when marking an
      artifact failed.
- [ ] Plan export ignores GetCompositePlan/DeriveCompositePlan errors and
      silently drops the affected runner job from the envelope.
- [ ] Runner S3 fetch retries permanent errors (bad creds, access denied)
      for the full 30-minute window.
- [ ] Stack-output values that fail JSON marshaling are dropped without
      surfacing an error.
- [ ] Runner binary fetch permits plain HTTP; config contract documents
      HTTPS/file only.
- [ ] DB-level uniqueness for concurrent bundle creation (insert race).

## Unverified

- [ ] Provider-mirror E2E: enable `terraform-provider-mirror` org flag,
      rebuild, verify runner logs use the mirror.

## Housekeeping

- [ ] `acme` demo stacks left running in sandbox-ht (intentional; tear down
      when no longer needed).
- [ ] KMS key 3ff079ac-a4e9-44af-9eaa-679b9258f0c3 self-deletes 2026-08-14.

## Next: day-2 ops (design in progress)

Cron actions, healthchecks, runbooks, drift checks, and a customer-side
portal to observe and trigger them — same zero-ingress/egress constraint,
S3 as the only dispatch/observation channel, plans come from the bundle.

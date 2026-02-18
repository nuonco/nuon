# Runner Failover & Scaling

This doc covers scenarios where runners fail, are replaced, or need to scale horizontally within a RunnerGroup.

See [concepts.md](concepts.md) for the election algorithm and [upgrades.md](upgrades.md) for
zero-downtime upgrade flows.

## ASG Scale-Out (Multiple Cloud Instances)

**Status: Not supported.** The current system does not handle scaling an ASG to 2+ instances within the same RunnerGroup.

### Problem

The `runner.nuon.co/id` EC2 tag is set at the launch template / ASG level during provisioning, so all instances in
the ASG share the **same Runner ID**. When a second VM boots:

1. It calls `RunnerAuthAWS` with the same `runner.nuon.co/id` tag value.
2. `getRunnerWithGroup()` finds the existing Runner record — there is no "create if not found" path.
3. Both VMs receive tokens for the same Runner identity.
4. Both poll the same event loop (keyed by Runner ID), causing duplicate execution or lost jobs.

### Why `AdminCreateRunnerInGroup` Doesn't Help

The admin endpoint uses find-or-create semantics **keyed by platform**. Two cloud VMs of the same platform
(e.g. `aws-eks`) would match the same existing runner — it would not create a second record.

### Possible Approaches

**Option A: Instance-aware registration in `RunnerAuthAWS`**

Change the auth flow so that when a new EC2 `instanceID` authenticates against a RunnerGroup, the API creates a new
Runner record automatically. The EC2 tag would identify the *group* (or a template runner), not the individual runner.

- `RunnerAuthAWS` sees unknown instanceID → creates Runner in the group → issues token for the new record.
- Subsequent calls with the same instanceID return the existing Runner.
- ASG scale-in / instance termination triggers runner cleanup + leader re-election.

**Option B: Runner-side self-registration on boot**

The runner binary calls an `AdminCreateRunnerInGroup`-style endpoint on startup, keyed by instance ID instead of
platform, so each VM gets its own Runner record.

- Requires changing the registration key from platform to instance identity.
- Runner binary needs group ID (from EC2 tags or config) instead of a pre-provisioned runner ID.

Either approach plugs into the existing leader election — multiple active runners in the group, `ElectLeader` picks
the oldest untainted runner, and jobs route to the leader.

### Impact on Rolling Upgrades

Once ASG scale-out is supported, the rolling upgrade workflow works naturally: each VM has its own Runner record with
independent taint/version/status tracking. The workflow upgrades them one at a time using the same
taint → drain → upgrade → untaint sequence.

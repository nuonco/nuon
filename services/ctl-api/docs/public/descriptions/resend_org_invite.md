Resend the invite email for an existing pending org invite.

The invite must be in "pending" status. Accepted and revoked invites cannot be resent.
Resend preserves the stored role and scope, revalidates delegation against the resender's current grants and credential
limits, and replaces the private source authorization proof. It cannot retarget or broaden an invitation. The invite
and its delivery signal are persisted atomically; delivery is notified after commit.

Invites without explicit migrated scope fail closed. An invite with valid scope but absent or expired source proof
can receive fresh proof through an authorized resend. Revocation or expiry of that source credential before acceptance
requires another authorized resend; it does not remove membership already granted by successful acceptance.

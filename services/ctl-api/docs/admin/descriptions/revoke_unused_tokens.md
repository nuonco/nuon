Revokes every token of the runner's service account that no live runner process is using.

Each heartbeat records the token the process sent it with. The signal fails without revoking anything until every live
process has recorded its token. A pending process that has recorded a token keeps it; one that has not does not block
the signal. The newest token and any token created in the last 12 hours are always kept.

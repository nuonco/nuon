Permanently delete the logs and trace spans of a finished action run, for example after a command printed a secret.

Only org admins can delete logs, and only after the run has ended, including runs that failed, timed out, or were
cancelled. The run itself and its outputs are kept. The
log stream is marked with `logs_deleted_at` and `logs_deleted_by_id` from the first deletion. Calling this again
deletes any records that arrived late and keeps the original marker.

Once the logs are deleted, the log stream's write token is revoked, new writes to the stream are dropped, and reads of
the stream return no logs or spans. Trace spans are removed in the background after the call returns.

Rotate any secret that was exposed: deleting the logs does not undo the exposure, and copies held in the ingestion
queue expire on their own retention schedule. If that queue is replayed before then, call this endpoint again.

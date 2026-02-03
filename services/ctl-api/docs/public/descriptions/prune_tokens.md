Prune old tokens for a specific runner by invalidating all tokens except the most recent one.

This is useful for cleaning up old tokens without disrupting the currently running runner.
The latest token (by creation time) is preserved, ensuring the active runner continues to function.

Returns the count of invalidated tokens for the specified runner.

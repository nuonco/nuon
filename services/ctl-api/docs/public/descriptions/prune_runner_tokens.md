Prune old tokens for an install's runner by invalidating all tokens except the most recent one.

This is useful for cleaning up old tokens without disrupting the currently running install runner.
The latest token (by creation time) is preserved, ensuring the active runner continues to function.

Returns the count of invalidated tokens.

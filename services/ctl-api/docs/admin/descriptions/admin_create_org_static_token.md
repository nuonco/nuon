Create a static token, with access to an org. Can be used in CI with Github Actions or via the CLI, api etc.

The duration field allows you to set an expiry, which by default is 365 days. Since we do not currently support invalidating tokens, the best way to invalidate a token is by recreating it, with a duration of `0s`. Since tokens are per org, and do not support multiple, this will basically replace the previous version of a token.

**NOTE** you can only have a single static token at a time per org.

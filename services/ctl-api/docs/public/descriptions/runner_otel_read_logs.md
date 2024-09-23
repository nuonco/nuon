Runner log retrieval endpoint.

Supports pagination via a header:`X-Nuon-API-Offset`. This header can be passed
back to the api and controls the timestamp from which the pagination on the
request.

The endpoint returns the offset for the next page in the header:
`X-Nuon-API-Next`. To use the next page, use that value in the
`X-Nuon-API-Offset` header.

The implicit offset in a request that provides to `X-Nuon-API-Offset` is 0. This
returns the first page.

This endpoint accepts two query params: `job_id` and `job_execution_id`. Neither
is required.

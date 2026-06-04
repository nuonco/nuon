Update input values for an install.

This endpoint accepts a partial subset of inputs and merges them with the install's existing
inputs, so callers only need to send the inputs they want to change. Inputs sourced from the
`install_stack` (customer source) are managed by the install stack and are rejected if supplied.

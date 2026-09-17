package api

// MCPPoliciesInstructions is included in MCP server instructions so clients
// run the local policies CLI extension when editing app IAM, instead of
// inventing overlap and boundary checks.
const MCPPoliciesInstructions = "When authoring or changing IAM under permissions/ or break_glass/, use the local nuon policies CLI extension. If `nuon policies --help` fails, run `nuon ext install nuonco/nuon-ext-policies`. From the app directory, run `nuon policies --output json check-overlap <role>.toml` for each role file (for example maintenance.toml) and `nuon policies --output json check-boundaries`. Fix overlapping actions and high-severity boundary drift before `nuon apps sync`."

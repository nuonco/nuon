# Contract: create an install

`POST /v1/apps/{app_id}/installs`

Language-neutral request/response contract. Combine with
`install-inputs-schema.md` (drives `inputs`) and `architecture-and-security.md`
(whitelist + injection rules).

## Request body

```jsonc
{
  "name": "acme-prod",              // REQUIRED. string. install display name.
  "inputs": {                        // map<string,string>. keys MUST come from
    "region": "us-west-2",           // the app input schema. values are strings.
    "instance_size": "large"
  },

  // Exactly one cloud account block, matching the app's cloud_platform.
  // SERVER-OWNED — never accept from the customer.
  "aws_account":   { /* ...aws account fields... */ },
  "azure_account": { /* ...azure account fields... */ },
  "gcp_account":   { /* ...gcp account fields... */ },

  // Optional, SERVER-OWNED:
  "labels":         { "tenant_id": "cust_123" },  // key/value, merged into install
  "metadata":       { /* HelpersInstallMetadata */ },
  "install_config": { /* HelpersCreateInstallConfigParams — pin a config version */ }
}
```

Field ownership:

| Field                                  | Set by  | Notes                                        |
| -------------------------------------- | ------- | -------------------------------------------- |
| `name`                                 | customer| required; validate non-empty                 |
| `inputs`                               | customer| whitelisted to schema keys only              |
| `aws_account` / `azure_account` / `gcp_account` | server | one block, chosen by app cloud platform |
| `labels`                               | server  | attach tenant identity here                  |
| `metadata`, `install_config`           | server  | optional                                     |

## Response (201/200)

Returns the created install object. Key fields to return to the client:

```jsonc
{
  "id": "install_...",
  "name": "acme-prod",
  "app_id": "app_...",
  "cloud_platform": "aws",
  "sandbox_status": "...",
  "runner_status": "...",
  "created_at": "..."
}
```

Return `id` + status fields to the UI so it can transition to a status view and
poll (`lifecycle.md`). Do NOT echo back server-owned secrets or the raw cloud
account block.

## Proxy handler outline (pseudocode, any language)

```
POST /api/installs (body: { name, inputs }):
  customer = authenticate(request)                 # vendor's own session
  authorizeCreate(customer, APP_ID)                # tenant → app mapping (TODO)

  schema = getInputSchema(APP_ID)                  # cache OK
  errors = validateAgainstSchema(body.inputs, schema)   # required, types, unknown keys
  if body.name is empty or errors: return 422

  ctlBody = {
    name:   body.name,
    inputs: pick(body.inputs, schema.inputNames),  # whitelist
    <cloudPlatform>_account: SERVER_CLOUD_ACCOUNT, # injected
    labels: { tenant_id: customer.tenantId },      # injected
  }

  resp = nuon.post(`/v1/apps/${APP_ID}/installs`, ctlBody)  # Bearer + X-Nuon-Org-ID
  if !resp.ok: log(resp); return mapError(resp.status)

  return 201 { id, name, sandbox_status, runner_status, created_at }
```

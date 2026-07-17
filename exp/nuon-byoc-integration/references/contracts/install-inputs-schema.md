# Contract: app install-input schema

`GET /v1/apps/{app_id}/input-latest-config`

Returns the app's declared install inputs. Use it to (a) render the create form
dynamically and (b) build the server-side whitelist/validation. Never hardcode
the field list — it drifts from the app config.

## Response shape (relevant fields)

```jsonc
{
  "id": "input_config_...",
  "app_id": "app_...",
  "input_groups": [ { "id": "...", "name": "...", "display_name": "..." } ],
  "inputs": [
    {
      "name": "region",              // key used in the install `inputs` map
      "display_name": "AWS Region",  // label for the form
      "description": "Region to deploy into",
      "type": "string",             // string | number | bool | yaml | hcl | json ...
      "required": true,
      "default": "us-west-2",
      "sensitive": false,            // if true, mask in the UI, don't log
      "internal": true,              // if true, NOT customer-facing — exclude
      "source": "customer",          // "customer" | "vendor" (defaults to "vendor")
      "group_id": "...",             // optional grouping
      "index": 0                     // display order
    }
  ]
}
```

## How each field drives generation

| Field          | Form behavior                          | Server behavior                          |
| -------------- | -------------------------------------- | ---------------------------------------- |
| `name`         | field key; sent in `inputs` map        | whitelist key; reject anything not listed|
| `display_name` | field label (fallback to `name`)       | —                                        |
| `description`  | helper text                            | —                                        |
| `type`         | input control (text/number/checkbox)   | coerce/validate value type               |
| `required`     | required field, block submit if empty  | reject if missing                        |
| `default`      | prefill                                | apply if omitted                         |
| `sensitive`    | masked input, never logged             | never log value                          |
| `internal`     | **exclude from the form entirely**     | reject if customer supplies it           |
| `source`       | render ONLY `source == "customer"`     | whitelist = customer inputs only; reject vendor inputs from the browser |
| `index`        | sort order                             | —                                        |
| `group_id`     | group fields under `input_groups`      | —                                        |

### `source` is the primary customer/vendor split

Every input has `source`: `"customer"` (the vendor chose to expose it to their
customers) or `"vendor"` (the default — the vendor sets it, or it falls back to
the app-config default). **The customer-facing UI and the server whitelist must
include only `source == "customer"` inputs.** Vendor inputs are never rendered to
or accepted from the customer; they resolve from the app config server-side. Do
not rely on the name (e.g. the `nuon_component_override_v1_` prefix) to decide —
use `source`. Combine with `internal` (still always excluded):

```
customerFacing = inputs.filter(i => i.source === "customer" && !i.internal)
```

An app may expose zero customer inputs — then the form shows only the install
name, which is correct.

Values are transmitted to ctl-api as **strings** in the `inputs` map — coerce
UI values (numbers, booleans) to their string form when building the request.

## Retrieval options

- **At runtime (recommended):** the proxy's `GET /api/install-inputs` endpoint
  fetches this live (with the server-side token) and returns only the
  customer-facing subset (`source == "customer"` and not `internal`).
- **At build time (exploration):** `nuon --output json apps input-config
  --app-id <app_id>` shows the input definitions (incl. `source`) for confirming
  shapes while generating code.

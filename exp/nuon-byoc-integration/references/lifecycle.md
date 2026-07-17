# Install lifecycle (language-neutral)

Creating an install is asynchronous. `POST /v1/apps/{app_id}/installs` returns a
`201`/`200` with an install object almost immediately, but the install is not yet
provisioned — a sandbox is created, then components deploy. The UI must reflect
this.

## Status fields on the install object

The install object (`GET /v1/installs/{install_id}`) exposes several status
axes — surface the ones relevant to the flow:

| Field                                  | Meaning                                    |
| -------------------------------------- | ------------------------------------------ |
| `sandbox_status`                       | provisioning state of the install sandbox  |
| `runner_status`                        | health of the install's runner             |
| `composite_component_status`           | rolled-up status across components          |
| `component_statuses`                   | per-component map                           |
| `install_states`                       | historical state transitions               |

`sandbox_status` / `runner_status` are the most useful for a "is my install
ready?" indicator right after creation.

## Recommended UI flow after create

1. `POST /api/installs` → returns `{ id, name, sandbox_status, ... }`.
2. Navigate to a status view keyed by install `id`.
3. Poll `GET /api/installs/:id` on an interval (e.g. every 5–10s) until the
   sandbox/runner reach a terminal healthy state, then stop polling.
4. Show clear intermediate ("Provisioning…"), success, and error states.

Do not block the create request waiting for provisioning to finish — it can take
minutes.

## Beyond create (future contract files)

The same proxy pattern extends to:

- **Deploy** components into an install.
- **Update inputs** (`POST /v1/installs/{install_id}/inputs`).
- **Teardown / delete** the install.

Each becomes its own file under `contracts/` and its own narrow, authorized proxy
endpoint. Keep the whitelist + tenant-authorization + error-mapping invariants
identical for every one.

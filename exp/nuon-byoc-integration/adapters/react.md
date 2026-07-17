# Adapter: React (frontend)

Reference implementation of the customer-facing UI. Idioms only — the flow lives
in `references/lifecycle.md`. Copyable snippets are in `assets/client/`.

## API layer

See `assets/client/installsApi.js`. Typed calls to the vendor proxy (NOT
ctl-api):

- `getInstallInputs()` → `GET /api/install-inputs`
- `createInstall({ name, inputs })` → `POST /api/installs`
- `getInstall(id)` → `GET /api/installs/:id`

If the repo uses React Query / SWR, wrap these as hooks
(`useInstallInputs`, `useCreateInstall`, `useInstall(id)` with polling).
Otherwise use `useEffect` + local state.

## Dynamic form

See `assets/client/CreateInstallForm.jsx`. Renders fields from the fetched input
schema:

- one control per input, chosen by `type` (text / number / checkbox),
- `display_name` as label, `description` as helper text,
- `required` enforced client-side, `default` prefilled,
- `sensitive` → password-style masked input,
- inputs sorted by `index`, optionally grouped by `input_groups`.

Submit builds `{ name, inputs }` (coerce values to strings) and calls
`createInstall`. Disable the button while pending. On success, navigate to the
status view with the returned install `id`.

## Status view (async)

See `assets/client/InstallStatus.jsx`. Polls `getInstall(id)` every ~5s, shows
`sandbox_status` / `runner_status`, stops polling on a terminal state, renders
provisioning / ready / error states.

## Notes

- Match the repo's component conventions (design system, form primitives, router).
  Reuse existing inputs/buttons rather than dropping in bespoke markup.
- The client never sees Nuon credentials or the ctl-api URL.

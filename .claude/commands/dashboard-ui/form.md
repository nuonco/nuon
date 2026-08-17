---
name: dashboard-ui:form
description: Use when building a form inside a modal in the dashboard-ui
---

Forms use **TanStack Form + Zod** behind the field-aware `Form*` wrappers in
`client/components/common/form/`. The form owns field state + validation; the container owns the
async mutation. Never hand-roll `useState`-per-field, `FormData`, or `forwardRef`/`requestSubmit`.

## When is it a form? (falsifiable rule)

**Always use TanStack Form + Zod for any surface that collects user input.** Skip it ONLY if:

- **E1 — no inputs:** buttons + static text only → plain confirm modal.
- **E2 — confirmation gate:** the only input is compared to a literal and NOT sent in the request
  body (type-to-confirm delete) → plain confirm modal.

There is no "it's an editor" exception. Bespoke direct-manipulation editors exist only by UXDR
sign-off (allowlist: `DeploymentPlanEditor`). Anything else with inputs → TanStack.

## Recipe

1. **Co-located, hand-written Zod schema** in `schema.ts` next to the component. One schema per
   resource; mode-specific fields conditional via `.refine`/`.superRefine`. Do NOT derive from
   OpenAPI types. **Use FLAT fields — never nested-object fields** (`channelId`/`channelName`, not
   `channel: {id,name}`); nested objects break `canSubmit` validation in TanStack Form v1.

2. **`useForm`** with `defaultValues`, `validators: { onMount: schema, onChange: schema }`
   (onMount gives an accurate `canSubmit` from the start; errors still only display after touch),
   and `onSubmit: ({ value }) => onSubmit(value)` calling the container's callback.

3. **Wrap the fields in a real `<form>` element** — it IS a form, so use the element. This turns off
   browser autocomplete across every field and gives correct semantics:
   ```tsx
   <form autoComplete="off" noValidate onSubmit={(e) => e.preventDefault()}
     className="flex flex-col gap-6">
     {/* FormErrorBanner + fields */}
   </form>
   ```
   **Submission is still driven by the modal's `primaryActionTrigger`, not native submit** — the
   trigger is a footer prop rendered *outside* the form, so `preventDefault` the form's `onSubmit`
   and drive it from the button. Read reactive state with `useStore` (the trigger is a prop, not a
   child), no `forwardRef`, no `requestSubmit`:
   ```tsx
   const form = useForm({ defaultValues, validators: { onMount: schema, onChange: schema },
     onSubmit: ({ value }) => onSubmit(value) })
   const canSubmit = useStore(form.store, (s) => s.canSubmit)
   // primaryActionTrigger:
   //   disabled: !canSubmit || isPending
   //   children: isPending ? '<gerund>…' : '<Verb> <thing>'
   //   onClick: () => form.handleSubmit()
   ```
   `isPending` comes from the container's mutation (async lives there), `canSubmit` from the form
   (validity). Pass `disabled={isPending}` to every field so inputs lock while submitting.

4. **Every field is `<FormX field={field} … />`** — one shape, no exceptions:
   ```tsx
   <form.Field name="name">
     {(field) => <FormInput field={field} labelProps={{ labelText: 'Name' }} />}
   </form.Field>
   ```
   Wrappers: `FormInput`, `FormTextarea`, `FormSelect`, `FormCheckbox`, `FormToggle`,
   `FormRadioGroup` (options array → one field), `FormCodeInput`, plus composite wrappers
   (`FormMatchPicker`, `FormInterestsPicker`, `FormChannelSelect`, …) co-located with their widget.
   Wrappers drive `error`/`errorMessage` from Zod — do NOT pass native `required`.

5. **Two-value widgets** (e.g. ChannelSelect emits id + name) bind via a primary `field` + an
   `onName` companion that calls `form.setFieldValue('otherField', name)`.

6. **Errors → in-form banner, never a toast.** Use `FormErrorBanner`:
   `<FormErrorBanner error={error} fallback="Unable to create webhook" />`. The Zod error message
   must differ from a field's `helperText` or they collide.

7. **Success → close + toast** (contract): container `onSuccess` calls `removeModal(props.modalId)`
   then `addToast`. Toast heading is a plain string, past tense ("Webhook created"); entity names in
   the `<Text>` body (see `COPY_STYLE.md`). Navigate only for a canonical destination; edits never
   navigate.

8. **Edit reuses create** (contract §5) **when create and edit hit the SAME endpoint** (webhooks,
   channel subs): one component + schema per resource with a `mode: 'create' | 'edit'` prop; edit
   prefills; mode-specific fields conditional; never a forked `EditXModal`. **When create and edit hit
   DIFFERENT endpoints** (branches, OIDC policies, the install editors), keep them as **separate
   in-place forms** — do NOT force-collapse them into one `mode` component. For a genuine edit-only
   form, gate the submit button on a "no change" check too (e.g. `name === currentName`) alongside
   `canSubmit`.

9. **Buttons:** create = "Create {thing}" + PlusIcon; edit = "Save changes", no icon. Submitting =
   gerund + Loading icon.

10. **Ladle behavior test required** — a sibling `e2e/specs-ladle/<form>.spec.ts` driving the story
    (no backend): assert disabled-until-valid and error-on-touch. See
    `e2e/specs-ladle/create-api-token.spec.ts` as the template. Every wrapper + form also gets a
    `.stories.tsx` (spin up a tiny `useForm` to bind the field).

11. Always spread `{...props}` onto `<Modal>` (see `dashboard-ui:user-flow`).

## Anti-Patterns

- **No** `useState`-per-field, `new FormData(...)`, or `forwardRef` + `requestSubmit` — TanStack Form
  owns state; the modal trigger calls `form.handleSubmit()`. **DO** wrap fields in a
  `<form autoComplete="off" noValidate onSubmit={(e) => e.preventDefault()}>` element (it's a form —
  the element turns off autocomplete); just don't wire the mutation to native submit.
- **No** nested-object fields — flatten them (`canSubmit` breaks on nested objects in v1). Array
  fields (`mode="array"` of `{...}` rows) are fine — that's different from a single field whose value
  is an object.
- **No** cross-row / array-level rules inside the Zod schema — TanStack v1 will NOT clear a
  form-level standard-schema error mapped to an ARRAY field even after it's satisfied (the button
  stays disabled forever). Put "at least one row must …" checks as a **live computed gate** on the
  submit button instead: `const rows = useStore(form.store, s => s.values.items); disabled: !canSubmit
  || !hasValidRow(rows)`. Per-row field rules (`z.string().min(1)` on `items[i].key`) are fine.
- **No** native `required` on fields — Zod is the sole validation source (avoids double-validation).
- **No** error toasts — errors are an in-form `FormErrorBanner`.
- **No** OpenAPI-derived schemas — hand-write Zod (wire shape ≠ UX validation).

## Advanced patterns (only when needed)

- **Programmatic set + dirty-tracking** (a picker autofills fields while preserving user edits): do
  NOT use TanStack's `field.meta.isDirty` — a programmatic `form.setFieldValue` also flips it. Keep a
  local `isXDirty` flag, set it only from the field's own `onChange` (bind a raw `Input`/`Select` to
  `form.Field` so you can add the side-effect), and skip the autofill when dirty. See
  `CreateOIDCTrustPolicy`.
- **Container-owned ephemeral state** (repo browser, role selector) stays a prop passed into the form,
  not a form field — the form owns only the committed values; cross-checks against it (e.g. "a repo
  must be selected") run imperatively inside `form.onSubmit`. See branches / run-action forms.
- **Multi-page wizard**: keep the page/step state local; gate each step's Next on `canSubmit` (or the
  relevant subset) instead of the old `reportValidity()` dance. See `RunRunbookForm`.

## Canonical sources

- Simple: `client/components/api-tokens/CreateApiToken/`
- Custom validation: `client/components/webhooks/CreateWebhook/`
- Composite widgets: `client/components/slack/CreateChannelSubscription/`
- Dynamic array + dirty-tracking + array-gate: `client/components/oidc-trust-policies/CreateOIDCTrustPolicy/`
- Multi-page wizard: `client/components/runbooks/RunRunbook/RunRunbookForm.tsx`
- Plan (completed): `services/dashboard-ui/.planning/ux/011-plan-tanstack-form-migration.md`

---
name: dashboard-ui:form
description: Use when building a form inside a modal in the dashboard-ui
---

This skill enforces useMutation for form submission, FormData extraction from the submit event, requestSubmit for HTML5 validation, and Toast/Banner for feedback.

## Steps

1. Create a `formRef` and wire it to the `<form>` element:
   ```typescript
   const formRef = useRef<HTMLFormElement>(null)
   ```

2. Use `useMutation` for the submission — never `useState(isLoading)` + raw `fetch()`:
   ```typescript
   const { mutate, isPending, error } = useMutation({
     mutationFn: (body: TMyBody) => myApiCall({ body, orgId: org.id }),
     onSuccess: () => {
       addToast(<Toast heading="Webhook created" theme="success"><Text>Created {name} for this org.</Text></Toast>)
       removeModal(props.modalId)
     },
   })
   ```
   Toast headings are plain strings (no JSX) per `services/dashboard-ui/COPY_STYLE.md` — past tense naming the thing that happened ("Webhook created"), never generic "Success"; entity names go in the `<Text>` description.

3. In the modal's `primaryActionTrigger.onClick`, call `formRef.current?.requestSubmit()` — NOT `mutate(...)` directly. This fires the form's `onSubmit` and triggers HTML5 `required` field validation before the mutation runs.

4. In the form's `onSubmit` handler, extract fields via `FormData` — not from React state:
   ```typescript
   const handleFormSubmit = (e: FormEvent<HTMLFormElement>) => {
     e.preventDefault()
     const formData = new FormData(e.currentTarget)
     mutate({ fieldName: formData.get('fieldName') as string })
   }
   ```

5. Show inline mutation errors with `Banner` inside the form:
   `{error && <Banner theme="error">{error?.error}</Banner>}`

6. Show success feedback via `addToast` from `useToast()`, then close the modal via `removeModal(props.modalId)`.

7. Always spread `{...props}` onto `<Modal>` — see `dashboard-ui:user-flow` for why this is required.

## Anti-Patterns

- **Do not** call `mutate(body)` directly from `primaryActionTrigger.onClick` — call `formRef.current?.requestSubmit()` instead so `required` fields are validated before submission
- **Do not** manage form fields with individual `useState` variables — use `new FormData(e.currentTarget)` in the submit handler
- **Do not** use `useEffect` to react to form field changes — form side-effects belong in the submit handler
- **Do not** show mutation errors via `addToast` — use `Banner theme="error"` inline so users can correct the form

Canonical source: `client/components/installs/management/RunAdhocAction.tsx`

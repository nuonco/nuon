---
name: dashboard-ui:user-flow
description: Use when implementing a button → modal → confirmation action in the dashboard-ui
---

This skill enforces the two-component (Button + Modal) pattern using useSurfaces, Modal from surfaces/, and useMutation — never raw fetch or bespoke modal divs.

## Steps

1. Create two components in the same file: `MyActionModal` and `MyActionButton`. Place the file in the relevant domain directory (e.g., `client/components/installs/management/MyAction.tsx`). If the component needs a container/component split, follow the pattern in `dashboard-ui:component`.

2. In the Modal component, use `({ item, ...props }: IMyAction & IModal)` — keep `...props` as a rest; never destructure `modalId`, `isVisible`, or `onClose` from it.

3. Use `useMutation` for the async action — never call API functions inside a click handler directly. **Always invalidate related queries in `onSuccess`** so lists/views that display the affected resource refresh immediately:
   ```typescript
   const queryClient = useQueryClient()
   const { mutate: execute, isPending } = useMutation({
     mutationFn: () => myApiCall({ itemId: item.id, orgId: org.id }),
     onSuccess: () => {
       queryClient.invalidateQueries({ queryKey: ['relevant-list', org.id] })
       addToast(<Toast heading="Workflow canceled" theme="success"><Text>Canceled the workflow for {item.name}.</Text></Toast>)
       removeModal(props.modalId)
     },
     onError: (err: TAPIError) => {
       addToast(<Toast heading="Workflow cancellation failed" theme="error"><Text>{err?.error || 'An unknown error occurred.'}</Text></Toast>)
     },
   })
   ```
   Find the right `queryKey` by checking the `useQuery` call in the container of the component that displays the data you're mutating.

   Toast headings are plain strings (no JSX/Badge) following `services/dashboard-ui/COPY_STYLE.md`: past tense for instant completions ("Plan approved"), present progressive for kicked-off jobs ("Deploying component"), "[thing] failed" for errors. Entity names and context go in the `<Text>` description.

4. Pass the action via `primaryActionTrigger` prop on `<Modal>`, not as a child button. Per `COPY_STYLE.md`: confirmation headings are a question ("Cancel workflow?"), button labels are verb + object (never "Confirm"/"OK"), gerund while pending, `variant: 'danger'` for destructive actions:
   ```typescript
   <Modal
     heading="Cancel workflow?"
     primaryActionTrigger={{ children: isPending ? 'Canceling workflow' : 'Cancel workflow', disabled: isPending, onClick: () => execute(), variant: 'danger' }}
     {...props}
   >
   ```

5. Spread `{...props}` onto `<Modal>` — required for `SurfacesProvider` to inject `modalId`. Without it, `removeModal(props.modalId)` fails silently and the modal never closes.

6. In the Button component, use `useSurfaces().addModal` and create the modal instance before passing:
   ```typescript
   const { addModal } = useSurfaces()
   const modal = <MyActionModal item={item} />
   return <Button onClick={() => addModal(modal)} {...props}>Action label</Button>
   ```

7. For panel flows (sliding side panel): follow the same pattern with `addPanel`/`removePanel` and `Panel` from `client/components/surfaces/Panel`.

Canonical source: `client/components/approvals/ApprovePlan.tsx`

## Anti-Patterns

- **Do not** call an API function inside a click handler — use `useMutation`
- **Do not** build a modal with `<div className="fixed inset-0 ...">` — use `Modal` from `client/components/surfaces/Modal`
- **Do not** destructure `modalId` from props — always spread `{...props}` onto `<Modal>`, or `removeModal(props.modalId)` will fail
- **Do not** use `useState(isOpen)` to control modal visibility — use `useSurfaces().addModal` / `removeModal`
